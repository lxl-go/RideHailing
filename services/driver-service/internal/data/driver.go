package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ride-hailing/services/driver-service/internal/biz"
)

type driverProfileModel struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	Name                string    `gorm:"column:name;type:varchar(64);not null;default:''"`
	Phone               string    `gorm:"column:phone;type:varchar(32);not null;default:''"`
	AvatarURL           string    `gorm:"column:avatar_url;type:varchar(255);not null;default:''"`
	ServiceStatus       int       `gorm:"column:service_status;type:tinyint;not null;default:1"`
	CertificationStatus int       `gorm:"column:certification_status;type:tinyint;not null;default:1"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (driverProfileModel) TableName() string {
	return "driver_profile"
}

type driverCertificationModel struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID         int64     `gorm:"column:driver_id;type:bigint;not null;uniqueIndex:uk_driver_certification"`
	RealName         string    `gorm:"column:real_name;type:varchar(64);not null;default:''"`
	IDCardNo         string    `gorm:"column:id_card_no;type:varchar(32);not null;default:''"`
	LicenseNo        string    `gorm:"column:license_no;type:varchar(64);not null;default:''"`
	LicenseType      string    `gorm:"column:license_type;type:varchar(32);not null;default:''"`
	City             string    `gorm:"column:city;type:varchar(64);not null;default:''"`
	VehicleLicenseNo string    `gorm:"column:vehicle_license_no;type:varchar(64);not null;default:''"`
	VehiclePhotoURL  string    `gorm:"column:vehicle_photo_url;type:varchar(255);not null;default:''"`
	FacePhotoURL     string    `gorm:"column:face_photo_url;type:varchar(255);not null;default:''"`
	Status           int       `gorm:"column:status;type:tinyint;not null;default:2"`
	RejectReason     string    `gorm:"column:reject_reason;type:varchar(255);not null;default:''"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (driverCertificationModel) TableName() string {
	return "driver_certification"
}

type certificationAuditModel struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	UserID           int64      `gorm:"column:user_id;type:bigint;not null;index:idx_user_role"`
	UserRole         int        `gorm:"column:user_role;type:tinyint;not null;default:2;index:idx_user_role"`
	CertType         int        `gorm:"column:cert_type;type:tinyint;not null;index:idx_cert_status"`
	CertNumber       string     `gorm:"column:cert_number;type:varchar(256);not null;default:''"`
	RealName         string     `gorm:"column:real_name;type:varchar(64);not null;default:''"`
	DriverLicenseNo  string     `gorm:"column:driver_license_no;type:varchar(64);not null;default:''"`
	LicenseType      string     `gorm:"column:license_type;type:varchar(32);not null;default:''"`
	City             string     `gorm:"column:city;type:varchar(64);not null;default:''"`
	FrontImageURL    string     `gorm:"column:front_image_url;type:varchar(512);not null;default:''"`
	BackImageURL     string     `gorm:"column:back_image_url;type:varchar(512);not null;default:''"`
	HandheldImageURL string     `gorm:"column:handheld_image_url;type:varchar(512);not null;default:''"`
	Status           int        `gorm:"column:status;type:tinyint;not null;default:0;index:idx_cert_status"`
	ReviewerID       int64      `gorm:"column:reviewer_id;type:bigint;not null;default:0"`
	RejectReason     string     `gorm:"column:reject_reason;type:varchar(512);not null;default:''"`
	SubmitCount      int        `gorm:"column:submit_count;type:int;not null;default:1"`
	ReviewedAt       *time.Time `gorm:"column:reviewed_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (certificationAuditModel) TableName() string {
	return "carpool_certification_audit"
}

type driverVehicleModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID    int64     `gorm:"column:driver_id;type:bigint;not null;index:idx_driver_vehicle"`
	PlateNo     string    `gorm:"column:plate_no;type:varchar(32);not null;default:''"`
	Brand       string    `gorm:"column:brand;type:varchar(64);not null;default:''"`
	Model       string    `gorm:"column:model;type:varchar(64);not null;default:''"`
	Color       string    `gorm:"column:color;type:varchar(32);not null;default:''"`
	VehicleType string    `gorm:"column:vehicle_type;type:varchar(32);not null;default:''"`
	Seats       int       `gorm:"column:seats;type:int;not null;default:4"`
	Status      int       `gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (driverVehicleModel) TableName() string {
	return "driver_vehicle"
}

