package carpool

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

type testCarpoolOrder struct {
	ID          uint64     `gorm:"column:id;primaryKey"`
	TripID      uint64     `gorm:"column:trip_id"`
	PassengerID uint64     `gorm:"column:passenger_id"`
	SeatsBooked int        `gorm:"column:seats_booked"`
	TotalPrice  float64    `gorm:"column:total_price"`
	Status      string     `gorm:"column:status"`
	AcceptedAt  *time.Time `gorm:"column:accepted_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (testCarpoolOrder) TableName() string {
	return "carpool_order"
}

type testCarpoolTrip struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	DriverID       uint64    `gorm:"column:driver_id"`
	StartLocation  string    `gorm:"column:start_location"`
	EndLocation    string    `gorm:"column:end_location"`
	DepartureTime  time.Time `gorm:"column:departure_time"`
	AvailableSeats int       `gorm:"column:available_seats"`
	PricePerSeat   float64   `gorm:"column:price_per_seat"`
	Status         string    `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (testCarpoolTrip) TableName() string {
	return "carpool_trip"
}

type testCarpoolPayment struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	OrderID     uint64    `gorm:"column:order_id"`
	OutTradeNo  string    `gorm:"column:out_trade_no"`
	TradeNo     string    `gorm:"column:trade_no"`
	TotalAmount float64   `gorm:"column:total_amount"`
	Status      string    `gorm:"column:status"`
	PaidAt      time.Time `gorm:"column:paid_at"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (testCarpoolPayment) TableName() string {
	return "carpool_payment"
}

func newOrderServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&testCarpoolTrip{},
		&testCarpoolOrder{},
		&testCarpoolPayment{},
		&carpoolModel.OrderRefund{},
		&carpoolModel.OrderStatusHistory{},
	))
	global.GVA_DB = db
	return db
}

func TestOrderServiceListsRealCarpoolOrdersWithStringIDs(t *testing.T) {
	db := newOrderServiceTestDB(t)
	now := time.Date(2026, 8, 4, 10, 0, 0, 0, time.Local)
	bigOrderID := uint64(9007199254740993)
	require.NoError(t, db.Create(&testCarpoolTrip{
		ID: 7001, DriverID: 9007199254740995, StartLocation: "Hongqiao", EndLocation: "Pudong",
		DepartureTime: now.Add(2 * time.Hour), AvailableSeats: 2, PricePerSeat: 64.275, Status: "published",
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&testCarpoolOrder{
		ID: bigOrderID, TripID: 7001, PassengerID: 9007199254740994, SeatsBooked: 2,
		TotalPrice: 128.55, Status: "paid", CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&testCarpoolPayment{
		ID: 8001, OrderID: bigOrderID, OutTradeNo: "ALIPAY-ORDER-001", TradeNo: "TRADE-001",
		TotalAmount: 128.55, Status: "paid", PaidAt: now, CreatedAt: now, UpdatedAt: now,
	}).Error)

	list, total, err := (&OrderService{}).ListOrders(context.Background(), carpoolReq.OrderSearch{Status: "paid"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, strconv.FormatUint(bigOrderID, 10), list[0].OrderNo)
	require.Equal(t, "carpool", list[0].ServiceType)
	require.Equal(t, 128.55, list[0].PayAmount)

	overview, err := (&OrderService{}).GetOverview(context.Background(), carpoolReq.OrderSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, overview.TotalOrders)
	require.EqualValues(t, 1, overview.PaidOrders)
	require.Equal(t, 128.55, overview.Revenue)
}

func TestOrderServiceRefundRulesHistoryAndBatch(t *testing.T) {
	db := newOrderServiceTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&[]testCarpoolTrip{
		{ID: 11001, DriverID: 2001, StartLocation: "Line", EndLocation: "One", DepartureTime: now.Add(3 * time.Hour), Status: "published"},
		{ID: 11002, DriverID: 2002, StartLocation: "Airport", EndLocation: "City", DepartureTime: now.Add(60 * time.Minute), Status: "published"},
		{ID: 11003, DriverID: 2003, StartLocation: "Railway", EndLocation: "City", DepartureTime: now.Add(24 * time.Hour), Status: "published"},
	}).Error)
	require.NoError(t, db.Create(&[]testCarpoolOrder{
		{ID: 10001, TripID: 11001, PassengerID: 1001, SeatsBooked: 1, TotalPrice: 100, Status: "paid", CreatedAt: now, UpdatedAt: now},
		{ID: 10002, TripID: 11002, PassengerID: 1002, SeatsBooked: 1, TotalPrice: 80, Status: "paid", CreatedAt: now, UpdatedAt: now},
		{ID: 10003, TripID: 11003, PassengerID: 1003, SeatsBooked: 1, TotalPrice: 120, Status: "completed", CreatedAt: now, UpdatedAt: now},
	}).Error)

	service := OrderService{}
	ctx := context.Background()

	full, err := service.ApplyRefund(ctx, carpoolReq.OrderRefundApply{
		OrderNo: "10001", Reason: "plan changed", IdempotentKey: "idem-full", Operator: "passenger",
	})
	require.NoError(t, err)
	require.Equal(t, "approved", full.Status)
	require.Equal(t, "auto", full.ReviewType)
	require.Equal(t, 100.00, full.RefundAmount)
	require.Equal(t, 0.00, full.CancelFee)

	again, err := service.ApplyRefund(ctx, carpoolReq.OrderRefundApply{
		OrderNo: "10001", Reason: "repeat click", IdempotentKey: "idem-full", Operator: "passenger",
	})
	require.NoError(t, err)
	require.Equal(t, full.RefundNo, again.RefundNo)

	partial, err := service.ApplyRefund(ctx, carpoolReq.OrderRefundApply{
		OrderNo: "10002", Reason: "temporary issue", IdempotentKey: "idem-partial", Operator: "passenger",
	})
	require.NoError(t, err)
	require.Equal(t, "approved", partial.Status)
	require.Equal(t, 72.00, partial.RefundAmount)
	require.Equal(t, 8.00, partial.CancelFee)

	manual, err := service.ApplyRefund(ctx, carpoolReq.OrderRefundApply{
		OrderNo: "10003", Reason: "service dispute", IdempotentKey: "idem-manual", Operator: "customer-service",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", manual.Status)
	require.Equal(t, "manual", manual.ReviewType)

	history, err := service.GetStatusHistory(ctx, "10001")
	require.NoError(t, err)
	require.NotEmpty(t, history)
	require.Equal(t, "paid", history[0].FromStatus)
	require.Equal(t, "refunded", history[0].ToStatus)

	batch, err := service.BatchRefund(ctx, carpoolReq.OrderBatchRefund{
		OrderNos: []string{"10001", "999999"}, Reason: "weather", Operator: "admin", IdempotentSeed: "batch-weather",
	})
	require.NoError(t, err)
	require.Len(t, batch.Items, 2)
	require.True(t, batch.Items[0].Success)
	require.False(t, batch.Items[1].Success)
}

func TestOrderServiceOverviewAggregatesOrders(t *testing.T) {
	db := newOrderServiceTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&[]testCarpoolTrip{
		{ID: 12001, DriverID: 2001, StartLocation: "A", EndLocation: "B", DepartureTime: now, Status: "published"},
		{ID: 12002, DriverID: 2002, StartLocation: "B", EndLocation: "C", DepartureTime: now, Status: "published"},
		{ID: 12003, DriverID: 2003, StartLocation: "C", EndLocation: "D", DepartureTime: now, Status: "published"},
	}).Error)
	require.NoError(t, db.Create(&[]testCarpoolOrder{
		{ID: 12011, TripID: 12001, PassengerID: 1001, SeatsBooked: 1, TotalPrice: 80, Status: "paid", CreatedAt: now, UpdatedAt: now},
		{ID: 12012, TripID: 12002, PassengerID: 1002, SeatsBooked: 1, TotalPrice: 120, Status: "completed", CreatedAt: now, UpdatedAt: now},
		{ID: 12013, TripID: 12003, PassengerID: 1003, SeatsBooked: 1, TotalPrice: 60, Status: "refunding", CreatedAt: now, UpdatedAt: now},
	}).Error)

	overview, err := (&OrderService{}).GetOverview(context.Background(), carpoolReq.OrderSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 3, overview.TotalOrders)
	require.EqualValues(t, 1, overview.PaidOrders)
	require.EqualValues(t, 1, overview.CompletedOrders)
	require.EqualValues(t, 1, overview.RefundingOrders)
	require.Equal(t, 200.0, overview.Revenue)
}
