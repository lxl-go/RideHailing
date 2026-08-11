package data

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/order-service/internal/biz"
)

func TestOrderRepoCreateAtomicDecrementsTripSeats(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}))
	depart := time.Now().Add(time.Hour)
	require.NoError(t, db.Create(&tripModel{
		ID:             1001,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     depart,
		SeatsAvailable: 4,
		Price:          20,
		Status:         biz.TripStatusRecruiting,
	}).Error)

	repo := NewOrderRepo(db, zap.NewNop())
	err = repo.CreateAtomic(context.Background(), &biz.Order{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 2,
		TotalPrice:  40,
		Status:      biz.OrderStatusPending,
	})

	require.NoError(t, err)
	var trip tripModel
	require.NoError(t, db.First(&trip, "id = ?", 1001).Error)
	require.Equal(t, 2, trip.SeatsAvailable)
}

func TestOrderRepoCreateAtomicAcceptsApprovedTripStatusFromTripService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             1001,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Now().Add(time.Hour),
		SeatsAvailable: 4,
		Price:          20,
		Status:         20,
	}).Error)

	repo := NewOrderRepo(db, zap.NewNop())
	err = repo.CreateAtomic(context.Background(), &biz.Order{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 2,
		TotalPrice:  40,
		Status:      biz.OrderStatusPending,
	})

	require.NoError(t, err)
	var trip tripModel
	require.NoError(t, db.First(&trip, "id = ?", 1001).Error)
	require.Equal(t, 2, trip.SeatsAvailable)
}

func TestOrderRepoCreateAtomicRejectsInsufficientSeatsWithoutOrder(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             1001,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC),
		SeatsAvailable: 1,
		Price:          20,
		Status:         biz.TripStatusRecruiting,
	}).Error)

	repo := NewOrderRepo(db, zap.NewNop())
	err = repo.CreateAtomic(context.Background(), &biz.Order{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 2,
		TotalPrice:  40,
		Status:      biz.OrderStatusPending,
	})

	require.ErrorIs(t, err, biz.ErrInsufficientSeats)
	var count int64
	require.NoError(t, db.Model(&orderModel{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
	var trip tripModel
	require.NoError(t, db.First(&trip, "id = ?", 1001).Error)
	require.Equal(t, 1, trip.SeatsAvailable)
}

func TestOrderRepoListByDriverReturnsOnlyDriverActiveOrders(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:          1001,
		DriverID:    2001,
		Origin:      "Shanghai Station",
		Destination: "Hongqiao Airport",
		Status:      biz.TripStatusRecruiting,
	}).Error)
	require.NoError(t, db.Create(&tripModel{
		ID:          1002,
		DriverID:    2999,
		Origin:      "Other Origin",
		Destination: "Other Destination",
		Status:      biz.TripStatusRecruiting,
	}).Error)
	now := time.Now()
	require.NoError(t, db.Create(&[]orderModel{
		{ID: 5001, TripID: 1001, PassengerID: 3001, SeatsBooked: 1, TotalPrice: 39.8, Status: biz.OrderStatusAccepted, AcceptedAt: &now},
		{ID: 5002, TripID: 1001, PassengerID: 3002, SeatsBooked: 1, TotalPrice: 39.8, Status: biz.OrderStatusPaid},
		{ID: 5003, TripID: 1002, PassengerID: 3003, SeatsBooked: 1, TotalPrice: 39.8, Status: biz.OrderStatusPickingUp, AcceptedAt: &now},
		{ID: 5004, TripID: 1001, PassengerID: 3004, SeatsBooked: 1, TotalPrice: 39.8, Status: biz.OrderStatusCompleted, AcceptedAt: &now},
	}).Error)
	repo := NewOrderRepo(db, zap.NewNop())

	orders, total, err := repo.ListByDriver(context.Background(), 2001, -1, 1, 20)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, orders, 1)
	require.Equal(t, int64(5001), orders[0].ID)
	require.Equal(t, int64(2001), orders[0].DriverID)

	orders, total, err = repo.ListByDriver(context.Background(), 2001, biz.OrderStatusAccepted, 1, 20)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, orders, 1)
	require.Equal(t, int64(5001), orders[0].ID)
}