type vehicleAuditModel struct {
	ID              int64          `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID        int64          `gorm:"column:driver_id;type:bigint;not null;index:idx_driver_status"`
	PlateNumber     string         `gorm:"column:plate_number;type:varchar(32);not null;uniqueIndex:uk_plate_number"`
	Brand           string         `gorm:"column:brand;type:varchar(64);not null;default:''"`
	Model           string         `gorm:"column:model;type:varchar(64);not null;default:''"`
	Color           string         `gorm:"column:color;type:varchar(32);not null;default:''"`
	Seats           int            `gorm:"column:seats;type:int;not null;default:4"`
	YearCheckDate   *time.Time     `gorm:"column:year_check_date"`
	InsuranceExpire *time.Time     `gorm:"column:insurance_expire"`
	Status          int            `gorm:"column:status;type:tinyint;not null;default:0;index:idx_driver_status"`
	ReviewerID      int64          `gorm:"column:reviewer_id;type:bigint;not null;default:0"`
	RejectReason    string         `gorm:"column:reject_reason;type:varchar(512);not null;default:''"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (vehicleAuditModel) TableName() string {
	return "carpool_vehicle_info"
}

type driverLocationPointModel struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID   int64     `gorm:"column:driver_id;type:bigint;not null;index:idx_driver_location_driver_time,priority:1;index:idx_driver_location_order_time,priority:1"`
	OrderID    int64     `gorm:"column:order_id;type:bigint;not null;default:0;index:idx_driver_location_order_time,priority:2"`
	Latitude   float64   `gorm:"column:latitude;type:decimal(10,6);not null"`
	Longitude  float64   `gorm:"column:longitude;type:decimal(10,6);not null"`
	Speed      float64   `gorm:"column:speed;type:decimal(8,2);not null;default:0"`
	Heading    float64   `gorm:"column:heading;type:decimal(8,2);not null;default:0"`
	ReportedAt time.Time `gorm:"column:reported_at;not null;index:idx_driver_location_driver_time,priority:2;index:idx_driver_location_order_time,priority:3"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (driverLocationPointModel) TableName() string {
	return "driver_location_point"
}

type realtimeMessageModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Topic     string    `gorm:"column:topic;type:varchar(64);not null;index:idx_realtime_message_topic"`
	UserID    uint64    `gorm:"column:user_id;not null;default:0;index:idx_realtime_message_user"`
	UserRole  string    `gorm:"column:user_role;type:varchar(32);not null;default:'';index:idx_realtime_message_role"`
	Payload   string    `gorm:"column:payload;type:text"`
	Delivered bool      `gorm:"column:delivered;type:tinyint;not null;default:0;index:idx_realtime_message_delivered"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

func (realtimeMessageModel) TableName() string {
	return "realtime_message"
}

type DriverRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewDriverRepo(db *gorm.DB, log *zap.Logger) *DriverRepo {
	return &DriverRepo{db: db, log: log}
}

func (r *DriverRepo) GetProfileByID(ctx context.Context, id int64) (*biz.DriverProfile, error) {
	var m driverProfileModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrDriverNotFound
		}
		return nil, err
	}
	return profileToDomain(&m), nil
}

func (r *DriverRepo) CreateProfile(ctx context.Context, profile *biz.DriverProfile) error {
	return r.db.WithContext(ctx).Create(profileToRecord(profile)).Error
}

func (r *DriverRepo) UpdateProfile(ctx context.Context, profile *biz.DriverProfile) error {
	result := r.db.WithContext(ctx).Model(&driverProfileModel{}).Where("id = ?", profile.ID).Updates(profileToRecord(profile))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrDriverNotFound
	}
	return nil
}

