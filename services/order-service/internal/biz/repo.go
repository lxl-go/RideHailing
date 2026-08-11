package biz

import (
	"context"
	"time"
)

type OrderRepo interface {
	GetTripForOrder(ctx context.Context, id int64) (*TripSnapshot, error)
	CreateAtomic(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id int64) (*Order, error)
	ListByPassenger(ctx context.Context, passengerID int64, status int, page, pageSize int) ([]Order, int64, error)
	ListByDriver(ctx context.Context, driverID int64, status int, page, pageSize int) ([]Order, int64, error)
	ListPendingByTrip(ctx context.Context, tripID int64, page, pageSize int) ([]Order, int64, error)
	ListPendingByDriver(ctx context.Context, driverID int64, page, pageSize int) ([]Order, int64, error)
	ListDriverIncome(ctx context.Context, driverID int64, start, end time.Time, page, pageSize int) ([]DriverIncomeRecord, int64, float64, error)
	ApplyOrderTransition(ctx context.Context, transition OrderTransition) error
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*Payment, error)
	GetPaymentByOutTradeNo(ctx context.Context, outTradeNo string) (*Payment, error)
	MarkPaymentPaid(ctx context.Context, outTradeNo, alipayTradeNo string, notifyPayload string) (*Payment, bool, error)
}
