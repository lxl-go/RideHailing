package biz

import "context"

type TripRepo interface {
	SearchTrips(ctx context.Context, origin, destination, departDate string, page, pageSize int) ([]Trip, int64, error)
	GetByID(ctx context.Context, id int64) (*Trip, error)
	Create(ctx context.Context, trip *Trip) error
	ListByDriver(ctx context.Context, driverID int64, status int, page, pageSize int) ([]Trip, int64, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
	ListCoupons(ctx context.Context, passengerID int64, page, pageSize int) ([]Coupon, int64, error)
	ClaimCoupon(ctx context.Context, passengerID int64, couponNo string, idempotencyKey string) (*Coupon, bool, error)
	CreateDemand(ctx context.Context, demand *TripDemand) error
	ListDemandsByPassenger(ctx context.Context, passengerID int64, status int, page, pageSize int) ([]TripDemand, int64, error)
	CancelDemand(ctx context.Context, id int64, passengerID int64) error
	DeleteDriverTrip(ctx context.Context, id int64, driverID int64) error
}