func (r *DriverRepo) SaveCertification(ctx context.Context, cert *biz.DriverCertification) error {
	existing := driverCertificationModel{}
	err := r.db.WithContext(ctx).Where("driver_id = ?", cert.DriverID).First(&existing).Error
	if err == nil {
		cert.ID = existing.ID
		return r.db.WithContext(ctx).Model(&driverCertificationModel{}).Where("driver_id = ?", cert.DriverID).Updates(certToRecord(cert)).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.WithContext(ctx).Create(certToRecord(cert)).Error
}

func (r *DriverRepo) SaveCertificationAudit(ctx context.Context, audit *biz.CertificationAudit) error {
	existing := certificationAuditModel{}
	err := r.db.WithContext(ctx).Where("user_id = ? AND user_role = ? AND cert_type = ? AND status = ?", audit.UserID, audit.UserRole, audit.CertType, 0).First(&existing).Error
	if err == nil {
		audit.ID = existing.ID
		if audit.SubmitCount <= existing.SubmitCount {
			audit.SubmitCount = existing.SubmitCount + 1
		}
		return r.db.WithContext(ctx).Model(&certificationAuditModel{}).Where("id = ?", existing.ID).Updates(auditToRecord(audit)).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if audit.SubmitCount <= 0 {
		audit.SubmitCount = 1
	}
	return r.db.WithContext(ctx).Create(auditToRecord(audit)).Error
}

func (r *DriverRepo) GetCertification(ctx context.Context, driverID int64) (*biz.DriverCertification, error) {
	var m driverCertificationModel
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrCertificationNotFound
		}
		return nil, err
	}
	return certToDomain(&m), nil
}

func (r *DriverRepo) SaveVehicle(ctx context.Context, vehicle *biz.DriverVehicle) error {
	return r.db.WithContext(ctx).Create(vehicleToRecord(vehicle)).Error
}

func (r *DriverRepo) SaveVehicleAudit(ctx context.Context, audit *biz.DriverVehicleAudit) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingByID vehicleAuditModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", audit.ID).
			First(&existingByID).Error
		if err == nil {
			if existingByID.DriverID != audit.DriverID {
				return biz.ErrVehiclePlateInUse
			}
			if err := ensureVehicleAuditPlateAvailable(tx, audit.PlateNumber, audit.DriverID, audit.ID); err != nil {
				return err
			}
			return tx.Model(&vehicleAuditModel{}).
				Where("id = ?", audit.ID).
				Updates(vehicleAuditPendingUpdates(audit)).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var existing vehicleAuditModel
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("plate_number = ?", audit.PlateNumber).
			First(&existing).Error
		if err == nil {
			if existing.DriverID != audit.DriverID {
				return biz.ErrVehiclePlateInUse
			}
			audit.ID = existing.ID
			return tx.Model(&vehicleAuditModel{}).
				Where("id = ?", existing.ID).
				Updates(vehicleAuditPendingUpdates(audit)).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Create(vehicleAuditToRecord(audit)).Error
	})
}

func ensureVehicleAuditPlateAvailable(tx *gorm.DB, plateNo string, driverID int64, auditID int64) error {
	var existing vehicleAuditModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("plate_number = ? AND id <> ?", plateNo, auditID).
		First(&existing).Error
	if err == nil && existing.DriverID != driverID {
		return biz.ErrVehiclePlateInUse
	}
	if err == nil {
		return biz.ErrVehiclePlateInUse
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func vehicleAuditPendingUpdates(audit *biz.DriverVehicleAudit) map[string]interface{} {
	return map[string]interface{}{
		"plate_number":  audit.PlateNumber,
		"brand":         audit.Brand,
		"model":         audit.Model,
		"color":         audit.Color,
		"seats":         audit.Seats,
		"status":        biz.VehicleAuditStatusPending,
		"reviewer_id":   int64(0),
		"reject_reason": "",
	}
}

func (r *DriverRepo) MarkVehicleAuditDriverDeleted(ctx context.Context, driverID int64, plateNo string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var audit vehicleAuditModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("driver_id = ? AND plate_number = ?", driverID, plateNo).
			Order("updated_at DESC, created_at DESC").
			First(&audit).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return tx.Model(&vehicleAuditModel{}).
			Where("id = ?", audit.ID).
			Updates(map[string]interface{}{
				"status":        biz.VehicleAuditStatusDriverDeleted,
				"reject_reason": "",
			}).Error
	})
}

func (r *DriverRepo) GetVehicleByID(ctx context.Context, id int64) (*biz.DriverVehicle, error) {
	var m driverVehicleModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrVehicleNotFound
		}
		return nil, err
	}
	return vehicleToDomain(&m), nil
}

func (r *DriverRepo) UpdateVehicle(ctx context.Context, vehicle *biz.DriverVehicle) error {
	result := r.db.WithContext(ctx).Model(&driverVehicleModel{}).Where("id = ?", vehicle.ID).Updates(vehicleToRecord(vehicle))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrVehicleNotFound
	}
	return nil
}

