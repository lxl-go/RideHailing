package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	PassengerStatusEnabled  = 1
	PassengerStatusDisabled = 2
)

type PassengerProfile struct {
	ID                int64
	Nickname          string
	Phone             string
	AvatarURL         string
	CommonAddress     string
	PaymentPreference string
	Status            int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UpdatePassengerCommand struct {
	ID                int64
	Nickname          string
	Phone             string
	AvatarURL         string
	CommonAddress     string
	PaymentPreference string
}

type PassengerUsecase struct {
	log  *zap.Logger
	repo PassengerRepo
}

func NewPassengerUsecase(log *zap.Logger, repo PassengerRepo) *PassengerUsecase {
	return &PassengerUsecase{log: log, repo: repo}
}

func (uc *PassengerUsecase) EnsurePassenger(ctx context.Context, id int64, phone string) (*PassengerProfile, error) {
	if id <= 0 {
		return nil, ErrInvalidPassenger
	}
	phone = strings.TrimSpace(phone)
	profile, err := uc.repo.GetByID(ctx, id)
	if err == nil {
		if phone != "" && strings.TrimSpace(profile.Phone) == "" {
			profile.Phone = phone
			if err := uc.repo.Update(ctx, profile); err != nil {
				uc.log.Error("backfill passenger phone failed", zap.Error(err))
				return nil, err
			}
		}
		return profile, nil
	}
	if !errors.Is(err, ErrPassengerNotFound) {
		return nil, err
	}
	profile = &PassengerProfile{
		ID:       id,
		Nickname: fmt.Sprintf("Passenger %d", id),
		Phone:    phone,
		Status:   PassengerStatusEnabled,
	}
	if err := uc.repo.Create(ctx, profile); err != nil {
		uc.log.Error("create passenger profile failed", zap.Error(err))
		return nil, err
	}
	return profile, nil
}

func (uc *PassengerUsecase) GetPassenger(ctx context.Context, id int64) (*PassengerProfile, error) {
	if id <= 0 {
		return nil, ErrInvalidPassenger
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *PassengerUsecase) UpdatePassenger(ctx context.Context, cmd UpdatePassengerCommand) (*PassengerProfile, error) {
	if cmd.ID <= 0 {
		return nil, ErrInvalidPassenger
	}
	profile, err := uc.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	profile.Nickname = strings.TrimSpace(cmd.Nickname)
	profile.Phone = strings.TrimSpace(cmd.Phone)
	profile.AvatarURL = strings.TrimSpace(cmd.AvatarURL)
	profile.CommonAddress = strings.TrimSpace(cmd.CommonAddress)
	profile.PaymentPreference = strings.TrimSpace(cmd.PaymentPreference)
	if profile.Status == 0 {
		profile.Status = PassengerStatusEnabled
	}
	if err := uc.repo.Update(ctx, profile); err != nil {
		uc.log.Error("update passenger profile failed", zap.Error(err))
		return nil, err
	}
	return profile, nil
}
