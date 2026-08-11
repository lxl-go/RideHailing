package biz

import "context"

type ReviewRepo interface {
	GetOrderForReview(ctx context.Context, orderID int64) (*OrderSnapshot, error)
	ExistsByOrderAndUser(ctx context.Context, orderID, fromUserID int64) (bool, error)
	Create(ctx context.Context, review *Review) error
	GetByOrderAndUser(ctx context.Context, orderID, fromUserID int64) (*Review, error)
}