func (r *DriverRepo) ListVehicles(ctx context.Context, driverID int64) ([]biz.DriverVehicle, error) {
	var models []driverVehicleModel
	if err := r.db.WithContext(ctx).Where("driver_id = ? AND status = ?", driverID, biz.VehicleStatusActive).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]biz.DriverVehicle, len(models))
	for i := range models {
		items[i] = *vehicleToDomain(&models[i])
	}
	return items, nil
}

func (r *DriverRepo) ListVehicleAudits(ctx context.Context, driverID int64) ([]biz.DriverVehicleAudit, error) {
	var models []vehicleAuditModel
	if err := r.db.WithContext(ctx).
		Where("driver_id = ? AND status IN ?", driverID, []int{biz.VehicleAuditStatusPending, biz.VehicleAuditStatusRejected}).
		Order("updated_at DESC, created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]biz.DriverVehicleAudit, len(models))
	for i := range models {
		items[i] = *vehicleAuditToDomain(&models[i])
	}
	return items, nil
}

func (r *DriverRepo) ListDriverMessages(ctx context.Context, driverID int64) ([]biz.DriverMessage, error) {
	var models []realtimeMessageModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_role = ? AND delivered = ?", driverID, "driver", false).
		Where("topic IN ?", []string{"vehicle.audit", "certification.audit"}).
		Order("created_at DESC, id DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]biz.DriverMessage, len(models))
	for i := range models {
		items[i] = *messageToDomain(&models[i])
	}
	return items, nil
}

func (r *DriverRepo) AckDriverMessage(ctx context.Context, driverID int64, messageID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var msg realtimeMessageModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND user_role = ?", messageID, driverID, "driver").
			First(&msg).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if msg.Delivered {
			return nil
		}
		return tx.Model(&realtimeMessageModel{}).
			Where("id = ? AND user_id = ? AND user_role = ?", messageID, driverID, "driver").
			Update("delivered", true).Error
	})
}

func (r *DriverRepo) SaveLocation(ctx context.Context, point *biz.DriverLocationPoint) error {
	return r.db.WithContext(ctx).Create(locationToRecord(point)).Error
}

func (r *DriverRepo) ReplayTrack(ctx context.Context, query biz.TrackReplayQuery) ([]biz.DriverLocationPoint, int64, error) {
	db := r.db.WithContext(ctx).Model(&driverLocationPointModel{}).Where("driver_id = ?", query.DriverID)
	if query.OrderID > 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 200
	}
	var models []driverLocationPointModel
	if err := db.Order("reported_at ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	items := make([]biz.DriverLocationPoint, len(models))
	for i := range models {
		items[i] = *locationToDomain(&models[i])
	}
	return items, total, nil
}

