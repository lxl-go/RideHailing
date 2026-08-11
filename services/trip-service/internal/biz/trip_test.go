package biz

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeTripRepo struct {
	items       []Trip
	coupons     []Coupon
	demands     []TripDemand
	claimedCode string
	deletedTrip int64
}

func (r *fakeTripRepo) SearchTrips(context.Context, string, string, string, int, int) ([]Trip, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func (r *fakeTripRepo) GetByID(_ context.Context, id int64) (*Trip, error) {
	for _, item := range r.items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, ErrTripNotFound
}

func (r *fakeTripRepo) Create(_ context.Context, trip *Trip) error {
	r.items = append(r.items, *trip)
	return nil
}

func (r *fakeTripRepo) ListByDriver(_ context.Context, driverID int64, _ int, _ int, _ int) ([]Trip, int64, error) {
	var out []Trip
	for _, item := range r.items {
		if item.DriverID == driverID {
			out = append(out, item)
		}
	}
	return out, int64(len(out)), nil
}

func (r *fakeTripRepo) UpdateStatus(_ context.Context, id int64, status int) error {
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Status = status
			return nil
		}
	}
	return ErrTripNotFound
}

func (r *fakeTripRepo) ListCoupons(_ context.Context, passengerID int64, page, pageSize int) ([]Coupon, int64, error) {
	return r.coupons, int64(len(r.coupons)), nil
}

func (r *fakeTripRepo) ClaimCoupon(_ context.Context, passengerID int64, couponNo string, idempotencyKey string) (*Coupon, bool, error) {
	coupon := Coupon{ID: 9001, CouponNo: couponNo, CouponCode: r.claimedCode, Name: "New Passenger Cash", Status: CouponStatusUnused, Claimed: true}
	r.coupons = append(r.coupons, coupon)
	return &coupon, false, nil
}

func (r *fakeTripRepo) CreateDemand(_ context.Context, demand *TripDemand) error {
	r.demands = append(r.demands, *demand)
	return nil
}

func (r *fakeTripRepo) ListDemandsByPassenger(_ context.Context, passengerID int64, status int, page, pageSize int) ([]TripDemand, int64, error) {
	var out []TripDemand
	for _, demand := range r.demands {
		if demand.PassengerID == passengerID {
			out = append(out, demand)
		}
	}
	return out, int64(len(out)), nil
}

func (r *fakeTripRepo) CancelDemand(_ context.Context, id int64, passengerID int64) error {
	for i := range r.demands {
		if r.demands[i].ID == id && r.demands[i].PassengerID == passengerID {
			r.demands[i].Status = DemandStatusCancelled
			return nil
		}
	}
	return ErrDemandNotFound
}

func (r *fakeTripRepo) DeleteDriverTrip(_ context.Context, id int64, driverID int64) error {
	r.deletedTrip = id
	return nil
}

func TestPublishTripDefaultsSeatsAndStatus(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &fakeTripRepo{}
	uc := NewTripUsecase(node, zap.NewNop(), repo)

	trip, err := uc.PublishTrip(context.Background(), PublishTripCommand{
		DriverID:    2001,
		Origin:      "A",
		Destination: "B",
		DepartTime:  time.Now().Add(20 * time.Minute),
		ArriveTime:  time.Now().Add(80 * time.Minute),
		SeatsTotal:  3,
		Price:       19.9,
	})

	require.NoError(t, err)
	require.NotZero(t, trip.ID)
	require.Equal(t, 3, trip.SeatsAvailable)
	require.Equal(t, TripStatusPending, trip.Status)
	require.Len(t, repo.items, 1)
}

func TestPublishTripPersistsDisplayLocationsAndCoordinates(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &fakeTripRepo{}
	uc := NewTripUsecase(node, zap.NewNop(), repo)

	trip, err := uc.PublishTrip(context.Background(), PublishTripCommand{
		DriverID:    2001,
		Origin:      "Old origin address",
		OriginName:  "Hangzhou East Railway Station",
		OriginLat:   30.29123,
		OriginLng:   120.21234,
		Destination: "Old destination address",
		DestName:    "Xihu Cultural Square",
		DestLat:     30.27987,
		DestLng:     120.16543,
		DepartTime:  time.Now().Add(20 * time.Minute),
		ArriveTime:  time.Now().Add(80 * time.Minute),
		SeatsTotal:  3,
		Price:       19.9,
	})

	require.NoError(t, err)
	require.Equal(t, "Hangzhou East Railway Station", trip.Origin)
	require.Equal(t, "Hangzhou East Railway Station", trip.OriginName)
	require.Equal(t, "Xihu Cultural Square", trip.Destination)
	require.Equal(t, "Xihu Cultural Square", trip.DestName)
	require.Equal(t, 30.29123, trip.OriginLat)
	require.Equal(t, 120.21234, trip.OriginLng)
	require.Equal(t, 30.27987, trip.DestLat)
	require.Equal(t, 120.16543, trip.DestLng)
	require.Equal(t, trip.OriginName, repo.items[0].OriginName)
	require.Equal(t, trip.DestName, repo.items[0].DestName)
}

func TestClaimCouponDelegatesToRepoWithPassengerAndCoupon(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &fakeTripRepo{claimedCode: "UC1001"}
	uc := NewTripUsecase(node, zap.NewNop(), repo)

	coupon, duplicated, err := uc.ClaimCoupon(context.Background(), ClaimCouponCommand{
		PassengerID:    3001,
		CouponNo:       "CP202607290001",
		IdempotencyKey: "claim-1",
	})

	require.NoError(t, err)
	require.False(t, duplicated)
	require.Equal(t, "CP202607290001", coupon.CouponNo)
	require.Equal(t, "UC1001", coupon.CouponCode)
	require.True(t, coupon.Claimed)
}

func TestPublishDemandCreatesPendingPassengerDemand(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &fakeTripRepo{}
	uc := NewTripUsecase(node, zap.NewNop(), repo)
	depart := time.Date(2026, 8, 5, 8, 30, 0, 0, time.UTC)

	demand, err := uc.PublishDemand(context.Background(), PublishDemandCommand{
		PassengerID: 3001,
		Origin:      "A",
		Destination: "B",
		DepartTime:  depart,
		Seats:       2,
		Budget:      45.5,
		Remark:      "near gate",
	})

	require.NoError(t, err)
	require.NotZero(t, demand.ID)
	require.Equal(t, int64(3001), demand.PassengerID)
	require.Equal(t, DemandStatusPending, demand.Status)
	require.Len(t, repo.demands, 1)
}

func TestDeleteDriverTripRejectsInvalidOwnerInput(t *testing.T) {
	node, err := snowflake.NewNode(1)
	require.NoError(t, err)
	repo := &fakeTripRepo{}
	uc := NewTripUsecase(node, zap.NewNop(), repo)

	err = uc.DeleteDriverTrip(context.Background(), DeleteTripCommand{TripID: 0, DriverID: 2001})

	require.ErrorIs(t, err, ErrInvalidTrip)
	require.Zero(t, repo.deletedTrip)
}
