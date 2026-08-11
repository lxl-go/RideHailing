package data

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/passenger-service/internal/biz"
)

type passengerModel struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	Nickname          string    `gorm:"column:nickname;type:varchar(64);not null;default:''"`
	Phone             string    `gorm:"column:phone;type:varchar(32);not null;default:''"`
	AvatarURL         string    `gorm:"column:avatar_url;type:varchar(255);not null;default:''"`
	CommonAddress     string    `gorm:"column:common_address;type:varchar(255);not null;default:''"`
	PaymentPreference string    `gorm:"column:payment_preference;type:varchar(64);not null;default:''"`
	Status            int       `gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (passengerModel) TableName() string {
	return "passenger_profile"
}

type PassengerRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewPassengerRepo(db *gorm.DB, log *zap.Logger) *PassengerRepo {
	return &PassengerRepo{db: db, log: log}
}

func (r *PassengerRepo) GetByID(ctx context.Context, id int64) (*biz.PassengerProfile, error) {
	var m passengerModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPassengerNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *PassengerRepo) Create(ctx context.Context, profile *biz.PassengerProfile) error {
	return r.db.WithContext(ctx).Create(toRecord(profile)).Error
}

func (r *PassengerRepo) Update(ctx context.Context, profile *biz.PassengerProfile) error {
	result := r.db.WithContext(ctx).Model(&passengerModel{}).Where("id = ?", profile.ID).Updates(toRecord(profile))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrPassengerNotFound
	}
	return nil
}

func toDomain(m *passengerModel) *biz.PassengerProfile {
	return &biz.PassengerProfile{
		ID:                m.ID,
		Nickname:          m.Nickname,
		Phone:             m.Phone,
		AvatarURL:         m.AvatarURL,
		CommonAddress:     m.CommonAddress,
		PaymentPreference: m.PaymentPreference,
		Status:            m.Status,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func toRecord(e *biz.PassengerProfile) *passengerModel {
	return &passengerModel{
		ID:                e.ID,
		Nickname:          e.Nickname,
		Phone:             e.Phone,
		AvatarURL:         e.AvatarURL,
		CommonAddress:     e.CommonAddress,
		PaymentPreference: e.PaymentPreference,
		Status:            e.Status,
	}
}
