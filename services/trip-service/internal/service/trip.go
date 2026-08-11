package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
	"ride-hailing/services/trip-service/internal/biz"
	"ride-hailing/services/trip-service/internal/data"
)

type TripService struct {
	tripv1.UnimplementedTripServiceServer
	uc   *biz.TripUsecase
	amap *data.AMapClient
}

func NewTripService(uc *biz.TripUsecase, amap ...*data.AMapClient) *TripService {
	service := &TripService{uc: uc}
	if len(amap) > 0 {
		service.amap = amap[0]
	}
	return service
}

func (s *TripService) ValidateLocation(ctx context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error) {
	if s.amap == nil || req == nil || req.Location == nil {
		return nil, status.Error(codes.InvalidArgument, "location service is unavailable")
	}
	locations, err := s.amap.Search(ctx, req.Keyword, "")
	if err != nil || len(locations) == 0 {
		return nil, status.Error(codes.Unavailable, "route service unavailable, please retry")
	}
	input := req.Location
	for _, candidate := range locations {
		if input.PoiId != "" && input.PoiId == candidate.POIID {
			return locationToReply(candidate), nil
		}
		if input.Name != "" && input.Name == candidate.Name &&
			input.Longitude == candidate.Longitude && input.Latitude == candidate.Latitude {
			return locationToReply(candidate), nil
		}
	}
	return locationToReply(locations[0]), nil
}

func locationToReply(location data.Location) *tripv1.ValidateLocationReply {
	return &tripv1.ValidateLocationReply{Location: &tripv1.LocationInput{
		PoiId:            location.POIID,
		Name:             location.Name,
		FormattedAddress: location.FormattedAddress,
		Longitude:        location.Longitude,
		Latitude:         location.Latitude,
	}}
}

func locationDisplayName(location *tripv1.LocationInput) string {
	if location == nil {
		return ""
	}
	if name := strings.TrimSpace(location.Name); name != "" {
		return name
	}
	return strings.TrimSpace(location.FormattedAddress)
}

func (s *TripService) SuggestLocations(ctx context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error) {
	if s.amap == nil || req == nil || strings.TrimSpace(req.Keyword) == "" {
		return nil, status.Error(codes.InvalidArgument, "location service is unavailable")
	}
	locations, err := s.amap.Search(ctx, strings.TrimSpace(req.Keyword), "")
	if err != nil || len(locations) == 0 {
		return nil, status.Error(codes.Unavailable, "route service unavailable, please retry")
	}
	limit := int(req.Limit)
	if limit <= 0 || limit > len(locations) {
		limit = len(locations)
	}
	reply := &tripv1.SuggestLocationsReply{Locations: make([]*tripv1.LocationInput, 0, limit)}
	for _, candidate := range locations[:limit] {
		reply.Locations = append(reply.Locations, &tripv1.LocationInput{
			PoiId:            candidate.POIID,
			Name:             candidate.Name,
			FormattedAddress: candidate.FormattedAddress,
			Longitude:        candidate.Longitude,
			Latitude:         candidate.Latitude,
		})
	}
	return reply, nil
}

func (s *TripService) PreviewTripPrice(ctx context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error) {
	if s.amap == nil || req == nil || req.Origin == nil || req.Destination == nil {
		return nil, status.Error(codes.InvalidArgument, "origin and destination are required")
	}
	depart, err := parseTime(req.DepartTime)
	if err != nil || depart.Before(time.Now().Add(15*time.Minute)) {
		return nil, status.Error(codes.InvalidArgument, "departure must be at least 15 minutes from now")
	}
	origin := data.Location{POIID: req.Origin.PoiId, Name: req.Origin.Name, FormattedAddress: req.Origin.FormattedAddress, Longitude: req.Origin.Longitude, Latitude: req.Origin.Latitude}
	destination := data.Location{POIID: req.Destination.PoiId, Name: req.Destination.Name, FormattedAddress: req.Destination.FormattedAddress, Longitude: req.Destination.Longitude, Latitude: req.Destination.Latitude}
	route, err := s.amap.DrivingRoute(ctx, origin, destination)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "route service unavailable, please retry")
	}
	arrival := depart.Add(time.Duration(route.DurationSeconds) * time.Second)
	price := data.CalculatePrice(data.DefaultPricingRule(), route.DistanceMeters, route.DurationSeconds, depart)
	return &tripv1.PreviewTripPriceReply{Origin: req.Origin, Destination: req.Destination, RouteDistanceMeters: int32(route.DistanceMeters), RouteDurationSeconds: int32(route.DurationSeconds), ArriveTime: formatTime(arrival), Price: price, PricingRuleVersion: data.DefaultPricingRule().Version}, nil
}

