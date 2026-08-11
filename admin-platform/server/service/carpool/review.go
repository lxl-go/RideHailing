package carpool

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/model/carpool"
	req "ride-hailing/admin-server/model/carpool/request"
)

type ReviewService struct{}

var ReviewServiceApp = new(ReviewService)
var auditNotifier = new(AuditNotifier)

const (
	auditStatusPending       = 0
	auditStatusApproved      = 1
	auditStatusRejected      = 2
	auditStatusDriverDeleted = 3

	driverCertStatusPending  = 2
	driverCertStatusApproved = 3
	driverCertStatusRejected = 4

	driverVehicleStatusActive = 1
)

func (s *ReviewService) ListAudits(ctx context.Context, search req.CertReviewListSearch) (list []carpool.CertificationAudit, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpool.CertificationAudit{})
	if search.Status != nil {
		db = db.Where("status = ?", *search.Status)
	}
	if search.UserID > 0 {
		db = db.Where("user_id = ?", search.UserID)
	}
	if search.CertType != nil {
		db = db.Where("cert_type = ?", *search.CertType)
	}
	if search.Keyword != "" {
		db = db.Where("real_name LIKE ? OR cert_number LIKE ?", "%"+search.Keyword+"%", "%"+search.Keyword+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	limit, offset := search.LimitOffset()
	err = db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&list).Error
	return
}

func (s *ReviewService) GetAudit(ctx context.Context, id int64) (*carpool.CertificationAudit, error) {
	var audit carpool.CertificationAudit
	err := global.GVA_DB.WithContext(ctx).First(&audit, id).Error
	if err != nil {
		return nil, err
	}
	return &audit, nil
}

func (s *ReviewService) ApproveAudit(ctx context.Context, id int64, reviewerID int64) error {
	var audit *carpool.CertificationAudit
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockedAudit, err := lockPendingCertificationAudit(tx, id)
		if err != nil {
			return err
		}
		now := time.Now()
		updates := map[string]interface{}{
			"status":                auditStatusApproved,
			"reviewer_id":           reviewerID,
			"reviewed_at":           &now,
			"review_duration_hours": reviewDurationHours(lockedAudit.CreatedAt, now),
		}
		result := tx.Model(&carpool.CertificationAudit{}).
			Where("id = ? AND status = ?", id, auditStatusPending).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := syncDriverCertificationReviewResult(tx, lockedAudit, driverCertStatusApproved, ""); err != nil {
			return err
		}
		if err := createCertificationAuditMessage(tx, lockedAudit, "approved", auditStatusApproved, ""); err != nil {
			return err
		}
		lockedAudit.Status = auditStatusApproved
		lockedAudit.ReviewerID = reviewerID
		lockedAudit.ReviewedAt = &now
		lockedAudit.ReviewDurationHours = updates["review_duration_hours"].(int)
		audit = lockedAudit
		return nil
	})
	if err != nil {
		return err
	}
	auditNotifier.NotifyCertificationAuditResult(ctx, audit, "approved")
	return nil
}

func (s *ReviewService) RejectAudit(ctx context.Context, id int64, reviewerID int64, reason string) error {
	if reason == "" {
		return errors.New("reject reason is required")
	}
	var audit *carpool.CertificationAudit
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockedAudit, err := lockPendingCertificationAudit(tx, id)
		if err != nil {
			return err
		}
		now := time.Now()
		updates := map[string]interface{}{
			"status":                auditStatusRejected,
			"reviewer_id":           reviewerID,
			"reject_reason":         reason,
			"reviewed_at":           &now,
			"review_duration_hours": reviewDurationHours(lockedAudit.CreatedAt, now),
		}
		result := tx.Model(&carpool.CertificationAudit{}).
			Where("id = ? AND status = ?", id, auditStatusPending).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := syncDriverCertificationReviewResult(tx, lockedAudit, driverCertStatusRejected, reason); err != nil {
			return err
		}
		if err := createCertificationAuditMessage(tx, lockedAudit, "rejected", auditStatusRejected, reason); err != nil {
			return err
		}
		lockedAudit.Status = auditStatusRejected
		lockedAudit.ReviewerID = reviewerID
		lockedAudit.RejectReason = reason
		lockedAudit.ReviewedAt = &now
		lockedAudit.ReviewDurationHours = updates["review_duration_hours"].(int)
		audit = lockedAudit
		return nil
	})
	if err != nil {
		return err
	}
	auditNotifier.NotifyCertificationAuditResult(ctx, audit, "rejected")
	return nil
}

