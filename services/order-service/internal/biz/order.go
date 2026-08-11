package biz

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"

	"ride-hailing/pkg/authz"
)

const (
	OrderStatusPending = iota
	OrderStatusConfirmed
	OrderStatusCompleted
	OrderStatusCancelled
	OrderStatusPaid
	OrderStatusPickingUp
	OrderStatusDelivering
)

const OrderStatusAccepted = OrderStatusConfirmed

const (
	TripStatusApproved   = 20
	TripStatusRecruiting = TripStatusApproved
)

const (
	OrderActionAccept        = "accept"
	OrderActionReject        = "reject"
	OrderActionCancel        = "cancel"
	OrderActionStartPickup   = "start_pickup"
	OrderActionStartDelivery = "start_delivery"
	OrderActionComplete      = "complete"
)

const (
	PaymentStatusPending = iota
	PaymentStatusPaid
	PaymentStatusRefunded
)

const PaymentChannelAlipaySandbox = "alipay_sandbox"

type Order struct {
	ID           int64
	TripID       int64
	PassengerID  int64
	DriverID     int64
	Origin       string
	Destination  string
	DepartTime   time.Time
	SeatsBooked  int
	TotalPrice   float64
	Status       int
	AcceptedAt   *time.Time
	RejectReason string
	RejectedAt   *time.Time
	RefundAmount float64
	RefundedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Payment struct {
	ID            int64
	OrderID       int64
	PassengerID   int64
	OutTradeNo    string
	AlipayTradeNo string
	Channel       string
	TotalAmount   string
	Status        int
	PaidAt        *time.Time
	NotifyPayload string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TripSnapshot struct {
	ID             int64
	DriverID       int64
	Origin         string
	Destination    string
	DepartTime     time.Time
	SeatsAvailable int
	Price          float64
	Status         int
}

type CreateOrderCommand struct {
	TripID      int64
	PassengerID int64
	SeatsBooked int
}

type CreatePaymentCommand struct {
	OrderID     int64
	PassengerID int64
	Channel     string
}

type MarkPaymentPaidCommand struct {
	OutTradeNo    string
	AlipayTradeNo string
	AppID         string
	TotalAmount   string
	TradeStatus   string
	NotifyPayload string
}

type OrderActionCommand struct {
	OrderID        int64
	ActorID        int64
	IdempotencyKey string
	RejectReason   string
}

type OrderTransition struct {
	OrderID        int64
	ActorID        int64
	Action         string
	IdempotencyKey string
	FromStatuses   []int
	ToStatus       int
	RestoreTripID  int64
	RestoreSeats   int
	AcceptedAt     *time.Time
	RejectReason   string
	RejectedAt     *time.Time
	RefundAmount   float64
	RefundedAt     *time.Time
	RefundPayment  bool
}

type DriverIncomeQuery struct {
	DriverID  int64
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
}

type DriverIncomeRecord struct {
	OrderID     int64
	PassengerID int64
	TripID      int64
	Origin      string
	Destination string
	Amount      float64
	Status      int
	AcceptedAt  time.Time
}

type DriverIncomeSummary struct {
	TodayOrders     int64
	TodayIncome     float64
	PendingWithdraw float64
	Records         []DriverIncomeRecord
}

type OrderUsecase struct {
	node *snowflake.Node
	log  *zap.Logger
	repo OrderRepo
}

func NewOrderUsecase(node *snowflake.Node, log *zap.Logger, repo OrderRepo) *OrderUsecase {
	return &OrderUsecase{node: node, log: log, repo: repo}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*Order, error) {
	if cmd.TripID <= 0 || cmd.PassengerID <= 0 || cmd.SeatsBooked <= 0 {
		return nil, ErrInvalidOrder
	}
	trip, err := uc.repo.GetTripForOrder(ctx, cmd.TripID)
	if err != nil {
		return nil, err
	}
	if trip.Status != TripStatusRecruiting {
		return nil, ErrTripNotAvailable
	}
	if !trip.DepartTime.IsZero() && !trip.DepartTime.After(time.Now()) {
		return nil, ErrTripNotAvailable
	}
	if trip.SeatsAvailable < cmd.SeatsBooked {
		return nil, ErrInsufficientSeats
	}
	order := &Order{
		ID:          uc.node.Generate().Int64(),
		TripID:      cmd.TripID,
		PassengerID: cmd.PassengerID,
		SeatsBooked: cmd.SeatsBooked,
		TotalPrice:  trip.Price * float64(cmd.SeatsBooked),
		Status:      OrderStatusPending,
	}
	if err := uc.repo.CreateAtomic(ctx, order); err != nil {
		uc.log.Error("create order failed", zap.Error(err))
		return nil, err
	}
	return order, nil
}

func (uc *OrderUsecase) CreatePayment(ctx context.Context, cmd CreatePaymentCommand) (*Payment, error) {
	if cmd.OrderID <= 0 || cmd.PassengerID <= 0 {
		return nil, ErrInvalidPayment
	}
	channel := strings.TrimSpace(cmd.Channel)
	if channel == "" {
		channel = PaymentChannelAlipaySandbox
	}
	order, err := uc.repo.GetByID(ctx, cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if err := authz.RequireOwner(cmd.PassengerID, order.PassengerID, "order"); err != nil {
		return nil, ErrNotOrderOwner
	}
	if existing, err := uc.repo.GetPaymentByOrderID(ctx, cmd.OrderID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrPaymentNotFound) {
		return nil, err
	}
	paymentID := uc.node.Generate().Int64()
	payment := &Payment{
		ID:          paymentID,
		OrderID:     order.ID,
		PassengerID: order.PassengerID,
		OutTradeNo:  "PAY" + strconv.FormatInt(paymentID, 10),
		Channel:     channel,
		TotalAmount: formatPaymentAmount(order.TotalPrice),
		Status:      PaymentStatusPending,
	}
	if err := uc.repo.CreatePayment(ctx, payment); err != nil {
		uc.log.Error("create payment failed", zap.Int64("order_id", cmd.OrderID), zap.Error(err))
		return nil, err
	}
	return payment, nil
}

func (uc *OrderUsecase) MarkPaymentPaid(ctx context.Context, cmd MarkPaymentPaidCommand) (*Payment, bool, error) {
	cmd.OutTradeNo = strings.TrimSpace(cmd.OutTradeNo)
	cmd.AlipayTradeNo = strings.TrimSpace(cmd.AlipayTradeNo)
	cmd.TotalAmount = normalizePaymentAmount(cmd.TotalAmount)
	cmd.TradeStatus = strings.TrimSpace(cmd.TradeStatus)
	if cmd.OutTradeNo == "" || cmd.AlipayTradeNo == "" || cmd.TotalAmount == "" {
		return nil, false, ErrInvalidPayment
	}
	if cmd.TradeStatus != "TRADE_SUCCESS" && cmd.TradeStatus != "TRADE_FINISHED" {
		return nil, false, ErrPaymentNotSuccessful
	}
	payment, err := uc.repo.GetPaymentByOutTradeNo(ctx, cmd.OutTradeNo)
	if err != nil {
		return nil, false, err
	}
	if normalizePaymentAmount(payment.TotalAmount) != cmd.TotalAmount {
		return nil, false, ErrPaymentAmountMismatch
	}
	updated, duplicated, err := uc.repo.MarkPaymentPaid(ctx, cmd.OutTradeNo, cmd.AlipayTradeNo, cmd.NotifyPayload)
	if err != nil {
		return nil, false, err
	}
	uc.log.Info("payment marked paid",
		zap.Int64("order_id", updated.OrderID),
		zap.String("out_trade_no", updated.OutTradeNo),
		zap.Bool("duplicated", duplicated),
	)
	return updated, duplicated, nil
}

func (uc *OrderUsecase) GetPaymentStatus(ctx context.Context, outTradeNo string, orderID, passengerID int64) (*Payment, error) {
	outTradeNo = strings.TrimSpace(outTradeNo)
	var payment *Payment
	var err error
	if outTradeNo != "" {
		payment, err = uc.repo.GetPaymentByOutTradeNo(ctx, outTradeNo)
	} else if orderID > 0 {
		payment, err = uc.repo.GetPaymentByOrderID(ctx, orderID)
	} else {
		return nil, ErrInvalidPayment
	}
	if err != nil {
		return nil, err
	}
	if passengerID > 0 && payment.PassengerID != passengerID {
		return nil, ErrNotOrderOwner
	}
	return payment, nil
}

func (uc *OrderUsecase) CancelOrder(ctx context.Context, cmd OrderActionCommand) error {
	if cmd.OrderID <= 0 || cmd.ActorID <= 0 {
		return ErrInvalidOrder
	}
	order, err := uc.repo.GetByID(ctx, cmd.OrderID)
	if err != nil {
		return err
	}
	if err := authz.RequireOwner(cmd.ActorID, order.PassengerID, "order"); err != nil {
		return ErrNotOrderOwner
	}
	if err := uc.repo.ApplyOrderTransition(ctx, OrderTransition{
		OrderID:        cmd.OrderID,
		ActorID:        cmd.ActorID,
		Action:         OrderActionCancel,
		IdempotencyKey: normalizeActionKey(cmd.IdempotencyKey),
		FromStatuses:   []int{OrderStatusPending, OrderStatusPaid},
		ToStatus:       OrderStatusCancelled,
		RestoreTripID:  order.TripID,
		RestoreSeats:   order.SeatsBooked,
	}); err != nil {
		return err
	}
	uc.log.Info("order cancelled",
		zap.Int64("order_id", cmd.OrderID),
		zap.Int64("passenger_id", cmd.ActorID),
		zap.String("idempotency_key", normalizeActionKey(cmd.IdempotencyKey)),
	)
	return nil
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, passengerID int64, status int, page, pageSize int) ([]Order, int64, error) {
	if passengerID <= 0 {
		return nil, 0, ErrInvalidOrder
	}
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.ListByPassenger(ctx, passengerID, status, page, pageSize)
}

func (uc *OrderUsecase) ListDriverOrders(ctx context.Context, driverID int64, status int, page, pageSize int) ([]Order, int64, error) {
	if driverID <= 0 {
		return nil, 0, ErrInvalidOrder
	}
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.ListByDriver(ctx, driverID, status, page, pageSize)
}

func (uc *OrderUsecase) GetOrderDetail(ctx context.Context, id, passengerID int64) (*Order, *TripSnapshot, error) {
	if id <= 0 || passengerID <= 0 {
		return nil, nil, ErrInvalidOrder
	}
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	trip, _ := uc.repo.GetTripForOrder(ctx, order.TripID)
	if err := authz.RequireOwner(passengerID, order.PassengerID, "order"); err != nil {
		if trip == nil || authz.RequireOwner(passengerID, trip.DriverID, "trip") != nil {
			return nil, nil, ErrNotOrderOwner
		}
	}
	return order, trip, nil
}

func (uc *OrderUsecase) PendingOrders(ctx context.Context, driverID, tripID int64, page, pageSize int) ([]Order, int64, error) {
	if driverID <= 0 {
		return nil, 0, ErrInvalidOrder
	}
	page, pageSize = normalizePage(page, pageSize)
	if tripID > 0 {
		trip, err := uc.repo.GetTripForOrder(ctx, tripID)
		if err != nil {
			return nil, 0, err
		}
		if err := authz.RequireOwner(driverID, trip.DriverID, "trip"); err != nil {
			return nil, 0, ErrNotTripOwner
		}
		return uc.repo.ListPendingByTrip(ctx, tripID, page, pageSize)
	}
	return uc.repo.ListPendingByDriver(ctx, driverID, page, pageSize)
}

func (uc *OrderUsecase) AcceptOrder(ctx context.Context, cmd OrderActionCommand) error {
	now := time.Now()
	return uc.handleDriverOrderTransition(ctx, cmd, OrderActionTransitionOptions{
		Action:       OrderActionAccept,
		FromStatuses: []int{OrderStatusPaid},
		ToStatus:     OrderStatusAccepted,
		AcceptedAt:   &now,
	})
}

func (uc *OrderUsecase) RejectOrder(ctx context.Context, cmd OrderActionCommand) error {
	reason := strings.TrimSpace(cmd.RejectReason)
	if reason == "" {
		return ErrRejectReasonRequired
	}
	now := time.Now()
	return uc.handleDriverOrderTransition(ctx, cmd, OrderActionTransitionOptions{
		Action:        OrderActionReject,
		FromStatuses:  []int{OrderStatusPaid},
		ToStatus:      OrderStatusCancelled,
		RestoreSeats:  true,
		RejectReason:  reason,
		RejectedAt:    &now,
		RefundedAt:    &now,
		RefundPayment: true,
	})
}

func (uc *OrderUsecase) StartPickup(ctx context.Context, cmd OrderActionCommand) error {
	return uc.handleDriverOrderTransition(ctx, cmd, OrderActionTransitionOptions{
		Action:       OrderActionStartPickup,
		FromStatuses: []int{OrderStatusAccepted},
		ToStatus:     OrderStatusPickingUp,
	})
}

func (uc *OrderUsecase) StartDelivery(ctx context.Context, cmd OrderActionCommand) error {
	return uc.handleDriverOrderTransition(ctx, cmd, OrderActionTransitionOptions{
		Action:       OrderActionStartDelivery,
		FromStatuses: []int{OrderStatusPickingUp},
		ToStatus:     OrderStatusDelivering,
	})
}

func (uc *OrderUsecase) CompleteOrder(ctx context.Context, cmd OrderActionCommand) error {
	return uc.handleDriverOrderTransition(ctx, cmd, OrderActionTransitionOptions{
		Action:       OrderActionComplete,
		FromStatuses: []int{OrderStatusDelivering},
		ToStatus:     OrderStatusCompleted,
	})
}

type OrderActionTransitionOptions struct {
	Action        string
	FromStatuses  []int
	ToStatus      int
	RestoreSeats  bool
	AcceptedAt    *time.Time
	RejectReason  string
	RejectedAt    *time.Time
	RefundedAt    *time.Time
	RefundPayment bool
}

func (uc *OrderUsecase) handleDriverOrderTransition(ctx context.Context, cmd OrderActionCommand, opts OrderActionTransitionOptions) error {
	if cmd.OrderID <= 0 || cmd.ActorID <= 0 {
		return ErrInvalidOrder
	}
	order, err := uc.repo.GetByID(ctx, cmd.OrderID)
	if err != nil {
		return err
	}
	trip, err := uc.repo.GetTripForOrder(ctx, order.TripID)
	if err != nil {
		return err
	}
	if err := authz.RequireOwner(cmd.ActorID, trip.DriverID, "trip"); err != nil {
		return ErrNotTripOwner
	}
	transition := OrderTransition{
		OrderID:        cmd.OrderID,
		ActorID:        cmd.ActorID,
		Action:         opts.Action,
		IdempotencyKey: normalizeActionKey(cmd.IdempotencyKey),
		FromStatuses:   opts.FromStatuses,
		ToStatus:       opts.ToStatus,
		AcceptedAt:     opts.AcceptedAt,
		RejectReason:   opts.RejectReason,
		RejectedAt:     opts.RejectedAt,
		RefundedAt:     opts.RefundedAt,
		RefundPayment:  opts.RefundPayment,
	}
	if opts.RestoreSeats {
		transition.RestoreTripID = order.TripID
		transition.RestoreSeats = order.SeatsBooked
		transition.RefundAmount = order.TotalPrice
	}
	if err := uc.repo.ApplyOrderTransition(ctx, transition); err != nil {
		return err
	}
	uc.log.Info("order transition applied",
		zap.Int64("order_id", cmd.OrderID),
		zap.Int64("driver_id", cmd.ActorID),
		zap.String("action", opts.Action),
		zap.Int("to_status", opts.ToStatus),
		zap.String("idempotency_key", transition.IdempotencyKey),
	)
	return nil
}

func (uc *OrderUsecase) DriverIncome(ctx context.Context, query DriverIncomeQuery) (*DriverIncomeSummary, error) {
	if query.DriverID <= 0 {
		return nil, ErrInvalidOrder
	}
	page, pageSize := normalizePage(query.Page, query.PageSize)
	start, end := normalizeIncomeRange(query.StartTime, query.EndTime)
	records, total, totalAmount, err := uc.repo.ListDriverIncome(ctx, query.DriverID, start, end, page, pageSize)
	if err != nil {
		return nil, err
	}
	totalAmount = roundMoney(totalAmount)
	return &DriverIncomeSummary{
		TodayOrders:     total,
		TodayIncome:     totalAmount,
		PendingWithdraw: totalAmount,
		Records:         records,
	}, nil
}

func (o *Order) canCancel() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusPaid
}

func normalizeActionKey(value string) string {
	return strings.TrimSpace(value)
}

func statusIn(status int, expected []int) bool {
	for _, item := range expected {
		if status == item {
			return true
		}
	}
	return false
}

func TransitionStatusError(action string) error {
	switch action {
	case OrderActionCancel:
		return ErrOrderCannotCancel
	case OrderActionComplete:
		return ErrOrderCannotComplete
	default:
		return ErrOrderAlreadyHandled
	}
}

func transitionStatusError(action string) error {
	return TransitionStatusError(action)
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func normalizeIncomeRange(start, end time.Time) (time.Time, time.Time) {
	if start.IsZero() || end.IsZero() || !end.After(start) {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 0, 1)
	}
	return start, end
}

func CleanText(value string) string {
	return strings.TrimSpace(value)
}

func formatPaymentAmount(value float64) string {
	return strconv.FormatFloat(roundMoney(value), 'f', 2, 64)
}

func roundMoney(value float64) float64 {
	cents := int64(math.Round(value * 100))
	return float64(cents) / 100
}

func normalizePaymentAmount(value string) string {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return ""
	}
	return formatPaymentAmount(amount)
}
