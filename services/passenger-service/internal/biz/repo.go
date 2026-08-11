package biz

import "context"

type PassengerRepo interface {
	GetByID(ctx context.Context, id int64) (*PassengerProfile, error)
	Create(ctx context.Context, profile *PassengerProfile) error
	Update(ctx context.Context, profile *PassengerProfile) error
}
