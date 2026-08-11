package biz

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeOrderRepo struct {
	orders       []Order
	trips        map[int64]TripSnapshot
	payments     []Payment
	idempotency  map[string]struct{}
	incrementHit bool
	listDriverID int64
	listStatus   int
}

func (r *fakeOrderRepo) GetTripForOrder(_ context.Context, id int64) (*TripSnapshot, error) {
	trip, ok := r.trips[id]
	if !ok {
		return nil, ErrTripNotFound
	}
	return &trip, nil
}

func (r *fakeOrderRepo) CreateAtomic(_ context.Context, order *Order) error {
	r.orders = append(r.orders, *order)
	return nil
}

func (r *fakeOrderRepo) GetByID(_ context.Context, id int64) (*Order, error) {
	for i := range r.orders {
		if r.orders[i].ID == id {
			return &r.orders[i], nil
		}
	}
	return nil, ErrOrderNotFound
}

func (r *fakeOrderRepo) ListByPassenger(context.Context, int64, int, int, int) ([]Order, int64, error) {
	return r.orders, int64(len(r.orders)), nil
}

func (r *fakeOrderRepo) ListByDriver(_ context.Context, driverID int64, status int, page, pageSize int) ([]Order, int64, error) {
	r.listDriverID = driverID
	r.listStatus = status
	return r.orders, int64(len(r.orders)), nil
}

func (r *fakeOrderRepo) ListPendingByTrip(context.Context, int64, int, int) ([]Order, int64, error) {
	return r.orders, int64(len(r.orders)), nil
}

func (r *fakeOrderRepo) ListPendingByDriver(context.Context, int64, int, int) ([]Order, int64, error) {
	return r.orders, int64(len(r.orders)), nil
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, id int64, status int) error {
	for i := range r.orders {
		if r.orders[i].ID == id {
			r.orders[i].Status = status
			return nil
		}
	}
	return ErrOrderNotFound
}

func (r *fakeOrderRepo) IncrementTripSeats(_ context.Context, tripID int64, seats int) error {
	trip := r.trips[tripID]
	trip.SeatsAvailable += seats
	r.trips[tripID] = trip
	r.incrementHit = true
	return nil
}

func (r *fakeOrderRepo) ApplyOrderTransition(_ context.Context, transition OrderTransition) error {
	if r.idempotency == nil {
		r.idempotency = map[string]struct{}{}
	}
	bizKey := transition.Action + ":" + transition.IdempotencyKey
	if transition.IdempotencyKey != "" {
		if _, ok := r.idempotency[bizKey]; ok {
			return nil
		}
	}
	for i := range r.orders {
		if r.orders[i].ID != transition.OrderID {
			continue
		}
		if !statusIn(r.orders[i].Status, transition.FromStatuses) {
			return transitionStatusError(transition.Action)
		}
		r.orders[i].Status = transition.ToStatus
		if transition.AcceptedAt != nil {
			r.orders[i].AcceptedAt = transition.AcceptedAt
		}
		if transition.RejectReason != "" {
			r.orders[i].RejectReason = transition.RejectReason
		}
		if transition.RejectedAt != nil {
			r.orders[i].RejectedAt = transition.RejectedAt
		}
		if transition.RefundAmount > 0 {
			r.orders[i].RefundAmount = transition.RefundAmount
		}
		if transition.RefundedAt != nil {
			r.orders[i].RefundedAt = transition.RefundedAt
		}
		if transition.RefundPayment {
			for j := range r.payments {
				if r.payments[j].OrderID == transition.OrderID && r.payments[j].Status == PaymentStatusPaid {
					r.payments[j].Status = PaymentStatusRefunded
				}
			}
		}
		if transition.RestoreTripID > 0 && transition.RestoreSeats > 0 {
			trip := r.trips[transition.RestoreTripID]
			trip.SeatsAvailable += transition.RestoreSeats
			r.trips[transition.RestoreTripID] = trip
			r.incrementHit = true
		}
		if transition.IdempotencyKey != "" {
			r.idempotency[bizKey] = struct{}{}
		}
		return nil
	}
	return ErrOrderNotFound
}

