package data

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/trip-service/internal/biz"
)

func TestTripRepoCreateSearchAndListByDriver(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripModel{}))

	repo := NewTripRepo(db, zap.NewNop())
	depart := time.Now().Add(24 * time.Hour)
	trip := &biz.Trip{
		ID:             101,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     depart,
		ArriveTime:     depart.Add(time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 4,
		Price:          22.5,
		Status:         biz.TripStatusApproved,
	}

	require.NoError(t, repo.Create(context.Background(), trip))

	items, total, err := repo.SearchTrips(context.Background(), "A", "B", "", 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(101), items[0].ID)

	driverItems, driverTotal, err := repo.ListByDriver(context.Background(), 2001, biz.TripStatusApproved, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), driverTotal)
	require.Len(t, driverItems, 1)
}

func TestTripRepoSearchTripsOnlyReturnsBookableTrips(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripModel{}))
	repo := NewTripRepo(db, zap.NewNop())
	depart := time.Now().Add(2 * time.Hour)
	trips := []tripModel{
		{ID: 101, DriverID: 2001, Origin: "A", Destination: "B", DepartTime: depart, SeatsTotal: 4, SeatsAvailable: 3, Price: 22.5, Status: biz.TripStatusApproved},
		{ID: 102, DriverID: 2002, Origin: "A", Destination: "B", DepartTime: depart, SeatsTotal: 4, SeatsAvailable: 0, Price: 22.5, Status: biz.TripStatusApproved},
		{ID: 103, DriverID: 2003, Origin: "A", Destination: "B", DepartTime: time.Now().Add(-time.Hour), SeatsTotal: 4, SeatsAvailable: 3, Price: 22.5, Status: biz.TripStatusApproved},
		{ID: 104, DriverID: 2004, Origin: "A", Destination: "B", DepartTime: depart, SeatsTotal: 4, SeatsAvailable: 3, Price: 22.5, Status: biz.TripStatusPending},
		{ID: 105, DriverID: 2005, Origin: "A", Destination: "B", DepartTime: depart, SeatsTotal: 4, SeatsAvailable: 3, Price: 22.5, Status: biz.TripStatusApproved, IsDeleted: true},
	}
	require.NoError(t, db.Create(&trips).Error)

	items, total, err := repo.SearchTrips(context.Background(), "A", "B", "", 1, 20)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(101), items[0].ID)
}

func TestTripRepoClaimCouponIsIdempotentAndControlsStock(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&couponTemplateModel{}, &userCouponModel{}))
	require.NoError(t, db.Create(&couponTemplateModel{
		ID:           71001,
		CouponNo:     "CP202607290001",
		Name:         "New Passenger Cash",
		CouponType:   "cash",
		FaceValue:    20,
		ValidFrom:    time.Now().Add(-time.Hour),
		ValidTo:      time.Now().Add(time.Hour),
		ServiceScope: "carpool",
		TotalStock:   1,
		Status:       couponStatusEnabled,
	}).Error)
	repo := NewTripRepo(db, zap.NewNop())

	first, duplicated, err := repo.ClaimCoupon(context.Background(), 3001, "CP202607290001", "claim-1")
	require.NoError(t, err)
	require.False(t, duplicated)
	require.Equal(t, "CP202607290001", first.CouponNo)
	require.NotEmpty(t, first.CouponCode)

	second, duplicated, err := repo.ClaimCoupon(context.Background(), 3001, "CP202607290001", "claim-1")
	require.NoError(t, err)
	require.True(t, duplicated)
	require.Equal(t, first.CouponCode, second.CouponCode)

	_, _, err = repo.ClaimCoupon(context.Background(), 3002, "CP202607290001", "claim-2")
	require.ErrorIs(t, err, biz.ErrCouponStockExhausted)
}

func TestTripRepoDemandLifecyclePersistsAndCancelsByOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripDemandModel{}))
	repo := NewTripRepo(db, zap.NewNop())
	demand := &biz.TripDemand{
		ID:          81001,
		PassengerID: 3001,
		Origin:      "A",
		Destination: "B",
		DepartTime:  time.Date(2026, 8, 5, 8, 30, 0, 0, time.UTC),
		Seats:       2,
		Budget:      45.5,
		Remark:      "near gate",
		Status:      biz.DemandStatusPending,
	}

	require.NoError(t, repo.CreateDemand(context.Background(), demand))
	items, total, err := repo.ListDemandsByPassenger(context.Background(), 3001, biz.DemandStatusPending, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, int64(81001), items[0].ID)

	require.NoError(t, repo.CancelDemand(context.Background(), 81001, 3001))
	items, total, err = repo.ListDemandsByPassenger(context.Background(), 3001, biz.DemandStatusCancelled, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, biz.DemandStatusCancelled, items[0].Status)
}

func TestTripRepoDeleteDriverTripRejectsActiveOrders(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripModel{}, &orderModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             101,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Now().Add(time.Hour),
		ArriveTime:     time.Now().Add(2 * time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 3,
		Price:          20,
		Status:         biz.TripStatusApproved,
	}).Error)
	require.NoError(t, db.Create(&orderModel{ID: 5001, TripID: 101, Status: activeOrderStatusPaid}).Error)
	repo := NewTripRepo(db, zap.NewNop())

	err = repo.DeleteDriverTrip(context.Background(), 101, 2001)

	require.ErrorIs(t, err, biz.ErrTripHasActiveOrders)
}

func TestTripRepoDeleteDriverTripAllowsApprovedTripWithoutOrders(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripModel{}, &orderModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             103,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Now().Add(time.Hour),
		ArriveTime:     time.Now().Add(2 * time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 4,
		Price:          20,
		Status:         biz.TripStatusApproved,
	}).Error)
	repo := NewTripRepo(db, zap.NewNop())

	err = repo.DeleteDriverTrip(context.Background(), 103, 2001)

	require.NoError(t, err)
	var saved tripModel
	require.NoError(t, db.First(&saved, int64(103)).Error)
	require.True(t, saved.IsDeleted)
	require.Equal(t, biz.TripStatusCancelled, saved.Status)
}

func TestTripRepoDeleteDriverTripCancelsAndHidesTrip(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripModel{}, &orderModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             102,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Now().Add(time.Hour),
		ArriveTime:     time.Now().Add(2 * time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 4,
		Price:          20,
		Status:         biz.TripStatusPending,
	}).Error)
	repo := NewTripRepo(db, zap.NewNop())

	err = repo.DeleteDriverTrip(context.Background(), 102, 2001)

	require.NoError(t, err)
	var saved tripModel
	require.NoError(t, db.First(&saved, int64(102)).Error)
	require.True(t, saved.IsDeleted)
	require.Equal(t, biz.TripStatusCancelled, saved.Status)

	items, total, err := repo.ListByDriver(context.Background(), 2001, 0, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, items)
}

func TestTripRepoDeleteDriverTripIsIdempotentAfterOwnerDeleted(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&tripModel{}, &orderModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             104,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Now().Add(time.Hour),
		ArriveTime:     time.Now().Add(2 * time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 4,
		Price:          20,
		Status:         biz.TripStatusCancelled,
		IsDeleted:      true,
	}).Error)
	repo := NewTripRepo(db, zap.NewNop())

	err = repo.DeleteDriverTrip(context.Background(), 104, 2001)

	require.NoError(t, err)
}
