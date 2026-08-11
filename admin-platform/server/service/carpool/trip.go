package carpool

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/model/carpool"
	req "ride-hailing/admin-server/model/carpool/request"
)

type TripService struct{}

var (
	ErrTripDeleted        = errors.New("行程已被司机删除或取消，无法继续审核")
	ErrTripHasActiveOrder = errors.New("已通过行程已有接单，不能删除")
)

const (
	tripStatusPending       = 10
	tripStatusApproved      = 20
	tripStatusRejected      = 30
	tripStatusCancelled     = 5
	tripOrderStatusCanceled = 2
	tripOrderStatusRefunded = 3
)

func (s *TripService) ListTrips(ctx context.Context, search req.TripListSearch) (list []carpool.Trip, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpool.Trip{}).Where("is_deleted = ?", false)
	if search.Status != nil {
		db = db.Where("status = ?", *search.Status)
	}
	if search.PublisherID > 0 {
		db = db.Where("publisher_id = ?", search.PublisherID)
	}
	if search.Keyword != "" {
		db = db.Where("origin_name LIKE ? OR dest_name LIKE ?", "%"+search.Keyword+"%", "%"+search.Keyword+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	limit, offset := search.LimitOffset()
	err = db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&list).Error
	return
}

func (s *TripService) GetTrip(ctx context.Context, id int64) (*carpool.Trip, error) {
	var trip carpool.Trip
	err := global.GVA_DB.WithContext(ctx).First(&trip, id).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (s *TripService) DeactivateTrip(ctx context.Context, id int64, reason string) error {
	reason = strings.TrimSpace(reason)
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trip carpool.Trip
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&trip).Error; err != nil {
			return err
		}
		if trip.IsDeleted || trip.Status == tripStatusCancelled {
			return nil
		}
		if trip.Status == tripStatusApproved {
			active, err := hasActiveOrders(tx, id)
			if err != nil {
				return err
			}
			if active {
				return ErrTripHasActiveOrder
			}
		}
		return tx.Model(&carpool.Trip{}).Where("id = ? AND is_deleted = ?", id, false).Updates(map[string]interface{}{
			"is_deleted": true,
			"status":     tripStatusCancelled,
			"remarks":    reason,
			"updated_at": time.Now(),
		}).Error
	})
}

func (s *TripService) ReviewTrip(ctx context.Context, id int64, reviewerID int64, approved bool, reason string) error {
	if reviewerID <= 0 {
		return errors.New("审核人身份无效")
	}
	if !approved {
		reason = strings.TrimSpace(reason)
		if len([]rune(reason)) < 5 || len([]rune(reason)) > 200 {
			return errors.New("驳回原因长度必须为5-200字")
		}
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trip carpool.Trip
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&trip).Error; err != nil {
			return err
		}
		if trip.IsDeleted || trip.Status == tripStatusCancelled {
			return ErrTripDeleted
		}
		if trip.Status != tripStatusPending {
			return errors.New("当前行程不是待审核状态")
		}
		now := time.Now()
		updates := map[string]interface{}{"status": tripStatusApproved, "audit_operator_id": reviewerID, "audit_time": now, "reject_reason": ""}
		if !approved {
			updates["status"] = tripStatusRejected
			updates["reject_reason"] = reason
		}
		result := tx.Model(&carpool.Trip{}).Where("id = ? AND status = ? AND is_deleted = ?", id, tripStatusPending, false).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("当前行程不是待审核状态")
		}
		return nil
	})
}

func hasActiveOrders(tx *gorm.DB, tripID int64) (bool, error) {
	var active int64
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Table("carpool_order").
		Where("trip_id = ? AND status NOT IN ?", tripID, []int{tripOrderStatusCanceled, tripOrderStatusRefunded}).
		Count(&active).Error
	return active > 0, err
}
