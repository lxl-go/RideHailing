package biz

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ride-hailing/services/gateway-service/internal/data"
	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

type TripUsecase struct {
	client data.TripClient
}

func NewTripUsecase(client data.TripClient) *TripUsecase {
	return &TripUsecase{client: client}
}

func (uc *TripUsecase) SearchTrips(ctx context.Context, req data.SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	return uc.client.SearchTrips(ctx, req)
}

func (uc *TripUsecase) GetTripDetail(ctx context.Context, id int64) (*tripv1.GetTripDetailReply, error) {
	return uc.client.GetTripDetail(ctx, id)
}

func (uc *TripUsecase) PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	return uc.client.PublishTrip(ctx, req)
}

func (uc *TripUsecase) ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	return uc.client.ListDriverTrips(ctx, req)
}

func (uc *TripUsecase) UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) error {
	return uc.client.UpdateTripStatus(ctx, req)
}

func (uc *TripUsecase) ListCoupons(ctx context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error) {
	return uc.client.ListCoupons(ctx, req)
}

func (uc *TripUsecase) ClaimCoupon(ctx context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error) {
	return uc.client.ClaimCoupon(ctx, req)
}

func (uc *TripUsecase) PublishDemand(ctx context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error) {
	return uc.client.PublishDemand(ctx, req)
}

func (uc *TripUsecase) ListMyDemands(ctx context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error) {
	return uc.client.ListMyDemands(ctx, req)
}

func (uc *TripUsecase) CancelDemand(ctx context.Context, req *tripv1.CancelDemandRequest) error {
	return uc.client.CancelDemand(ctx, req)
}

func (uc *TripUsecase) DeleteTrip(ctx context.Context, req *tripv1.DeleteTripRequest) error {
	return uc.client.DeleteTrip(ctx, req)
}

func (uc *TripUsecase) ValidateLocation(ctx context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error) {
	client, ok := uc.client.(interface {
		ValidateLocation(context.Context, *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error)
	})
	if !ok {
		return nil, status.Error(codes.Unimplemented, "地点校验服务未配置")
	}
	return client.ValidateLocation(ctx, req)
}

func (uc *TripUsecase) PreviewTripPrice(ctx context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error) {
	client, ok := uc.client.(interface {
		PreviewTripPrice(context.Context, *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error)
	})
	if !ok {
		return nil, status.Error(codes.Unimplemented, "路线报价服务未配置")
	}
	return client.PreviewTripPrice(ctx, req)
}

func (uc *TripUsecase) SuggestLocations(ctx context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error) {
	client, ok := uc.client.(interface {
		SuggestLocations(context.Context, *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error)
	})
	if !ok {
		return nil, status.Error(codes.Unimplemented, "地点联想服务未配置")
	}
	return client.SuggestLocations(ctx, req)
}
