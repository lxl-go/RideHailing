package service

import (
	"context"

	"ride-hailing/services/gateway-service/internal/biz"
	"ride-hailing/services/gateway-service/internal/data"
	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

type TripService struct {
	uc *biz.TripUsecase
}

func NewTripService(uc *biz.TripUsecase) *TripService {
	return &TripService{uc: uc}
}

func (s *TripService) SearchTrips(ctx context.Context, req data.SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	return s.uc.SearchTrips(ctx, req)
}

func (s *TripService) GetTripDetail(ctx context.Context, id int64) (*tripv1.GetTripDetailReply, error) {
	return s.uc.GetTripDetail(ctx, id)
}

func (s *TripService) PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	return s.uc.PublishTrip(ctx, req)
}

func (s *TripService) ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	return s.uc.ListDriverTrips(ctx, req)
}

func (s *TripService) UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) error {
	return s.uc.UpdateTripStatus(ctx, req)
}

func (s *TripService) ListCoupons(ctx context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error) {
	return s.uc.ListCoupons(ctx, req)
}

func (s *TripService) ClaimCoupon(ctx context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error) {
	return s.uc.ClaimCoupon(ctx, req)
}

func (s *TripService) PublishDemand(ctx context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error) {
	return s.uc.PublishDemand(ctx, req)
}

func (s *TripService) ListMyDemands(ctx context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error) {
	return s.uc.ListMyDemands(ctx, req)
}

func (s *TripService) CancelDemand(ctx context.Context, req *tripv1.CancelDemandRequest) error {
	return s.uc.CancelDemand(ctx, req)
}

func (s *TripService) DeleteTrip(ctx context.Context, req *tripv1.DeleteTripRequest) error {
	return s.uc.DeleteTrip(ctx, req)
}

func (s *TripService) ValidateLocation(ctx context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error) {
	return s.uc.ValidateLocation(ctx, req)
}

func (s *TripService) PreviewTripPrice(ctx context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error) {
	return s.uc.PreviewTripPrice(ctx, req)
}

func (s *TripService) SuggestLocations(ctx context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error) {
	return s.uc.SuggestLocations(ctx, req)
}
