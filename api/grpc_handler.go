package api

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/pvz_service"
	pvz_v1 "github.com/Dmitrii-Dmitrii/pvz/proto/generated/pvz/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcHandler struct {
	pvz_v1.UnimplementedPVZServiceServer
	pvzService pvz_service.IPvzService
}

func NewGrpcHandler(pvzService pvz_service.IPvzService) *GrpcHandler {
	return &GrpcHandler{pvzService: pvzService}
}

func (h *GrpcHandler) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	log.Info().Msg("GetPVZList started")

	pvzList, err := h.pvzService.GetAllPvz(ctx)
	if err != nil {
		return &pvz_v1.GetPVZListResponse{}, err
	}

	var pvzs []*pvz_v1.PVZ
	for _, pvz := range pvzList {
		pvzs = append(pvzs, &pvz_v1.PVZ{
			Id:               pvz.Id.String(),
			RegistrationDate: timestamppb.New(pvz.RegistrationDate),
			City:             string(pvz.City),
		})
	}

	log.Info().Msgf("GetPVZList result: %v", pvzs)

	return &pvz_v1.GetPVZListResponse{
		Pvzs: pvzs,
	}, nil
}
