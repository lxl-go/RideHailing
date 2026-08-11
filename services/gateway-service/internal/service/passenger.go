package service

import (
	"context"

	"ride-hailing/services/gateway-service/internal/biz"
	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
)

type PassengerService struct {
	uc *biz.PassengerUsecase
}

func NewPassengerService(uc *biz.PassengerUsecase) *PassengerService {
	return &PassengerService{uc: uc}
}

func (s *PassengerService) EnsurePassenger(ctx context.Context, id int64, phone string) (*passengerv1.PassengerProfileReply, error) {
	return s.uc.EnsurePassenger(ctx, id, phone)
}

func (s *PassengerService) GetPassenger(ctx context.Context, id int64) (*passengerv1.PassengerProfileReply, error) {
	return s.uc.GetPassenger(ctx, id)
}

func (s *PassengerService) UpdatePassenger(ctx context.Context, req *passengerv1.UpdatePassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	return s.uc.UpdatePassenger(ctx, req)
}