func (r *fakeOrderRepo) ListDriverIncome(_ context.Context, driverID int64, start, end time.Time, page, pageSize int) ([]DriverIncomeRecord, int64, float64, error) {
	var records []DriverIncomeRecord
	var totalAmount float64
	for _, order := range r.orders {
		trip := r.trips[order.TripID]
		if trip.DriverID != driverID || order.AcceptedAt == nil {
			continue
		}
		if order.AcceptedAt.Before(start) || !order.AcceptedAt.Before(end) || order.Status == OrderStatusCancelled {
			continue
		}
		totalAmount += order.TotalPrice
		records = append(records, DriverIncomeRecord{
			OrderID:     order.ID,
			PassengerID: order.PassengerID,
			TripID:      order.TripID,
			Origin:      trip.Origin,
			Destination: trip.Destination,
			Amount:      order.TotalPrice,
			Status:      order.Status,
			AcceptedAt:  *order.AcceptedAt,
		})
	}
	total := int64(len(records))
	startIndex := (page - 1) * pageSize
	if startIndex >= len(records) {
		return nil, total, totalAmount, nil
	}
	endIndex := startIndex + pageSize
	if endIndex > len(records) {
		endIndex = len(records)
	}
	return records[startIndex:endIndex], total, totalAmount, nil
}

func (r *fakeOrderRepo) CreatePayment(_ context.Context, payment *Payment) error {
	r.payments = append(r.payments, *payment)
	return nil
}

