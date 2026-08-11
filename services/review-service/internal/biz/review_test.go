package biz

import (
	"context"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeReviewRepo struct {
	order   OrderSnapshot
	exists  bool
	reviews []Review
}

func (r *fakeReviewRepo) GetOrderForReview(context.Context, int64) (*OrderSnapshot, error) {
	return &r.order, nil
}

func (r *fakeReviewRepo) ExistsByOrderAndUser(context.Context, int64, int64) (bool, error) {
	return r.exists, nil
}

func (r *fakeReviewRepo) Create(_ context.Context, review *Review) error {
	r.reviews = append(r.reviews, *review)
	return nil
}

func (r *fakeReviewRepo) GetByOrderAndUser(_ context.Context, orderID, fromUserID int64) (*Review, error) {
	for _, review := range r.reviews {
		if review.OrderID == orderID && review.FromUserID == fromUserID {
			return &review, nil
		}
	}
	return nil, ErrReviewNotFound
}

func TestSubmitReviewAllowsPassengerOnCompletedOrder(t *testing.T) {
	node, err := snowflake.NewNode(3)
	require.NoError(t, err)
	repo := &fakeReviewRepo{order: OrderSnapshot{
		ID:          5001,
		PassengerID: 3001,
		DriverID:    2001,
		Status:      OrderStatusCompleted,
	}}
	uc := NewReviewUsecase(node, zap.NewNop(), repo)

	review, err := uc.SubmitReview(context.Background(), SubmitReviewCommand{
		OrderID:    5001,
		FromUserID: 3001,
		ToUserID:   2001,
		Rating:     5,
		Content:    "good",
	})

	require.NoError(t, err)
	require.NotZero(t, review.ID)
	require.Len(t, repo.reviews, 1)
}

func TestSubmitReviewRejectsDuplicateReview(t *testing.T) {
	node, err := snowflake.NewNode(3)
	require.NoError(t, err)
	repo := &fakeReviewRepo{
		exists: true,
		order:  OrderSnapshot{ID: 5001, PassengerID: 3001, DriverID: 2001, Status: OrderStatusCompleted},
	}
	uc := NewReviewUsecase(node, zap.NewNop(), repo)

	_, err = uc.SubmitReview(context.Background(), SubmitReviewCommand{
		OrderID:    5001,
		FromUserID: 3001,
		ToUserID:   2001,
		Rating:     5,
	})

	require.ErrorIs(t, err, ErrReviewAlreadyExists)
	require.Empty(t, repo.reviews)
}

func TestGetMyReviewReturnsExistingReviewStatus(t *testing.T) {
	node, err := snowflake.NewNode(3)
	require.NoError(t, err)
	repo := &fakeReviewRepo{reviews: []Review{{
		ID:         7001,
		OrderID:    5001,
		FromUserID: 3001,
		ToUserID:   2001,
		Rating:     5,
		Content:    "good",
	}}}
	uc := NewReviewUsecase(node, zap.NewNop(), repo)

	review, hasReview, err := uc.GetMyReview(context.Background(), 5001, 3001)

	require.NoError(t, err)
	require.True(t, hasReview)
	require.Equal(t, int64(7001), review.ID)
	require.Equal(t, 5, review.Rating)
}