func lockPendingCertificationAudit(tx *gorm.DB, id int64) (*carpool.CertificationAudit, error) {
	var audit carpool.CertificationAudit
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&audit, id).Error; err != nil {
		return nil, err
	}
	if audit.Status != auditStatusPending {
		return nil, gorm.ErrRecordNotFound
	}
	return &audit, nil
}

func syncDriverCertificationReviewResult(tx *gorm.DB, audit *carpool.CertificationAudit, status int, reason string) error {
	if audit == nil || audit.UserRole != 2 {
		return nil
	}
	var cert carpool.DriverCertification
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("driver_id = ?", audit.UserID).
		First(&cert).Error; err != nil {
		return err
	}
	certUpdates := map[string]interface{}{
		"status":        status,
		"reject_reason": reason,
	}
	if err := tx.Model(&carpool.DriverCertification{}).
		Where("driver_id = ?", audit.UserID).
		Updates(certUpdates).Error; err != nil {
		return err
	}

	var profile carpool.DriverProfile
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&profile, audit.UserID).Error; err != nil {
		return err
	}
	return tx.Model(&carpool.DriverProfile{}).
		Where("id = ?", audit.UserID).
		Update("certification_status", status).Error
}

func createCertificationAuditMessage(tx *gorm.DB, audit *carpool.CertificationAudit, result string, status int, reason string) error {
	if audit == nil || audit.UserRole != 2 {
		return nil
	}
	payload, err := json.Marshal(map[string]interface{}{
		"type":         "certification_audit",
		"result":       result,
		"auditId":      audit.ID,
		"rejectReason": reason,
		"status":       status,
		"title":        certificationMessageTitle(result),
	})
	if err != nil {
		return err
	}
	return tx.Create(&carpool.RealtimeMessage{
		Topic:    "certification.audit",
		UserID:   uint64(audit.UserID),
		UserRole: "driver",
		Payload:  string(payload),
	}).Error
}

func certificationMessageTitle(result string) string {
	return certificationMessageTitleCN(result)
}

func certificationMessageTitleCN(result string) string {
	if result == "rejected" {
		return "司机认证已驳回，请修改后重新提交"
	}
	return "司机认证已通过"
}

func certificationMessageTitleLegacy(result string) string {
	if result == "rejected" {
		return "司机认证已驳回，请修改后重新提交"
	}
	return "司机认证已通过"
}

func (s *ReviewService) ListVehicleReviews(ctx context.Context, search req.VehicleReviewListSearch) (list []carpool.VehicleInfo, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpool.VehicleInfo{})
	if search.Status != nil {
		db = db.Where("status = ?", *search.Status)
	}
	if search.DriverID > 0 {
		db = db.Where("driver_id = ?", search.DriverID)
	}
	if search.Keyword != "" {
		db = db.Where("plate_number LIKE ? OR brand LIKE ?", "%"+search.Keyword+"%", "%"+search.Keyword+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	limit, offset := search.LimitOffset()
	err = db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&list).Error
	return
}

