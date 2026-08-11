package biz

import (
	"context"

	"ride-hailing/services/gateway-service/internal/data"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type OrderUsecase struct {
	client data.OrderClient
}

func NewOrderUsecase(client data.OrderClient) *OrderUsecase {
	return &OrderUsecase{client: client}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, tripID, passengerID int64, seatsBooked int32) (*orderv1.CreateOrderReply, error) {
	return uc.client.CreateOrder(ctx, tripID, passengerID, seatsBooked)
}

func (uc *OrderUsecase) CancelOrder(ctx context.Context, id, passengerID int64, idempotencyKey string) error {
	return uc.client.CancelOrder(ctx, id, passengerID, idempotencyKey)
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, passengerID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	return uc.client.ListOrders(ctx, passengerID, status, page, pageSize)
}

func (uc *OrderUsecase) ListDriverOrders(ctx context.Context, driverID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	return uc.client.ListDriverOrders(ctx, driverID, status, page, pageSize)
}

func (uc *OrderUsecase) GetOrderDetail(ctx context.Context, id, passengerID int64) (*orderv1.GetOrderDetailReply, error) {
	return uc.client.GetOrderDetail(ctx, id, passengerID)
}

func (uc *OrderUsecase) PendingOrders(ctx context.Context, driverID, tripID int64, page, pageSize int32) (*orderv1.PendingOrdersReply, error) {
	return uc.client.PendingOrders(ctx, driverID, tripID, page, pageSize)
}

func (uc *OrderUsecase) AcceptOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return uc.client.AcceptOrder(ctx, id, driverID, idempotencyKey)
}

func (uc *OrderUsecase) RejectOrder(ctx context.Context, id, driverID int64, idempotencyKey, rejectReason string) error {
	return uc.client.RejectOrder(ctx, id, driverID, idempotencyKey, rejectReason)
}

func (uc *OrderUsecase) StartPickup(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return uc.client.StartPickup(ctx, id, driverID, idempotencyKey)
}

func (uc *OrderUsecase) StartDelivery(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return uc.client.StartDelivery(ctx, id, driverID, idempotencyKey)
}

func (uc *OrderUsecase) CompleteOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	return uc.client.CompleteOrder(ctx, id, driverID, idempotencyKey)
}

func (uc *OrderUsecase) GetDriverIncome(ctx context.Context, driverID int64, startTime, endTime string, page, pageSize int32) (*orderv1.DriverIncomeReply, error) {
	return uc.client.GetDriverIncome(ctx, driverID, startTime, endTime, page, pageSize)
}

func (uc *OrderUsecase) CreatePayment(ctx context.Context, orderID, passengerID int64, channel string) (*orderv1.CreatePaymentReply, error) {
	return uc.client.CreatePayment(ctx, orderID, passengerID, channel)
}

func (uc *OrderUsecase) MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error) {
	return uc.client.MarkPaymentPaid(ctx, req)
}

func (uc *OrderUsecase) GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error) {
	return uc.client.GetPaymentStatus(ctx, req)
}
