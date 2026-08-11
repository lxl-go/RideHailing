package biz

import (
	"context"

	"ride-hailing/services/gateway-service/internal/data"
	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
)

type PassengerUsecase struct {
	client data.PassengerClient
}

func NewPassengerUsecase(client data.PassengerClient) *PassengerUsecase {
	return &PassengerUsecase{client: client}
}

func (uc *PassengerUsecase) EnsurePassenger(ctx context.Context, id int64, phone string) (*passengerv1.PassengerProfileReply, error) {
	return uc.client.EnsurePassenger(ctx, id, phone)
}

func (uc *PassengerUsecase) GetPassenger(ctx context.Context, id int64) (*passengerv1.PassengerProfileReply, error) {
	return uc.client.GetPassenger(ctx, id)
}

func (uc *PassengerUsecase) UpdatePassenger(ctx context.Context, req *passengerv1.UpdatePassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	return uc.client.UpdatePassenger(ctx, req)
}
