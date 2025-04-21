package pvz_driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/generated"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/product_model"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/pvz_model"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/reception_model"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"time"
)

type PvzDriver struct {
	adapter drivers.Adapter
}

func NewPvzDriver(adapter drivers.Adapter) *PvzDriver {
	return &PvzDriver{adapter: adapter}
}

func (d *PvzDriver) CreatePvz(ctx context.Context, pvz *pvz_model.Pvz) error {
	_, err := d.adapter.Exec(ctx, drivers.QueryCreatePvz, pvz.Id, pvz.RegistrationDate, pvz.City)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreatePvz.Message)
		return custom_errors.ErrCreatePvz
	}

	return nil
}

func (d *PvzDriver) GetPvzFullInfo(ctx context.Context, limit, offset uint32, startInterval, endInterval *time.Time) ([]map[string]interface{}, error) {
	query, params := getQueryGetPvz(limit, offset, startInterval, endInterval)

	rows, err := d.adapter.Query(ctx, query, params...)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetPvz.Message)
		return nil, custom_errors.ErrGetPvz
	}
	defer rows.Close()

	pvzMap, err := scanRowsToGetPvz(rows)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, pvz)
	}

	return result, nil
}

func (d *PvzDriver) GetPvzById(ctx context.Context, id pgtype.UUID) (*pvz_model.Pvz, error) {
	var registrationDate time.Time
	var city pvz_model.City

	err := d.adapter.QueryRow(ctx, drivers.QueryGetPvzById, id).Scan(&registrationDate, &city)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, custom_errors.ErrPvzNotFound
		}

		log.Error().Err(err).Msg(custom_errors.ErrGetPvz.Message)
		return nil, custom_errors.ErrGetPvz
	}

	pvz := &pvz_model.Pvz{Id: id, RegistrationDate: registrationDate, City: city}
	return pvz, nil
}

func (d *PvzDriver) GetAllPvz(ctx context.Context) ([]pvz_model.Pvz, error) {
	rows, err := d.adapter.Query(ctx, drivers.QueryGetAllPvz)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetPvz.Message)
		return nil, custom_errors.ErrGetPvz
	}
	defer rows.Close()

	var pvzList []pvz_model.Pvz
	for rows.Next() {
		var id pgtype.UUID
		var registrationDate time.Time
		var city pvz_model.City

		err = rows.Scan(&id, &registrationDate, &city)
		if err != nil {
			log.Error().Err(err).Msg(custom_errors.ErrScanRow.Message)
			return nil, custom_errors.ErrScanRow
		}

		pvz := pvz_model.Pvz{
			Id:               id,
			RegistrationDate: registrationDate,
			City:             city,
		}

		pvzList = append(pvzList, pvz)
	}

	return pvzList, nil
}

func getQueryGetPvz(limit, offset uint32, startInterval, endInterval *time.Time) (string, []interface{}) {
	query := drivers.QueryGetPvz
	var params []interface{}
	paramCnt := 0

	if startInterval != nil || endInterval != nil {
		query += " WHERE "

		if startInterval != nil {
			paramCnt++
			query += fmt.Sprintf("r.reception_time >= $%d", paramCnt)
			params = append(params, *startInterval)
		}

		if startInterval != nil && endInterval != nil {
			query += " AND "
		}

		if endInterval != nil {
			paramCnt++
			query += fmt.Sprintf("r.reception_time <= $%d", paramCnt)
			params = append(params, *endInterval)
		}
	}

	paramCnt++
	query += fmt.Sprintf(" ORDER BY p.id, r.reception_time DESC LIMIT $%d", paramCnt)
	params = append(params, limit)
	paramCnt++
	query += fmt.Sprintf(" OFFSET $%d", paramCnt)
	params = append(params, offset)

	return query, params
}

func scanRowsToGetPvz(rows pgx.Rows) (map[pgtype.UUID]map[string]interface{}, error) {
	pvzMap := make(map[pgtype.UUID]map[string]interface{})

	for rows.Next() {
		var pvzId, receptionId, productId pgtype.UUID
		var registrationDate, receptionTime, addingTime *time.Time
		var pvzCity pvz_model.City
		var receptionStatus *reception_model.ReceptionStatus
		var productType *product_model.ProductType

		err := rows.Scan(
			&pvzId,
			&registrationDate,
			&pvzCity,
			&receptionId,
			&receptionTime,
			&receptionStatus,
			&productId,
			&addingTime,
			&productType,
		)
		if err != nil {
			log.Error().Err(err).Msg(custom_errors.ErrScanRow.Message)
			return nil, custom_errors.ErrScanRow
		}

		pvzIdDto, err := services.ConvertPgUuidToOpenAPI(pvzId)
		if err != nil {
			return nil, err
		}

		pvzObj, exists := pvzMap[pvzId]
		if !exists {
			pvzObj = map[string]interface{}{
				"pvz": generated.PVZ{
					Id:               &pvzIdDto,
					RegistrationDate: registrationDate,
					City:             generated.PVZCity(pvzCity),
				},
				"receptions": []map[string]interface{}{},
			}
			pvzMap[pvzId] = pvzObj
		}

		if !receptionId.Valid {
			continue
		}

		receptions := pvzObj["receptions"].([]map[string]interface{})
		var receptionObj map[string]interface{}

		receptionIdDto, err := services.ConvertPgUuidToOpenAPI(receptionId)
		if err != nil {
			return nil, err
		}

		receptionExists := false
		for i, reception := range receptions {

			if reception["reception"].(generated.Reception).Id.String() == receptionIdDto.String() {
				receptionExists = true
				receptionObj = receptions[i]
				break
			}
		}

		if !receptionExists {
			receptionObj = map[string]interface{}{
				"reception": generated.Reception{
					Id:       &receptionIdDto,
					PvzId:    pvzIdDto,
					DateTime: *registrationDate,
					Status:   generated.ReceptionStatus(*receptionStatus),
				},
				"products": []generated.Product{},
			}
			receptions = append(receptions, receptionObj)
			pvzObj["receptions"] = receptions
		}

		if !productId.Valid {
			continue
		}

		productIdDto, err := services.ConvertPgUuidToOpenAPI(productId)
		if err != nil {
			return nil, err
		}

		productDto := generated.Product{
			Id:          &productIdDto,
			ReceptionId: receptionIdDto,
			DateTime:    addingTime,
			Type:        generated.ProductType(*productType),
		}

		receptionObj["products"] = append(receptionObj["products"].([]generated.Product), productDto)
	}

	return pvzMap, nil
}