func profileToDomain(m *driverProfileModel) *biz.DriverProfile {
	return &biz.DriverProfile{
		ID:                  m.ID,
		Name:                m.Name,
		Phone:               m.Phone,
		AvatarURL:           m.AvatarURL,
		ServiceStatus:       m.ServiceStatus,
		CertificationStatus: m.CertificationStatus,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func profileToRecord(e *biz.DriverProfile) *driverProfileModel {
	return &driverProfileModel{
		ID:                  e.ID,
		Name:                e.Name,
		Phone:               e.Phone,
		AvatarURL:           e.AvatarURL,
		ServiceStatus:       e.ServiceStatus,
		CertificationStatus: e.CertificationStatus,
	}
}

func certToDomain(m *driverCertificationModel) *biz.DriverCertification {
	return &biz.DriverCertification{
		ID:               m.ID,
		DriverID:         m.DriverID,
		RealName:         m.RealName,
		IDCardNo:         m.IDCardNo,
		LicenseNo:        m.LicenseNo,
		LicenseType:      m.LicenseType,
		City:             m.City,
		VehicleLicenseNo: m.VehicleLicenseNo,
		VehiclePhotoURL:  m.VehiclePhotoURL,
		FacePhotoURL:     m.FacePhotoURL,
		Status:           m.Status,
		RejectReason:     m.RejectReason,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func certToRecord(e *biz.DriverCertification) *driverCertificationModel {
	return &driverCertificationModel{
		ID:               e.ID,
		DriverID:         e.DriverID,
		RealName:         e.RealName,
		IDCardNo:         e.IDCardNo,
		LicenseNo:        e.LicenseNo,
		LicenseType:      e.LicenseType,
		City:             e.City,
		VehicleLicenseNo: e.VehicleLicenseNo,
		VehiclePhotoURL:  e.VehiclePhotoURL,
		FacePhotoURL:     e.FacePhotoURL,
		Status:           e.Status,
		RejectReason:     e.RejectReason,
	}
}

func auditToRecord(e *biz.CertificationAudit) *certificationAuditModel {
	return &certificationAuditModel{
		ID:               e.ID,
		UserID:           e.UserID,
		UserRole:         e.UserRole,
		CertType:         e.CertType,
		CertNumber:       e.CertNumber,
		RealName:         e.RealName,
		DriverLicenseNo:  e.DriverLicenseNo,
		LicenseType:      e.LicenseType,
		City:             e.City,
		FrontImageURL:    e.FrontImageURL,
		BackImageURL:     e.BackImageURL,
		HandheldImageURL: e.HandheldImageURL,
		Status:           e.Status,
		SubmitCount:      e.SubmitCount,
	}
}

func vehicleToDomain(m *driverVehicleModel) *biz.DriverVehicle {
	return &biz.DriverVehicle{
		ID:          m.ID,
		DriverID:    m.DriverID,
		PlateNo:     m.PlateNo,
		Brand:       m.Brand,
		Model:       m.Model,
		Color:       m.Color,
		VehicleType: m.VehicleType,
		Seats:       m.Seats,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func vehicleToRecord(e *biz.DriverVehicle) *driverVehicleModel {
	return &driverVehicleModel{
		ID:          e.ID,
		DriverID:    e.DriverID,
		PlateNo:     e.PlateNo,
		Brand:       e.Brand,
		Model:       e.Model,
		Color:       e.Color,
		VehicleType: e.VehicleType,
		Seats:       e.Seats,
		Status:      e.Status,
	}
}

func vehicleAuditToDomain(m *vehicleAuditModel) *biz.DriverVehicleAudit {
	return &biz.DriverVehicleAudit{
		ID:           m.ID,
		DriverID:     m.DriverID,
		PlateNumber:  m.PlateNumber,
		Brand:        m.Brand,
		Model:        m.Model,
		Color:        m.Color,
		Seats:        m.Seats,
		Status:       m.Status,
		ReviewerID:   m.ReviewerID,
		RejectReason: m.RejectReason,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func vehicleAuditToRecord(e *biz.DriverVehicleAudit) *vehicleAuditModel {
	return &vehicleAuditModel{
		ID:           e.ID,
		DriverID:     e.DriverID,
		PlateNumber:  e.PlateNumber,
		Brand:        e.Brand,
		Model:        e.Model,
		Color:        e.Color,
		Seats:        e.Seats,
		Status:       e.Status,
		ReviewerID:   e.ReviewerID,
		RejectReason: e.RejectReason,
	}
}

func locationToDomain(m *driverLocationPointModel) *biz.DriverLocationPoint {
	return &biz.DriverLocationPoint{
		ID:         m.ID,
		DriverID:   m.DriverID,
		OrderID:    m.OrderID,
		Latitude:   m.Latitude,
		Longitude:  m.Longitude,
		Speed:      m.Speed,
		Heading:    m.Heading,
		ReportedAt: m.ReportedAt,
		CreatedAt:  m.CreatedAt,
	}
}

func messageToDomain(m *realtimeMessageModel) *biz.DriverMessage {
	title := ""
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(m.Payload), &payload); err == nil {
		if v, ok := payload["title"].(string); ok {
			title = v
		}
	}
	if title == "" {
		switch m.Topic {
		case "vehicle.audit":
			title = "车辆审核通知"
		case "certification.audit":
			title = "认证审核通知"
		default:
			title = "系统通知"
		}
	}
	return &biz.DriverMessage{
		ID:        int64(m.ID),
		Topic:     m.Topic,
		Title:     title,
		Payload:   m.Payload,
		Delivered: m.Delivered,
		CreatedAt: m.CreatedAt,
	}
}

func locationToRecord(e *biz.DriverLocationPoint) *driverLocationPointModel {
	return &driverLocationPointModel{
		ID:         e.ID,
		DriverID:   e.DriverID,
		OrderID:    e.OrderID,
		Latitude:   e.Latitude,
		Longitude:  e.Longitude,
		Speed:      e.Speed,
		Heading:    e.Heading,
		ReportedAt: e.ReportedAt,
	}
}
