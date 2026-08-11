package service

import (
	"context"

	"ride-hailing/services/gateway-service/internal/biz"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type OrderService struct {
	uc *biz.OrderUsecase
}

func NewOrderService(uc *biz.OrderUsecase) *OrderService {
	return &OrderService{uc: uc}
}

func (s *OrderService) CreateOrder(ctx context.Context, tripID, passengerID int64, seatsBooked int32) (*orderv1.CreateOrderReply, error) {
	return s.uc.CreateOrder(ctx, tripID, passengerID, seatsBooked)
}

func (s *OrderService) CancelOrder(ctx context.Context, id, passengerID int64, idempotencyKey string) error {
	return s.uc.CancelOrder(ctx, id, passengerID, idempotencyKey)
}

func (s *OrderService) ListOrders(ctx context.Context, passengerID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	return s.uc.ListOrders(ctx, passengerID, status, page, pageSize)
}

func (s *OrderService) ListDriverOrders(ctx context.Context, driverID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	return s.uc.ListDriverOrders(ctx, driverID, status, page, pageSize)
}

func (s *OrderService) GetOrderDetail(ctx context.Context, id, passengerID int64) (*orderv1.GetOrderDetailReply, error) {
	return s.uc.GetOrderDetail(ctx, id, passengerID)
}

func (s *OrderService) PendingOrders(ctx context.Context, driverID, tripID int64, page, pageSize int32) (*orderv1.PendingOrdersReply, error) {
	return s.uc.PendingOrders(ctx, driverID, tripID, page, pageSize)
}

func (s *OrderService) AcceptOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return s.uc.AcceptOrder(ctx, id, driverID, idempotencyKey)
}

func (s *OrderService) RejectOrder(ctx context.Context, id, driverID int64, idempotencyKey, rejectReason string) error {
	return s.uc.RejectOrder(ctx, id, driverID, idempotencyKey, rejectReason)
}

func (s *OrderService) StartPickup(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return s.uc.StartPickup(ctx, id, driverID, idempotencyKey)
}

func (s *OrderService) StartDelivery(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return s.uc.StartDelivery(ctx, id, driverID, idempotencyKey)
}

func (s *OrderService) CompleteOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return s.uc.CompleteOrder(ctx, id, driverID, idempotencyKey)
}

func (s *OrderService) GetDriverIncome(ctx context.Context, driverID int64, startTime, endTime string, page, pageSize int32) (*orderv1.DriverIncomeReply, error) {
	return s.uc.GetDriverIncome(ctx, driverID, startTime, endTime, page, pageSize)
}

func (s *OrderService) CreatePayment(ctx context.Context, orderID, passengerID int64, channel string) (*orderv1.CreatePaymentReply, error) {
	return s.uc.CreatePayment(ctx, orderID, passengerID, channel)
}

func (s *OrderService) MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error) {
	return s.uc.MarkPaymentPaid(ctx, req)
}

func (s *OrderService) GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error) {
	return s.uc.GetPaymentStatus(ctx, req)
}