func (r *fakeOrderRepo) GetPaymentByOrderID(_ context.Context, orderID int64) (*Payment, error) {
	for i := range r.payments {
		if r.payments[i].OrderID == orderID {
			return &r.payments[i], nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (r *fakeOrderRepo) GetPaymentByOutTradeNo(_ context.Context, outTradeNo string) (*Payment, error) {
	for i := range r.payments {
		if r.payments[i].OutTradeNo == outTradeNo {
			return &r.payments[i], nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (r *fakeOrderRepo) MarkPaymentPaid(_ context.Context, outTradeNo, alipayTradeNo string, notifyPayload string) (*Payment, bool, error) {
	for i := range r.payments {
		if r.payments[i].OutTradeNo != outTradeNo {
			continue
		}
		if r.payments[i].Status == PaymentStatusPaid {
			return &r.payments[i], true, nil
		}
		r.payments[i].Status = PaymentStatusPaid
		r.payments[i].AlipayTradeNo = alipayTradeNo
		r.payments[i].NotifyPayload = notifyPayload
		now := time.Now()
		r.payments[i].PaidAt = &now
		for j := range r.orders {
			if r.orders[j].ID == r.payments[i].OrderID {
				r.orders[j].Status = OrderStatusPaid
			}
		}
		return &r.payments[i], false, nil
	}
	return nil, false, ErrPaymentNotFound
}

func TestCreateOrderCalculatesTotalPriceAndPendingStatus(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{trips: map[int64]TripSnapshot{
		1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 3, Price: 19.9, Status: TripStatusRecruiting},
	}}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	order, err := uc.CreateOrder(context.Background(), CreateOrderCommand{
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 2,
	})

	require.NoError(t, err)
	require.NotZero(t, order.ID)
	require.Equal(t, OrderStatusPending, order.Status)
	require.Equal(t, 39.8, order.TotalPrice)
	require.Len(t, repo.orders, 1)
}

func TestCreateOrderAcceptsApprovedTripStatusFromTripService(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{trips: map[int64]TripSnapshot{
		1001: {ID: 1001, DriverID: 2001, DepartTime: time.Now().Add(time.Hour), SeatsAvailable: 3, Price: 19.9, Status: 20},
	}}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	order, err := uc.CreateOrder(context.Background(), CreateOrderCommand{
		TripID:      1001,
		PassengerID: 3001,
		SeatsBooked: 1,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1001), order.TripID)
	require.Len(t, repo.orders, 1)
}

func TestCreatePaymentCreatesAlipaySandboxPaymentForOwner(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 39.8, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  39.8,
			Status:      OrderStatusPending,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	payment, err := uc.CreatePayment(context.Background(), CreatePaymentCommand{
		OrderID:     5001,
		PassengerID: 3001,
		Channel:     PaymentChannelAlipaySandbox,
	})

	require.NoError(t, err)
	require.NotZero(t, payment.ID)
	require.Equal(t, int64(5001), payment.OrderID)
	require.Equal(t, int64(3001), payment.PassengerID)
	require.Equal(t, "39.80", payment.TotalAmount)
	require.Equal(t, PaymentStatusPending, payment.Status)
	require.Contains(t, payment.OutTradeNo, "PAY")
	require.Len(t, repo.payments, 1)
}

func TestMarkPaymentPaidIsIdempotentAndMarksOrderPaid(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 39.8, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  39.8,
			Status:      OrderStatusPending,
		}},
		payments: []Payment{{
			ID:          7001,
			OrderID:     5001,
			PassengerID: 3001,
			OutTradeNo:  "PAY7001",
			Channel:     PaymentChannelAlipaySandbox,
			TotalAmount: "39.80",
			Status:      PaymentStatusPending,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	payment, duplicated, err := uc.MarkPaymentPaid(context.Background(), MarkPaymentPaidCommand{
		OutTradeNo:    "PAY7001",
		AlipayTradeNo: "2026080422001",
		TotalAmount:   "39.80",
		TradeStatus:   "TRADE_SUCCESS",
		NotifyPayload: "raw=notify",
	})
	require.NoError(t, err)
	require.False(t, duplicated)
	require.Equal(t, PaymentStatusPaid, payment.Status)
	require.Equal(t, OrderStatusPaid, repo.orders[0].Status)

	_, duplicated, err = uc.MarkPaymentPaid(context.Background(), MarkPaymentPaidCommand{
		OutTradeNo:    "PAY7001",
		AlipayTradeNo: "2026080422001",
		TotalAmount:   "39.80",
		TradeStatus:   "TRADE_SUCCESS",
	})
	require.NoError(t, err)
	require.True(t, duplicated)
	require.Equal(t, OrderStatusPaid, repo.orders[0].Status)
}

func TestRejectOrderRestoresTripSeats(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 2,
			TotalPrice:  39.8,
			Status:      OrderStatusPaid,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	err = uc.RejectOrder(context.Background(), OrderActionCommand{
		OrderID:        5001,
		ActorID:        2001,
		IdempotencyKey: "reject-5001",
		RejectReason:   "车辆临时故障",
	})

	require.NoError(t, err)
	require.Equal(t, OrderStatusCancelled, repo.orders[0].Status)
	require.True(t, repo.incrementHit)
	require.Equal(t, 3, repo.trips[1001].SeatsAvailable)
}

func TestRejectOrderRequiresReasonAndRefundsPaidPayment(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 39.8, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  39.8,
			Status:      OrderStatusPaid,
		}},
		payments: []Payment{{
			ID:          7001,
			OrderID:     5001,
			PassengerID: 3001,
			OutTradeNo:  "PAY7001",
			Channel:     PaymentChannelAlipaySandbox,
			TotalAmount: "39.80",
			Status:      PaymentStatusPaid,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	err = uc.RejectOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "reject-missing-reason"})
	require.ErrorIs(t, err, ErrRejectReasonRequired)

	err = uc.RejectOrder(context.Background(), OrderActionCommand{
		OrderID:        5001,
		ActorID:        2001,
		IdempotencyKey: "reject-5001",
		RejectReason:   "车辆临时故障，无法接单",
	})
	require.NoError(t, err)
	require.Equal(t, OrderStatusCancelled, repo.orders[0].Status)
	require.Equal(t, "车辆临时故障，无法接单", repo.orders[0].RejectReason)
	require.NotNil(t, repo.orders[0].RejectedAt)
	require.NotNil(t, repo.orders[0].RefundedAt)
	require.Equal(t, 39.8, repo.orders[0].RefundAmount)
	require.Equal(t, PaymentStatusRefunded, repo.payments[0].Status)
}

func TestDriverLifecycleRequiresPaidThenPickupDeliveryComplete(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusPaid,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	require.NoError(t, uc.AcceptOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "accept-5001"}))
	require.Equal(t, OrderStatusAccepted, repo.orders[0].Status)
	require.NoError(t, uc.StartPickup(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "pickup-5001"}))
	require.Equal(t, OrderStatusPickingUp, repo.orders[0].Status)
	require.NoError(t, uc.StartDelivery(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "delivery-5001"}))
	require.Equal(t, OrderStatusDelivering, repo.orders[0].Status)
	require.NoError(t, uc.CompleteOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "complete-5001"}))
	require.Equal(t, OrderStatusCompleted, repo.orders[0].Status)
}

