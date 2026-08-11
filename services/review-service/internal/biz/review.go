package biz

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

const (
	MinRating = 1
	MaxRating = 5
)

const OrderStatusCompleted = 2

type Review struct {
	ID         int64
	OrderID    int64
	FromUserID int64
	ToUserID   int64
	Rating     int
	Content    string
	CreatedAt  time.Time
}

type OrderSnapshot struct {
	ID          int64
	PassengerID int64
	DriverID    int64
	Status      int
}

type SubmitReviewCommand struct {
	OrderID    int64
	FromUserID int64
	ToUserID   int64
	Rating     int
	Content    string
}

type ReviewUsecase struct {
	node *snowflake.Node
	log  *zap.Logger
	repo ReviewRepo
}

func NewReviewUsecase(node *snowflake.Node, log *zap.Logger, repo ReviewRepo) *ReviewUsecase {
	return &ReviewUsecase{node: node, log: log, repo: repo}
}

func (uc *ReviewUsecase) SubmitReview(ctx context.Context, cmd SubmitReviewCommand) (*Review, error) {
	if cmd.OrderID <= 0 || cmd.FromUserID <= 0 || cmd.ToUserID <= 0 || cmd.Rating < MinRating || cmd.Rating > MaxRating {
		return nil, ErrInvalidReview
	}
	order, err := uc.repo.GetOrderForReview(ctx, cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if order.Status != OrderStatusCompleted {
		return nil, ErrOrderNotCompleted
	}
	if cmd.FromUserID != order.PassengerID && cmd.FromUserID != order.DriverID {
		return nil, ErrReviewerNotAuthorized
	}
	exists, err := uc.repo.ExistsByOrderAndUser(ctx, cmd.OrderID, cmd.FromUserID)
	if err != nil {
		uc.log.Error("check review exists failed", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, ErrReviewAlreadyExists
	}
	review := &Review{
		ID:         uc.node.Generate().Int64(),
		OrderID:    cmd.OrderID,
		FromUserID: cmd.FromUserID,
		ToUserID:   cmd.ToUserID,
		Rating:     cmd.Rating,
		Content:    strings.TrimSpace(cmd.Content),
		CreatedAt:  time.Now(),
	}
	if err := uc.repo.Create(ctx, review); err != nil {
		uc.log.Error("create review failed", zap.Error(err))
		return nil, err
	}
	return review, nil
}

func (uc *ReviewUsecase) GetMyReview(ctx context.Context, orderID, fromUserID int64) (*Review, bool, error) {
	if orderID <= 0 || fromUserID <= 0 {
		return nil, false, ErrInvalidReview
	}
	review, err := uc.repo.GetByOrderAndUser(ctx, orderID, fromUserID)
	if err != nil {
		if errors.Is(err, ErrReviewNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return review, true, nil
}
