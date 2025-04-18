package pvz_driver

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/product_model"
	"pvz/internal/models/pvz_model"
	"pvz/internal/models/reception_model"
	"pvz/internal/services"
	"time"
)

type PvzDriver struct {
	rwdb *pgxpool.Pool
}

func NewPvzDriver(rwdb *pgxpool.Pool) *PvzDriver {
	return &PvzDriver{rwdb: rwdb}
}

func (d *PvzDriver) CreatePvz(ctx context.Context, pvz *pvzs.Pvz) (*pvzs.Pvz, error) {
	_, err := d.rwdb.Exec(ctx, drivers.QueryCreatePvz, pvz.Id, pvz.RegistrationDate, pvz.City)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreatePvz.Message)
		return nil, custom_errors.ErrCreatePvz
	}
	return pvz, nil
}

func (d *PvzDriver) GetPvz(ctx context.Context, limit, offset uint32, startInterval, endInterval *time.Time) ([]map[string]interface{}, error) {
	query, params := getQueryGetPvz(limit, offset, startInterval, endInterval)

	rows, err := d.rwdb.Query(ctx, query, params...)
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
		var pvzCity pvzs.City
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
				"reception_driver": []map[string]interface{}{},
			}
			pvzMap[pvzId] = pvzObj
		}

		if !receptionId.Valid {
			continue
		}

		receptions := pvzObj["reception_driver"].([]map[string]interface{})
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
				"product_driver": []generated.Product{},
			}
			receptions = append(receptions, receptionObj)
			pvzObj["reception_driver"] = receptions
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

		receptionObj["product_driver"] = append(receptionObj["product_driver"].([]generated.Product), productDto)
	}

	return pvzMap, nil
}