func TestDriverLifecycleRejectsWrongExpectedState(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusAccepted,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	err = uc.CompleteOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "complete-too-early"})

	require.ErrorIs(t, err, ErrOrderCannotComplete)
	require.Equal(t, OrderStatusAccepted, repo.orders[0].Status)
}

func TestOrderActionIdempotencyReturnsSuccessForRepeatedKey(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusPaid,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)
	cmd := OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "accept-repeat-5001"}

	require.NoError(t, uc.AcceptOrder(context.Background(), cmd))
	require.NoError(t, uc.AcceptOrder(context.Background(), cmd))
	require.Equal(t, OrderStatusAccepted, repo.orders[0].Status)
}

func TestAcceptOrderRecordsAcceptedAtAndDriverIncome(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, Origin: "A", Destination: "B", SeatsAvailable: 1, Price: 50, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 2,
			TotalPrice:  100,
			Status:      OrderStatusPaid,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	require.NoError(t, uc.AcceptOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2001, IdempotencyKey: "accept-5001"}))
	require.NotNil(t, repo.orders[0].AcceptedAt)

	start := repo.orders[0].AcceptedAt.Add(-time.Minute)
	end := repo.orders[0].AcceptedAt.Add(time.Minute)
	summary, err := uc.DriverIncome(context.Background(), DriverIncomeQuery{DriverID: 2001, StartTime: start, EndTime: end, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.EqualValues(t, 1, summary.TodayOrders)
	require.Equal(t, 100.0, summary.TodayIncome)
	require.Equal(t, 100.0, summary.PendingWithdraw)
	require.Len(t, summary.Records, 1)
	require.Equal(t, int64(5001), summary.Records[0].OrderID)
}

func TestPassengerCannotAccessAnotherPassengerOrder(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusPending,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	_, _, err = uc.GetOrderDetail(context.Background(), 5001, 3999)
	require.ErrorIs(t, err, ErrNotOrderOwner)

	err = uc.CancelOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 3999, IdempotencyKey: "cancel-not-owner"})
	require.ErrorIs(t, err, ErrNotOrderOwner)
	require.Equal(t, OrderStatusPending, repo.orders[0].Status)
}

func TestDriverCanAccessOwnTripOrderDetail(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusPending,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	order, trip, err := uc.GetOrderDetail(context.Background(), 5001, 2001)

	require.NoError(t, err)
	require.Equal(t, int64(5001), order.ID)
	require.Equal(t, int64(2001), trip.DriverID)
}

func TestDriverCannotHandleAnotherDriverTripOrder(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		trips: map[int64]TripSnapshot{
			1001: {ID: 1001, DriverID: 2001, SeatsAvailable: 1, Price: 19.9, Status: TripStatusRecruiting},
		},
		orders: []Order{{
			ID:          5001,
			TripID:      1001,
			PassengerID: 3001,
			SeatsBooked: 1,
			TotalPrice:  19.9,
			Status:      OrderStatusPending,
		}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	err = uc.AcceptOrder(context.Background(), OrderActionCommand{OrderID: 5001, ActorID: 2999, IdempotencyKey: "accept-not-owner"})
	require.ErrorIs(t, err, ErrNotTripOwner)
	require.Equal(t, OrderStatusPending, repo.orders[0].Status)

	_, _, err = uc.PendingOrders(context.Background(), 2999, 1001, 1, 20)
	require.ErrorIs(t, err, ErrNotTripOwner)
}

func TestListDriverOrdersRequiresDriverAndForwardsStatus(t *testing.T) {
	node, err := snowflake.NewNode(2)
	require.NoError(t, err)
	repo := &fakeOrderRepo{
		orders: []Order{{ID: 5001, TripID: 1001, PassengerID: 3001, Status: OrderStatusAccepted}},
	}
	uc := NewOrderUsecase(node, zap.NewNop(), repo)

	orders, total, err := uc.ListDriverOrders(context.Background(), 2001, OrderStatusPickingUp, 1, 20)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, int64(5001), orders[0].ID)
	require.Equal(t, int64(2001), repo.listDriverID)
	require.Equal(t, OrderStatusPickingUp, repo.listStatus)

	_, _, err = uc.ListDriverOrders(context.Background(), 0, -1, 1, 20)
	require.ErrorIs(t, err, ErrInvalidOrder)
}