func (s *ReviewService) HandleVehicleReview(ctx context.Context, id int64, reviewerID int64, status int, reason string) error {
	reason = strings.TrimSpace(reason)
	if status != auditStatusApproved && status != auditStatusRejected {
		return errors.New("invalid vehicle review status")
	}
	if status == auditStatusRejected && reason == "" {
		return errors.New("reject reason is required")
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockedInfo, err := lockPendingVehicleInfo(tx, id)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{
			"status":        status,
			"reviewer_id":   reviewerID,
			"reject_reason": "",
		}
		if status == auditStatusRejected {
			updates["reject_reason"] = reason
		}
		result := tx.Model(&carpool.VehicleInfo{}).
			Where("id = ? AND status = ?", id, auditStatusPending).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if status == auditStatusApproved {
			if err := syncApprovedDriverVehicle(tx, lockedInfo); err != nil {
				return err
			}
			return createVehicleAuditMessage(tx, lockedInfo, "approved", status, "")
		}
		return createVehicleAuditMessage(tx, lockedInfo, "rejected", status, reason)
	})
}

func lockPendingVehicleInfo(tx *gorm.DB, id int64) (*carpool.VehicleInfo, error) {
	var info carpool.VehicleInfo
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&info, id).Error; err != nil {
		return nil, err
	}
	if info.Status != auditStatusPending {
		return nil, gorm.ErrRecordNotFound
	}
	return &info, nil
}

func syncApprovedDriverVehicle(tx *gorm.DB, info *carpool.VehicleInfo) error {
	if info == nil {
		return nil
	}
	updates := map[string]interface{}{
		"driver_id":    info.DriverID,
		"plate_no":     info.PlateNumber,
		"brand":        info.Brand,
		"model":        info.Model,
		"vehicle_type": info.Model,
		"color":        info.Color,
		"seats":        normalizeVehicleSeats(info.Seats),
		"status":       driverVehicleStatusActive,
	}
	var existing carpool.DriverVehicle
	lock := tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := lock.Where("id = ? AND driver_id = ?", info.ID, info.DriverID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = lock.Where("driver_id = ? AND plate_no = ?", info.DriverID, info.PlateNumber).First(&existing).Error
	}
	if err == nil {
		return tx.Model(&carpool.DriverVehicle{}).
			Where("id = ?", existing.ID).
			Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&carpool.DriverVehicle{
		ID:          info.ID,
		DriverID:    info.DriverID,
		PlateNo:     info.PlateNumber,
		Brand:       info.Brand,
		Model:       info.Model,
		VehicleType: info.Model,
		Color:       info.Color,
		Seats:       normalizeVehicleSeats(info.Seats),
		Status:      driverVehicleStatusActive,
	}).Error
}

func normalizeVehicleSeats(seats int) int {
	if seats < 1 || seats > 9 {
		return 4
	}
	return seats
}

func createVehicleAuditMessage(tx *gorm.DB, info *carpool.VehicleInfo, result string, status int, reason string) error {
	if info == nil {
		return nil
	}
	payload, err := json.Marshal(map[string]interface{}{
		"type":         "vehicle_audit",
		"result":       result,
		"auditId":      info.ID,
		"plateNumber":  info.PlateNumber,
		"rejectReason": reason,
		"status":       status,
		"title":        vehicleMessageTitle(result),
	})
	if err != nil {
		return err
	}
	return tx.Create(&carpool.RealtimeMessage{
		Topic:    "vehicle.audit",
		UserID:   uint64(info.DriverID),
		UserRole: "driver",
		Payload:  string(payload),
	}).Error
}

func vehicleMessageTitle(result string) string {
	return vehicleMessageTitleCN(result)
}

func vehicleMessageTitleCN(result string) string {
	if result == "rejected" {
		return "车辆认证已驳回，请修改后重新提交"
	}
	return "车辆认证已通过"
}

func vehicleMessageTitleLegacy(result string) string {
	if result == "rejected" {
		return "车辆认证已驳回，请修改后重新提交"
	}
	return "车辆认证已通过"
}

func reviewDurationHours(start time.Time, end time.Time) int {
	if start.IsZero() || end.Before(start) {
		return 0
	}
	hours := int(end.Sub(start).Hours())
	if hours == 0 {
		return 1
	}
	return hours
}