func TestOrderRepoTransitionOrderStatusUsesExpectedStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}, &orderIdempotencyModel{}))
	require.NoError(t, db.Create(&orderModel{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 1,
		TotalPrice:  39.8,
		Status:      biz.OrderStatusPaid,
	}).Error)
	repo := NewOrderRepo(db, zap.NewNop())

	err = repo.ApplyOrderTransition(context.Background(), biz.OrderTransition{
		OrderID:        5001,
		ActorID:        2001,
		Action:         biz.OrderActionAccept,
		IdempotencyKey: "accept-5001",
		FromStatuses:   []int{biz.OrderStatusPaid},
		ToStatus:       biz.OrderStatusAccepted,
	})
	require.NoError(t, err)

	err = repo.ApplyOrderTransition(context.Background(), biz.OrderTransition{
		OrderID:        5001,
		ActorID:        2001,
		Action:         biz.OrderActionComplete,
		IdempotencyKey: "complete-too-early",
		FromStatuses:   []int{biz.OrderStatusDelivering},
		ToStatus:       biz.OrderStatusCompleted,
	})
	require.ErrorIs(t, err, biz.ErrOrderCannotComplete)
	var order orderModel
	require.NoError(t, db.First(&order, "id = ?", 5001).Error)
	require.Equal(t, biz.OrderStatusAccepted, order.Status)
}

func TestOrderRepoCancelWithSeatRestoreIsTransactional(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}, &orderIdempotencyModel{}))
	require.NoError(t, db.Create(&orderModel{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 2,
		TotalPrice:  39.8,
		Status:      biz.OrderStatusPaid,
	}).Error)
	repo := NewOrderRepo(db, zap.NewNop())

	err = repo.ApplyOrderTransition(context.Background(), biz.OrderTransition{
		OrderID:        5001,
		ActorID:        3001,
		Action:         biz.OrderActionCancel,
		IdempotencyKey: "cancel-5001",
		FromStatuses:   []int{biz.OrderStatusPending, biz.OrderStatusPaid},
		ToStatus:       biz.OrderStatusCancelled,
		RestoreTripID:  9999,
		RestoreSeats:   2,
	})

	require.ErrorIs(t, err, biz.ErrTripNotFound)
	var order orderModel
	require.NoError(t, db.First(&order, "id = ?", 5001).Error)
	require.Equal(t, biz.OrderStatusPaid, order.Status)
}

func TestOrderRepoIdempotentActionDoesNotApplyTwice(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}, &orderIdempotencyModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             1001,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC),
		SeatsAvailable: 0,
		Price:          20,
		Status:         biz.TripStatusRecruiting,
	}).Error)
	require.NoError(t, db.Create(&orderModel{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 2,
		TotalPrice:  39.8,
		Status:      biz.OrderStatusPaid,
	}).Error)
	repo := NewOrderRepo(db, zap.NewNop())
	cmd := biz.OrderTransition{
		OrderID:        5001,
		ActorID:        2001,
		Action:         biz.OrderActionReject,
		IdempotencyKey: "reject-repeat-5001",
		FromStatuses:   []int{biz.OrderStatusPaid},
		ToStatus:       biz.OrderStatusCancelled,
		RestoreTripID:  1001,
		RestoreSeats:   2,
	}

	require.NoError(t, repo.ApplyOrderTransition(context.Background(), cmd))
	require.NoError(t, repo.ApplyOrderTransition(context.Background(), cmd))
	var trip tripModel
	require.NoError(t, db.First(&trip, "id = ?", 1001).Error)
	require.Equal(t, 2, trip.SeatsAvailable)
}