func (s *TripService) SearchTrips(ctx context.Context, req *tripv1.SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	items, total, err := s.uc.SearchTrips(ctx, req.Origin, req.Destination, req.DepartDate, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.SearchTripsReply{Total: total, Items: tripsToProto(items)}, nil
}

func (s *TripService) GetTripDetail(ctx context.Context, req *tripv1.GetTripDetailRequest) (*tripv1.GetTripDetailReply, error) {
	item, err := s.uc.GetTripDetail(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.GetTripDetailReply{Trip: tripToProto(item)}, nil
}

func (s *TripService) PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	depart, err := parseTime(req.DepartTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid depart_time")
	}
	if req.Origin == nil || req.Destination == nil {
		return nil, status.Error(codes.InvalidArgument, "origin and destination are required")
	}
	if s.amap == nil {
		return nil, status.Error(codes.Unavailable, "route service unavailable, please retry")
	}
	origin := data.Location{POIID: req.Origin.PoiId, Name: req.Origin.Name, FormattedAddress: req.Origin.FormattedAddress, Longitude: req.Origin.Longitude, Latitude: req.Origin.Latitude}
	destination := data.Location{POIID: req.Destination.PoiId, Name: req.Destination.Name, FormattedAddress: req.Destination.FormattedAddress, Longitude: req.Destination.Longitude, Latitude: req.Destination.Latitude}
	route, err := s.amap.DrivingRoute(ctx, origin, destination)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "route service unavailable, please retry")
	}
	arrive := depart.Add(time.Duration(route.DurationSeconds) * time.Second)
	originName := locationDisplayName(req.Origin)
	destinationName := locationDisplayName(req.Destination)
	item, err := s.uc.PublishTrip(ctx, biz.PublishTripCommand{
		RequestID:            req.RequestId,
		DriverID:             req.DriverId,
		Origin:               originName,
		OriginName:           originName,
		OriginLat:            req.Origin.Latitude,
		OriginLng:            req.Origin.Longitude,
		Destination:          destinationName,
		DestName:             destinationName,
		DestLat:              req.Destination.Latitude,
		DestLng:              req.Destination.Longitude,
		DepartTime:           depart,
		ArriveTime:           arrive,
		SeatsTotal:           int(req.SeatsTotal),
		Price:                data.CalculatePrice(data.DefaultPricingRule(), route.DistanceMeters, route.DurationSeconds, depart),
		RouteDistanceMeters:  route.DistanceMeters,
		RouteDurationSeconds: route.DurationSeconds,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.PublishTripReply{TripId: item.ID, Status: int32(item.Status), Price: item.Price, ArriveTime: formatTime(item.ArriveTime)}, nil
}

func (s *TripService) ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	items, total, err := s.uc.ListDriverTrips(ctx, req.DriverId, int(req.Status), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.ListDriverTripsReply{Total: total, Items: tripsToProto(items)}, nil
}

func (s *TripService) UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) (*tripv1.UpdateTripStatusReply, error) {
	if err := s.uc.UpdateTripStatus(ctx, req.Id, int(req.Status)); err != nil {
		return nil, mapError(err)
	}
	return &tripv1.UpdateTripStatusReply{}, nil
}

