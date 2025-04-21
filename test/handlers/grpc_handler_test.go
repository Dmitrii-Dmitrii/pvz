package handlers

import (
	"context"
	"errors"
	"github.com/Dmitrii-Dmitrii/pvz/api"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/pvz_model"
	pvz_v1 "github.com/Dmitrii-Dmitrii/pvz/proto/generated/pvz/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestGetPVZList(t *testing.T) {
	ctx := context.Background()

	t.Run("Get PVZ List", func(t *testing.T) {
		mockService := new(MockPvzService)
		handler := api.NewGrpcHandler(mockService)

		now := time.Now()
		id1 := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		id2 := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockPvzList := []pvz_model.Pvz{
			{
				Id:               id1,
				RegistrationDate: now.AddDate(0, -1, 0),
				City:             pvz_model.City("Москва"),
			},
			{
				Id:               id2,
				RegistrationDate: now.AddDate(0, -2, 0),
				City:             pvz_model.City("Санкт-Петербург"),
			},
		}

		mockService.On("GetAllPvz", ctx).Return(mockPvzList, nil)

		expectedResponse := &pvz_v1.GetPVZListResponse{
			Pvzs: []*pvz_v1.PVZ{
				{
					Id:               id1.String(),
					RegistrationDate: timestamppb.New(now.AddDate(0, -1, 0)),
					City:             "Москва",
				},
				{
					Id:               id2.String(),
					RegistrationDate: timestamppb.New(now.AddDate(0, -2, 0)),
					City:             "Санкт-Петербург",
				},
			},
		}

		response, err := handler.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

		assert.NoError(t, err)
		assert.Equal(t, len(expectedResponse.Pvzs), len(response.Pvzs))

		for i, pvz := range response.Pvzs {
			assert.Equal(t, expectedResponse.Pvzs[i].Id, pvz.Id)
			assert.Equal(t, expectedResponse.Pvzs[i].City, pvz.City)

			assert.Equal(t, expectedResponse.Pvzs[i].RegistrationDate.AsTime().Unix(), pvz.RegistrationDate.AsTime().Unix())
		}

		mockService.AssertExpectations(t)
	})

	t.Run("Get PVZ List with empty result", func(t *testing.T) {
		mockService := new(MockPvzService)
		handler := api.NewGrpcHandler(mockService)

		var emptyPvzList []pvz_model.Pvz
		mockService.On("GetAllPvz", ctx).Return(emptyPvzList, nil)

		response, err := handler.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Pvzs)
		mockService.AssertExpectations(t)
	})

	t.Run("Get PVZ List with error", func(t *testing.T) {
		mockService := new(MockPvzService)
		handler := api.NewGrpcHandler(mockService)

		expectedError := errors.New("database connection error")
		mockService.On("GetAllPvz", ctx).Return([]pvz_model.Pvz{}, expectedError)

		response, err := handler.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Pvzs)
		mockService.AssertExpectations(t)
	})
}
