package service

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "ride-hailing/services/order-service/api/order/v1"
	"ride-hailing/services/order-service/internal/biz"
)

type OrderService struct {
	orderv1.UnimplementedOrderServiceServer
	uc *biz.OrderUsecase
}

func NewOrderService(uc *biz.OrderUsecase) *OrderService {
	return &OrderService{uc: uc}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderReply, error) {
	order, err := s.uc.CreateOrder(ctx, biz.CreateOrderCommand{
		TripID:      req.TripId,
		PassengerID: req.PassengerId,
		SeatsBooked: int(req.SeatsBooked),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CreateOrderReply{OrderId: order.ID, TotalPrice: order.TotalPrice}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderReply, error) {
	if err := s.uc.CancelOrder(ctx, biz.OrderActionCommand{
		OrderID:        req.Id,
		ActorID:        req.PassengerId,
		IdempotencyKey: req.IdempotencyKey,
	}); err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CancelOrderReply{}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersReply, error) {
	var (
		items []biz.Order
		total int64
		err   error
	)
	if req.DriverId > 0 {
		items, total, err = s.uc.ListDriverOrders(ctx, req.DriverId, int(req.Status), int(req.Page), int(req.PageSize))
	} else {
		items, total, err = s.uc.ListOrders(ctx, req.PassengerId, int(req.Status), int(req.Page), int(req.PageSize))
	}
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.ListOrdersReply{Total: total, Items: ordersToProto(items)}, nil
}

func (s *OrderService) GetOrderDetail(ctx context.Context, req *orderv1.GetOrderDetailRequest) (*orderv1.GetOrderDetailReply, error) {
	order, trip, err := s.uc.GetOrderDetail(ctx, req.Id, req.PassengerId)
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.GetOrderDetailReply{Order: orderToProto(order, trip)}, nil
}

func (s *OrderService) PendingOrders(ctx context.Context, req *orderv1.PendingOrdersRequest) (*orderv1.PendingOrdersReply, error) {
	items, total, err := s.uc.PendingOrders(ctx, req.DriverId, req.TripId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.PendingOrdersReply{Total: total, Items: pendingOrdersToProto(items)}, nil
}

func (s *OrderService) AcceptOrder(ctx context.Context, req *orderv1.AcceptOrderRequest) (*orderv1.AcceptOrderReply, error) {
	if err := s.uc.AcceptOrder(ctx, biz.OrderActionCommand{
		OrderID:        req.Id,
		ActorID:        req.DriverId,
		IdempotencyKey: req.IdempotencyKey,
	}); err != nil {
		return nil, mapError(err)
	}
	return &orderv1.AcceptOrderReply{}, nil
}

func (s *OrderService) RejectOrder(ctx context.Context, req *orderv1.RejectOrderRequest) (*orderv1.RejectOrderReply, error) {
	if err := s.uc.RejectOrder(ctx, biz.OrderActionCommand{
		OrderID:        req.Id,
		ActorID:        req.DriverId,
		IdempotencyKey: req.IdempotencyKey,
		RejectReason:   req.RejectReason,
	}); err != nil {
		return nil, mapError(err)
	}
	return &orderv1.RejectOrderReply{}, nil
}

func (s *OrderService) StartPickup(ctx context.Context, req *orderv1.StartPickupRequest) (*orderv1.StartPickupReply, error) {
	if err := s.uc.StartPickup(ctx, biz.OrderActionCommand{
		OrderID:        req.Id,
		ActorID:        req.DriverId,
		IdempotencyKey: req.IdempotencyKey,
	}); err != nil {
		return nil, mapError(err)
	}
	return &orderv1.StartPickupReply{}, nil
}

func (s *OrderService) StartDelivery(ctx context.Context, req *orderv1.StartDeliveryRequest) (*orderv1.StartDeliveryReply, error) {
	if err := s.uc.StartDelivery(ctx, biz.OrderActionCommand{
		OrderID:        req.Id,
		ActorID:        req.DriverId,
		IdempotencyKey: req.IdempotencyKey,
	}); err != nil {
		return nil, mapError(err)
	}
	return &orderv1.StartDeliveryReply{}, nil
}

func (s *OrderService) CompleteOrder(ctx context.Context, req *orderv1.CompleteOrderRequest) (*orderv1.CompleteOrderReply, error) {
	if err := s.uc.CompleteOrder(ctx, biz.OrderActionCommand{
		OrderID:        req.Id,
		ActorID:        req.DriverId,
		IdempotencyKey: req.IdempotencyKey,
	}); err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CompleteOrderReply{}, nil
}

func (s *OrderService) CreatePayment(ctx context.Context, req *orderv1.CreatePaymentRequest) (*orderv1.CreatePaymentReply, error) {
	payment, err := s.uc.CreatePayment(ctx, biz.CreatePaymentCommand{
		OrderID:     req.OrderId,
		PassengerID: req.PassengerId,
		Channel:     req.Channel,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CreatePaymentReply{
		PaymentId:   payment.ID,
		OrderId:     payment.OrderID,
		OutTradeNo:  payment.OutTradeNo,
		TotalAmount: payment.TotalAmount,
		Status:      int32(payment.Status),
	}, nil
}

func (s *OrderService) MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error) {
	payment, duplicated, err := s.uc.MarkPaymentPaid(ctx, biz.MarkPaymentPaidCommand{
		OutTradeNo:    req.OutTradeNo,
		AlipayTradeNo: req.AlipayTradeNo,
		AppID:         req.AppId,
		TotalAmount:   req.TotalAmount,
		TradeStatus:   req.TradeStatus,
		NotifyPayload: req.NotifyPayload,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.MarkPaymentPaidReply{
		PaymentId:   payment.ID,
		OrderId:     payment.OrderID,
		OutTradeNo:  payment.OutTradeNo,
		Status:      int32(payment.Status),
		OrderStatus: int32(biz.OrderStatusPaid),
		Duplicated:  duplicated,
	}, nil
}

func (s *OrderService) GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error) {
	payment, err := s.uc.GetPaymentStatus(ctx, req.OutTradeNo, req.OrderId, req.PassengerId)
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.GetPaymentStatusReply{
		PaymentId:   payment.ID,
		OrderId:     payment.OrderID,
		OutTradeNo:  payment.OutTradeNo,
		TotalAmount: payment.TotalAmount,
		Status:      int32(payment.Status),
	}, nil
}

func (s *OrderService) GetDriverIncome(ctx context.Context, req *orderv1.DriverIncomeRequest) (*orderv1.DriverIncomeReply, error) {
	start, err := parseProtoTime(req.StartTime)
	if err != nil {
		return nil, mapError(biz.ErrInvalidOrder)
	}
	end, err := parseProtoTime(req.EndTime)
	if err != nil {
		return nil, mapError(biz.ErrInvalidOrder)
	}
	summary, err := s.uc.DriverIncome(ctx, biz.DriverIncomeQuery{
		DriverID:  req.DriverId,
		StartTime: start,
		EndTime:   end,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.DriverIncomeReply{
		TodayOrders:     summary.TodayOrders,
		TodayIncome:     summary.TodayIncome,
		PendingWithdraw: summary.PendingWithdraw,
		Records:         driverIncomeRecordsToProto(summary.Records),
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidOrder), errors.Is(err, biz.ErrTripNotAvailable), errors.Is(err, biz.ErrInsufficientSeats), errors.Is(err, biz.ErrInvalidPayment), errors.Is(err, biz.ErrPaymentAmountMismatch), errors.Is(err, biz.ErrPaymentNotSuccessful), errors.Is(err, biz.ErrRejectReasonRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrOrderNotFound), errors.Is(err, biz.ErrTripNotFound), errors.Is(err, biz.ErrPaymentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrNotOrderOwner), errors.Is(err, biz.ErrNotTripOwner), errors.Is(err, biz.ErrOrderCannotCancel), errors.Is(err, biz.ErrOrderCannotComplete), errors.Is(err, biz.ErrOrderAlreadyHandled):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func ordersToProto(items []biz.Order) []*orderv1.OrderItem {
	out := make([]*orderv1.OrderItem, len(items))
	for i := range items {
		out[i] = orderToProto(&items[i], nil)
	}
	return out
}

func pendingOrdersToProto(items []biz.Order) []*orderv1.PendingOrderItem {
	out := make([]*orderv1.PendingOrderItem, len(items))
	for i := range items {
		out[i] = &orderv1.PendingOrderItem{
			Id:          items[i].ID,
			TripId:      items[i].TripID,
			PassengerId: items[i].PassengerID,
			SeatsBooked: int32(items[i].SeatsBooked),
			TotalPrice:  items[i].TotalPrice,
			Status:      int32(items[i].Status),
			CreatedAt:   formatTime(items[i].CreatedAt),
		}
	}
	return out
}

func orderToProto(order *biz.Order, trip *biz.TripSnapshot) *orderv1.OrderItem {
	if order == nil {
		return nil
	}
	item := &orderv1.OrderItem{
		Id:           order.ID,
		TripId:       order.TripID,
		PassengerId:  order.PassengerID,
		DriverId:     order.DriverID,
		Origin:       order.Origin,
		Destination:  order.Destination,
		DepartTime:   formatTime(order.DepartTime),
		SeatsBooked:  int32(order.SeatsBooked),
		TotalPrice:   order.TotalPrice,
		Status:       int32(order.Status),
		CreatedAt:    formatTime(order.CreatedAt),
		AcceptedAt:   formatTimePtr(order.AcceptedAt),
		RejectReason: order.RejectReason,
		RejectedAt:   formatTimePtr(order.RejectedAt),
		RefundAmount: order.RefundAmount,
		RefundedAt:   formatTimePtr(order.RefundedAt),
	}
	if trip != nil {
		item.DriverId = trip.DriverID
		item.Origin = trip.Origin
		item.Destination = trip.Destination
		item.DepartTime = formatTime(trip.DepartTime)
	}
	return item
}

func driverIncomeRecordsToProto(items []biz.DriverIncomeRecord) []*orderv1.DriverIncomeRecord {
	out := make([]*orderv1.DriverIncomeRecord, len(items))
	for i := range items {
		out[i] = &orderv1.DriverIncomeRecord{
			OrderId:     items[i].OrderID,
			PassengerId: items[i].PassengerID,
			TripId:      items[i].TripID,
			Origin:      items[i].Origin,
			Destination: items[i].Destination,
			Amount:      items[i].Amount,
			Status:      int32(items[i].Status),
			AcceptedAt:  formatTime(items[i].AcceptedAt),
		}
	}
	return out
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

func parseProtoTime(value string) (time.Time, error) {
	value = biz.CleanText(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02 15:04:05", value)
}
