package data

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/review-service/internal/biz"
)

type reviewModel struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	OrderID    int64     `gorm:"column:order_id;type:bigint;not null;uniqueIndex:uk_order_from_user"`
	FromUserID int64     `gorm:"column:from_user_id;type:bigint;not null;uniqueIndex:uk_order_from_user"`
	ToUserID   int64     `gorm:"column:to_user_id;type:bigint;not null;index"`
	Rating     int       `gorm:"column:rating;type:tinyint;not null"`
	Content    string    `gorm:"column:content;type:text"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (reviewModel) TableName() string {
	return "carpool_review"
}

type orderJoinModel struct {
	ID          int64 `gorm:"column:id"`
	PassengerID int64 `gorm:"column:passenger_id"`
	DriverID    int64 `gorm:"column:driver_id"`
	Status      int   `gorm:"column:status"`
}

type ReviewRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewReviewRepo(db *gorm.DB, log *zap.Logger) *ReviewRepo {
	return &ReviewRepo{db: db, log: log}
}

func (r *ReviewRepo) GetOrderForReview(ctx context.Context, orderID int64) (*biz.OrderSnapshot, error) {
	var m orderJoinModel
	err := r.db.WithContext(ctx).Table("carpool_order").
		Select("carpool_order.id, carpool_order.passenger_id, carpool_order.status, carpool_trip.driver_id").
		Joins("JOIN carpool_trip ON carpool_trip.id = carpool_order.trip_id").
		Where("carpool_order.id = ?", orderID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, err
	}
	return &biz.OrderSnapshot{ID: m.ID, PassengerID: m.PassengerID, DriverID: m.DriverID, Status: m.Status}, nil
}

func (r *ReviewRepo) ExistsByOrderAndUser(ctx context.Context, orderID, fromUserID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&reviewModel{}).Where("order_id = ? AND from_user_id = ?", orderID, fromUserID).Count(&count).Error
	return count > 0, err
}

func (r *ReviewRepo) Create(ctx context.Context, review *biz.Review) error {
	return r.db.WithContext(ctx).Create(toRecord(review)).Error
}

func (r *ReviewRepo) GetByOrderAndUser(ctx context.Context, orderID, fromUserID int64) (*biz.Review, error) {
	var m reviewModel
	err := r.db.WithContext(ctx).Where("order_id = ? AND from_user_id = ?", orderID, fromUserID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrReviewNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func toRecord(e *biz.Review) *reviewModel {
	return &reviewModel{
		ID:         e.ID,
		OrderID:    e.OrderID,
		FromUserID: e.FromUserID,
		ToUserID:   e.ToUserID,
		Rating:     e.Rating,
		Content:    e.Content,
		CreatedAt:  e.CreatedAt,
	}
}

func toDomain(m *reviewModel) *biz.Review {
	return &biz.Review{
		ID:         m.ID,
		OrderID:    m.OrderID,
		FromUserID: m.FromUserID,
		ToUserID:   m.ToUserID,
		Rating:     m.Rating,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt,
	}
}