func (s *TripService) ListCoupons(ctx context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error) {
	items, total, err := s.uc.ListCoupons(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.ListCouponsReply{Total: total, Items: couponsToProto(items)}, nil
}

func (s *TripService) ClaimCoupon(ctx context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error) {
	coupon, duplicated, err := s.uc.ClaimCoupon(ctx, biz.ClaimCouponCommand{
		PassengerID:    req.UserId,
		CouponNo:       req.CouponNo,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.ClaimCouponReply{Coupon: couponToProto(coupon), Duplicated: duplicated}, nil
}

func (s *TripService) PublishDemand(ctx context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error) {
	depart, err := parseFlexibleTime(req.DepartTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid depart_time")
	}
	demand, err := s.uc.PublishDemand(ctx, biz.PublishDemandCommand{
		PassengerID: req.PassengerId,
		Origin:      req.Origin,
		Destination: req.Destination,
		DepartTime:  depart,
		Seats:       int(req.Seats),
		Budget:      req.Budget,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.PublishDemandReply{Demand: demandToProto(demand)}, nil
}

func (s *TripService) ListMyDemands(ctx context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error) {
	items, total, err := s.uc.ListMyDemands(ctx, req.PassengerId, int(req.Status), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &tripv1.ListMyDemandsReply{Total: total, Items: demandsToProto(items)}, nil
}

func (s *TripService) CancelDemand(ctx context.Context, req *tripv1.CancelDemandRequest) (*tripv1.CancelDemandReply, error) {
	if err := s.uc.CancelDemand(ctx, req.Id, req.PassengerId); err != nil {
		return nil, mapError(err)
	}
	return &tripv1.CancelDemandReply{}, nil
}

func (s *TripService) DeleteTrip(ctx context.Context, req *tripv1.DeleteTripRequest) (*tripv1.DeleteTripReply, error) {
	if err := s.uc.DeleteDriverTrip(ctx, biz.DeleteTripCommand{TripID: req.Id, DriverID: req.DriverId}); err != nil {
		return nil, mapError(err)
	}
	return &tripv1.DeleteTripReply{}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidTrip), errors.Is(err, biz.ErrInvalidCoupon), errors.Is(err, biz.ErrInvalidDemand), errors.Is(err, biz.ErrDemandCannotCancel), errors.Is(err, biz.ErrTripHasActiveOrders), errors.Is(err, biz.ErrTripTimeConflict), errors.Is(err, biz.ErrTripNotPending), errors.Is(err, biz.ErrInvalidReview), errors.Is(err, biz.ErrCouponStockExhausted), errors.Is(err, biz.ErrDuplicateTripRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrRedisUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, biz.ErrTripNotFound), errors.Is(err, biz.ErrCouponNotFound), errors.Is(err, biz.ErrDemandNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func tripsToProto(items []biz.Trip) []*tripv1.TripItem {
	out := make([]*tripv1.TripItem, len(items))
	for i := range items {
		out[i] = tripToProto(&items[i])
	}
	return out
}

func tripToProto(item *biz.Trip) *tripv1.TripItem {
	if item == nil {
		return nil
	}
	return &tripv1.TripItem{
		Id:                   item.ID,
		DriverId:             item.DriverID,
		Origin:               item.Origin,
		Destination:          item.Destination,
		DepartTime:           formatTime(item.DepartTime),
		ArriveTime:           formatTime(item.ArriveTime),
		SeatsTotal:           int32(item.SeatsTotal),
		SeatsAvailable:       int32(item.SeatsAvailable),
		Price:                item.Price,
		Status:               int32(item.Status),
		CreatedAt:            formatTime(item.CreatedAt),
		RejectReason:         item.RejectReason,
		AuditOperatorId:      item.AuditOperatorID,
		AuditTime:            formatTimePtr(item.AuditTime),
		RouteDistanceMeters:  int32(item.RouteDistanceMeters),
		RouteDurationSeconds: int32(item.RouteDurationSeconds),
		IsDeleted:            item.IsDeleted,
	}
}

func couponsToProto(items []biz.Coupon) []*tripv1.CouponItem {
	out := make([]*tripv1.CouponItem, len(items))
	for i := range items {
		out[i] = couponToProto(&items[i])
	}
	return out
}

func couponToProto(item *biz.Coupon) *tripv1.CouponItem {
	if item == nil {
		return nil
	}
	return &tripv1.CouponItem{
		Id:              item.ID,
		CouponNo:        item.CouponNo,
		CouponCode:      item.CouponCode,
		Name:            item.Name,
		CouponType:      item.CouponType,
		FaceValue:       item.FaceValue,
		DiscountRate:    item.DiscountRate,
		ThresholdAmount: item.ThresholdAmount,
		ValidFrom:       formatTime(item.ValidFrom),
		ValidTo:         formatTime(item.ValidTo),
		Status:          item.Status,
		Claimed:         item.Claimed,
	}
}

func demandsToProto(items []biz.TripDemand) []*tripv1.TripDemandItem {
	out := make([]*tripv1.TripDemandItem, len(items))
	for i := range items {
		out[i] = demandToProto(&items[i])
	}
	return out
}

func demandToProto(item *biz.TripDemand) *tripv1.TripDemandItem {
	if item == nil {
		return nil
	}
	return &tripv1.TripDemandItem{
		Id:          item.ID,
		PassengerId: item.PassengerID,
		Origin:      item.Origin,
		Destination: item.Destination,
		DepartTime:  formatTime(item.DepartTime),
		Seats:       int32(item.Seats),
		Budget:      item.Budget,
		Remark:      item.Remark,
		Status:      int32(item.Status),
		CreatedAt:   formatTime(item.CreatedAt),
	}
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func parseFlexibleTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02 15:04", value, time.Local)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func formatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatTime(*value)
}
