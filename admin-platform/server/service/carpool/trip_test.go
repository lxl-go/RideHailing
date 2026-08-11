package carpool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	commonReq "ride-hailing/admin-server/model/common/request"
)

func newTripServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&carpoolModel.Trip{}, &tripServiceOrderModel{}))
	global.GVA_DB = db
	global.GVA_LOG = zap.NewNop()
	return db
}

type tripServiceOrderModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	TripID      int64     `gorm:"column:trip_id;type:bigint;not null;index"`
	PassengerID int64     `gorm:"column:passenger_id;type:bigint;not null;default:0"`
	SeatsBooked int       `gorm:"column:seats_booked;type:int;not null;default:1"`
	TotalPrice  float64   `gorm:"column:total_price;type:decimal(10,2);not null;default:0"`
	Status      int       `gorm:"column:status;type:tinyint;not null;default:0;index"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (tripServiceOrderModel) TableName() string {
	return "carpool_order"
}

func TestListTripsExcludesDriverDeletedTrips(t *testing.T) {
	db := newTripServiceTestDB(t)
	ctx := context.Background()
	now := time.Now().Add(time.Hour)
	require.NoError(t, db.Create(&carpoolModel.Trip{
		ID:             201,
		PublisherID:    2001,
		PublisherRole:  1,
		TripType:       1,
		OriginName:     "A",
		DestName:       "B",
		DepartureTime:  now,
		SeatsTotal:     4,
		SeatsAvailable: 4,
		ShareCost:      20,
		Status:         10,
		IsDeleted:      true,
	}).Error)
	require.NoError(t, db.Create(&carpoolModel.Trip{
		ID:             202,
		PublisherID:    2001,
		PublisherRole:  1,
		TripType:       1,
		OriginName:     "C",
		DestName:       "D",
		DepartureTime:  now,
		SeatsTotal:     4,
		SeatsAvailable: 4,
		ShareCost:      20,
		Status:         10,
	}).Error)

	list, total, err := (&TripService{}).ListTrips(ctx, carpoolReq.TripListSearch{
		PageInfo: commonReq.PageInfo{Page: 1, PageSize: 20},
		Status:   intPtr(10),
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, int64(202), list[0].ID)
}

func TestReviewTripRejectsDriverDeletedTripWithBusinessError(t *testing.T) {
	db := newTripServiceTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.Create(&carpoolModel.Trip{
		ID:             203,
		PublisherID:    2001,
		PublisherRole:  1,
		TripType:       1,
		OriginName:     "A",
		DestName:       "B",
		DepartureTime:  time.Now().Add(time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 4,
		ShareCost:      20,
		Status:         5,
		IsDeleted:      true,
		Remarks:        "driver deleted",
	}).Error)

	err := (&TripService{}).ReviewTrip(ctx, 203, 88, true, "")

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrTripDeleted))
}

func TestReviewTripApprovesPendingTripAndListReflectsStatus(t *testing.T) {
	db := newTripServiceTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.Create(&carpoolModel.Trip{
		ID:             204,
		PublisherID:    2001,
		PublisherRole:  1,
		TripType:       1,
		OriginName:     "A",
		DestName:       "B",
		DepartureTime:  time.Now().Add(time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 4,
		ShareCost:      20,
		Status:         tripStatusPending,
	}).Error)

	err := (&TripService{}).ReviewTrip(ctx, 204, 88, true, "")

	require.NoError(t, err)
	var saved carpoolModel.Trip
	require.NoError(t, db.First(&saved, int64(204)).Error)
	require.Equal(t, tripStatusApproved, saved.Status)
	require.Equal(t, int64(88), saved.AuditOperatorID)
	require.NotNil(t, saved.AuditTime)

	list, total, err := (&TripService{}).ListTrips(ctx, carpoolReq.TripListSearch{
		PageInfo: commonReq.PageInfo{Page: 1, PageSize: 20},
		Status:   intPtr(tripStatusApproved),
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, int64(204), list[0].ID)
}

func TestDeactivateTripRejectsApprovedTripWithActiveOrder(t *testing.T) {
	db := newTripServiceTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.Create(&carpoolModel.Trip{
		ID:             205,
		PublisherID:    2001,
		PublisherRole:  1,
		TripType:       1,
		OriginName:     "A",
		DestName:       "B",
		DepartureTime:  time.Now().Add(time.Hour),
		SeatsTotal:     4,
		SeatsAvailable: 3,
		ShareCost:      20,
		Status:         tripStatusApproved,
	}).Error)
	require.NoError(t, db.Table("carpool_order").Create(map[string]interface{}{
		"id":           6001,
		"trip_id":      205,
		"passenger_id": 3001,
		"seats_booked": 1,
		"total_price":  20,
		"status":       4,
		"created_at":   time.Now(),
		"updated_at":   time.Now(),
	}).Error)

	err := (&TripService{}).DeactivateTrip(ctx, 205, "driver request")

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrTripHasActiveOrder))
	var saved carpoolModel.Trip
	require.NoError(t, db.First(&saved, int64(205)).Error)
	require.False(t, saved.IsDeleted)
	require.Equal(t, tripStatusApproved, saved.Status)
}

func intPtr(value int) *int {
	return &value
}
