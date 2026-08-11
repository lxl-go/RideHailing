package service

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
	"ride-hailing/services/trip-service/internal/biz"
)

type serviceFakeRepo struct {
	items   []biz.Trip
	coupons []biz.Coupon
	demands []biz.TripDemand
}

func (r *serviceFakeRepo) SearchTrips(context.Context, string, string, string, int, int) ([]biz.Trip, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func (r *serviceFakeRepo) GetByID(_ context.Context, id int64) (*biz.Trip, error) {
	return &r.items[0], nil
}

func (r *serviceFakeRepo) Create(_ context.Context, trip *biz.Trip) error {
	r.items = append(r.items, *trip)
	return nil
}

func (r *serviceFakeRepo) ListByDriver(context.Context, int64, int, int, int) ([]biz.Trip, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func (r *serviceFakeRepo) UpdateStatus(context.Context, int64, int) error {
	return nil
}

func (r *serviceFakeRepo) ListCoupons(context.Context, int64, int, int) ([]biz.Coupon, int64, error) {
	return r.coupons, int64(len(r.coupons)), nil
}

func (r *serviceFakeRepo) ClaimCoupon(context.Context, int64, string, string) (*biz.Coupon, bool, error) {
	coupon := biz.Coupon{ID: 9001, CouponNo: "CP202607290001", CouponCode: "UC1001", Name: "New Passenger Cash", Status: biz.CouponStatusUnused, Claimed: true}
	return &coupon, false, nil
}

func (r *serviceFakeRepo) CreateDemand(_ context.Context, demand *biz.TripDemand) error {
	r.demands = append(r.demands, *demand)
	return nil
}

func (r *serviceFakeRepo) ListDemandsByPassenger(context.Context, int64, int, int, int) ([]biz.TripDemand, int64, error) {
	return r.demands, int64(len(r.demands)), nil
}

func (r *serviceFakeRepo) CancelDemand(context.Context, int64, int64) error {
	return nil
}

func (r *serviceFakeRepo) DeleteDriverTrip(context.Context, int64, int64) error {
	return nil
}

func TestSearchTripsMapsDomainToReply(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	uc := biz.NewTripUsecase(node, zap.NewNop(), &serviceFakeRepo{items: []biz.Trip{{
		ID:             101,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC),
		ArriveTime:     time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC),
		SeatsTotal:     4,
		SeatsAvailable: 3,
		Price:          20,
		Status:         biz.TripStatusRecruiting,
	}}})
	svc := NewTripService(uc)

	reply, err := svc.SearchTrips(context.Background(), &tripv1.SearchTripsRequest{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Equal(t, int64(1), reply.Total)
	require.Equal(t, int64(101), reply.Items[0].Id)
	require.Equal(t, int32(3), reply.Items[0].SeatsAvailable)
}

func TestClaimCouponMapsReply(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	uc := biz.NewTripUsecase(node, zap.NewNop(), &serviceFakeRepo{})
	svc := NewTripService(uc)

	reply, err := svc.ClaimCoupon(context.Background(), &tripv1.ClaimCouponRequest{
		UserId:         3001,
		CouponNo:       "CP202607290001",
		IdempotencyKey: "claim-1",
	})

	require.NoError(t, err)
	require.Equal(t, "CP202607290001", reply.Coupon.CouponNo)
	require.Equal(t, "UC1001", reply.Coupon.CouponCode)
	require.True(t, reply.Coupon.Claimed)
}

func TestPublishDemandMapsReply(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &serviceFakeRepo{}
	uc := biz.NewTripUsecase(node, zap.NewNop(), repo)
	svc := NewTripService(uc)

	reply, err := svc.PublishDemand(context.Background(), &tripv1.PublishDemandRequest{
		PassengerId: 3001,
		Origin:      "A",
		Destination: "B",
		DepartTime:  "2026-08-05 08:30",
		Seats:       2,
		Budget:      45.5,
		Remark:      "near gate",
	})

	require.NoError(t, err)
	require.NotZero(t, reply.Demand.Id)
	require.Equal(t, int64(3001), reply.Demand.PassengerId)
	require.Equal(t, int32(biz.DemandStatusPending), reply.Demand.Status)
}