func TestOrderRepoRejectRefundsPaymentAndStoresReasonInTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}, &paymentModel{}, &orderIdempotencyModel{}))
	require.NoError(t, db.Create(&tripModel{
		ID:             1001,
		DriverID:       2001,
		Origin:         "A",
		Destination:    "B",
		DepartTime:     time.Now().Add(time.Hour),
		SeatsAvailable: 0,
		Price:          39.8,
		Status:         biz.TripStatusRecruiting,
	}).Error)
	require.NoError(t, db.Create(&orderModel{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 1,
		TotalPrice:  39.8,
		Status:      biz.OrderStatusPaid,
	}).Error)
	require.NoError(t, db.Create(&paymentModel{
		ID:          7001,
		OrderID:     5001,
		PassengerID: 3001,
		OutTradeNo:  "PAY7001",
		Channel:     biz.PaymentChannelAlipaySandbox,
		TotalAmount: "39.80",
		Status:      biz.PaymentStatusPaid,
	}).Error)
	rejectedAt := time.Now()
	refundedAt := rejectedAt.Add(time.Second)
	repo := NewOrderRepo(db, zap.NewNop())

	err = repo.ApplyOrderTransition(context.Background(), biz.OrderTransition{
		OrderID:        5001,
		ActorID:        2001,
		Action:         biz.OrderActionReject,
		IdempotencyKey: "reject-5001",
		FromStatuses:   []int{biz.OrderStatusPaid},
		ToStatus:       biz.OrderStatusCancelled,
		RestoreTripID:  1001,
		RestoreSeats:   1,
		RejectReason:   "vehicle fault",
		RejectedAt:     &rejectedAt,
		RefundAmount:   39.8,
		RefundedAt:     &refundedAt,
		RefundPayment:  true,
	})

	require.NoError(t, err)
	var order orderModel
	require.NoError(t, db.First(&order, "id = ?", 5001).Error)
	require.Equal(t, biz.OrderStatusCancelled, order.Status)
	require.Equal(t, "vehicle fault", order.RejectReason)
	require.NotNil(t, order.RejectedAt)
	require.NotNil(t, order.RefundedAt)
	require.Equal(t, 39.8, order.RefundAmount)

	var payment paymentModel
	require.NoError(t, db.First(&payment, "order_id = ?", 5001).Error)
	require.Equal(t, biz.PaymentStatusRefunded, payment.Status)

	var trip tripModel
	require.NoError(t, db.First(&trip, "id = ?", 1001).Error)
	require.Equal(t, 1, trip.SeatsAvailable)
}

func TestOrderRepoMarkPaymentPaidUpdatesPaymentAndOrder(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&orderModel{}, &tripModel{}, &paymentModel{}))
	require.NoError(t, db.Create(&orderModel{
		ID:          5001,
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 1,
		TotalPrice:  39.8,
		Status:      biz.OrderStatusPending,
	}).Error)
	require.NoError(t, db.Create(&paymentModel{
		ID:          7001,
		OrderID:     5001,
		PassengerID: 3001,
		OutTradeNo:  "PAY7001",
		Channel:     biz.PaymentChannelAlipaySandbox,
		TotalAmount: "39.80",
		Status:      biz.PaymentStatusPending,
	}).Error)

	repo := NewOrderRepo(db, zap.NewNop())
	payment, duplicated, err := repo.MarkPaymentPaid(context.Background(), "PAY7001", "2026080422001", "raw=notify")

	require.NoError(t, err)
	require.False(t, duplicated)
	require.Equal(t, biz.PaymentStatusPaid, payment.Status)
	require.NotNil(t, payment.PaidAt)
	require.Equal(t, "2026080422001", payment.AlipayTradeNo)

	var order orderModel
	require.NoError(t, db.First(&order, "id = ?", 5001).Error)
	require.Equal(t, biz.OrderStatusPaid, order.Status)

	_, duplicated, err = repo.MarkPaymentPaid(context.Background(), "PAY7001", "2026080422001", "raw=notify")
	require.NoError(t, err)
	require.True(t, duplicated)
}
