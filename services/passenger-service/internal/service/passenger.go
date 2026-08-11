package service

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
	"ride-hailing/services/passenger-service/internal/biz"
)

type PassengerService struct {
	passengerv1.UnimplementedPassengerServiceServer
	uc *biz.PassengerUsecase
}

func NewPassengerService(uc *biz.PassengerUsecase) *PassengerService {
	return &PassengerService{uc: uc}
}

func (s *PassengerService) EnsurePassenger(ctx context.Context, req *passengerv1.EnsurePassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	profile, err := s.uc.EnsurePassenger(ctx, req.Id, req.Phone)
	if err != nil {
		return nil, mapError(err)
	}
	return &passengerv1.PassengerProfileReply{Passenger: toProto(profile)}, nil
}

func (s *PassengerService) GetPassenger(ctx context.Context, req *passengerv1.GetPassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	profile, err := s.uc.GetPassenger(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return &passengerv1.PassengerProfileReply{Passenger: toProto(profile)}, nil
}

func (s *PassengerService) UpdatePassenger(ctx context.Context, req *passengerv1.UpdatePassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	profile, err := s.uc.UpdatePassenger(ctx, biz.UpdatePassengerCommand{
		ID:                req.Id,
		Nickname:          req.Nickname,
		Phone:             req.Phone,
		AvatarURL:         req.AvatarUrl,
		CommonAddress:     req.CommonAddress,
		PaymentPreference: req.PaymentPreference,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &passengerv1.PassengerProfileReply{Passenger: toProto(profile)}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidPassenger):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrPassengerNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrPassengerDuplicate):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func toProto(profile *biz.PassengerProfile) *passengerv1.PassengerProfile {
	if profile == nil {
		return nil
	}
	return &passengerv1.PassengerProfile{
		Id:                profile.ID,
		Nickname:          profile.Nickname,
		Phone:             profile.Phone,
		AvatarUrl:         profile.AvatarURL,
		CommonAddress:     profile.CommonAddress,
		PaymentPreference: profile.PaymentPreference,
		Status:            int32(profile.Status),
		CreatedAt:         formatTime(profile.CreatedAt),
		UpdatedAt:         formatTime(profile.UpdatedAt),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
