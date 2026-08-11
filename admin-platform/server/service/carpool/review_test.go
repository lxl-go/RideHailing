package carpool

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
)

type driverProfileReviewTestModel struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	CertificationStatus int       `gorm:"column:certification_status;type:tinyint;not null;default:1"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (driverProfileReviewTestModel) TableName() string {
	return "driver_profile"
}

type driverCertificationReviewTestModel struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID     int64     `gorm:"column:driver_id;type:bigint;not null;uniqueIndex:uk_driver_certification"`
	Status       int       `gorm:"column:status;type:tinyint;not null;default:2"`
	RejectReason string    `gorm:"column:reject_reason;type:varchar(255);not null;default:''"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (driverCertificationReviewTestModel) TableName() string {
	return "driver_certification"
}

type driverVehicleReviewTestModel struct {
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

func (driverVehicleReviewTestModel) TableName() string {
	return "driver_vehicle"
}

func newReviewServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.CertificationAudit{},
		&carpoolModel.VehicleInfo{},
		&carpoolModel.RealtimeMessage{},
		&driverProfileReviewTestModel{},
		&driverCertificationReviewTestModel{},
		&driverVehicleReviewTestModel{},
	))
	global.GVA_DB = db
	global.GVA_LOG = zap.NewNop()
	return db
}

func TestRejectAuditSyncsDriverCertificationAndCreatesMessage(t *testing.T) {
	db := newReviewServiceTestDB(t)
	ctx := context.Background()
	audit := carpoolModel.CertificationAudit{
		ID:          9001,
		UserID:      2001,
		UserRole:    2,
		CertType:    1,
		CertNumber:  "110101199001011111",
		RealName:    "Bob",
		Status:      0,
		SubmitCount: 1,
		CreatedAt:   time.Now().Add(-2 * time.Hour),
	}
	require.NoError(t, db.Create(&audit).Error)
	require.NoError(t, db.Create(&driverProfileReviewTestModel{ID: 2001, CertificationStatus: 2}).Error)
	require.NoError(t, db.Create(&driverCertificationReviewTestModel{ID: 3001, DriverID: 2001, Status: 2}).Error)

	err := (&ReviewService{}).RejectAudit(ctx, audit.ID, 88, "驾驶证照片不清晰")

	require.NoError(t, err)
	var savedAudit carpoolModel.CertificationAudit
	require.NoError(t, db.First(&savedAudit, audit.ID).Error)
	require.Equal(t, 2, savedAudit.Status)
	require.Equal(t, "驾驶证照片不清晰", savedAudit.RejectReason)

	var cert driverCertificationReviewTestModel
	require.NoError(t, db.Where("driver_id = ?", int64(2001)).First(&cert).Error)
	require.Equal(t, 4, cert.Status)
	require.Equal(t, "驾驶证照片不清晰", cert.RejectReason)

	var profile driverProfileReviewTestModel
	require.NoError(t, db.First(&profile, int64(2001)).Error)
	require.Equal(t, 4, profile.CertificationStatus)

	var msg carpoolModel.RealtimeMessage
	require.NoError(t, db.Where("user_id = ? AND user_role = ? AND topic = ?", uint64(2001), "driver", "certification.audit").First(&msg).Error)
	require.Contains(t, msg.Payload, "rejected")
	require.Contains(t, msg.Payload, "驾驶证照片不清晰")
}

func TestApproveAuditSyncsDriverCertification(t *testing.T) {
	db := newReviewServiceTestDB(t)
	ctx := context.Background()
	audit := carpoolModel.CertificationAudit{
		ID:          9002,
		UserID:      2002,
		UserRole:    2,
		CertType:    1,
		CertNumber:  "110101199001011112",
		RealName:    "Alice",
		Status:      0,
		SubmitCount: 1,
		CreatedAt:   time.Now().Add(-time.Hour),
	}
	require.NoError(t, db.Create(&audit).Error)
	require.NoError(t, db.Create(&driverProfileReviewTestModel{ID: 2002, CertificationStatus: 2}).Error)
	require.NoError(t, db.Create(&driverCertificationReviewTestModel{ID: 3002, DriverID: 2002, Status: 2, RejectReason: "old"}).Error)

	err := (&ReviewService{}).ApproveAudit(ctx, audit.ID, 88)

	require.NoError(t, err)
	var cert driverCertificationReviewTestModel
	require.NoError(t, db.Where("driver_id = ?", int64(2002)).First(&cert).Error)
	require.Equal(t, 3, cert.Status)
	require.Empty(t, cert.RejectReason)

	var profile driverProfileReviewTestModel
	require.NoError(t, db.First(&profile, int64(2002)).Error)
	require.Equal(t, 3, profile.CertificationStatus)
}

func TestHandleVehicleReviewApproveSyncsActiveDriverVehicle(t *testing.T) {
	db := newReviewServiceTestDB(t)
	ctx := context.Background()
	info := carpoolModel.VehicleInfo{
		ID:          9101,
		DriverID:    2001,
		PlateNumber: "沪A12345",
		Brand:       "BYD",
		Model:       "秦 PLUS",
		Color:       "白色",
		Seats:       5,
		Status:      0,
		CreatedAt:   time.Now().Add(-time.Hour),
	}
	require.NoError(t, db.Create(&info).Error)

	err := (&ReviewService{}).HandleVehicleReview(ctx, info.ID, 88, 1, "")

	require.NoError(t, err)
	var savedInfo carpoolModel.VehicleInfo
	require.NoError(t, db.First(&savedInfo, info.ID).Error)
	require.Equal(t, 1, savedInfo.Status)
	require.Equal(t, int64(88), savedInfo.ReviewerID)

	var vehicle driverVehicleReviewTestModel
	require.NoError(t, db.Where("driver_id = ? AND plate_no = ?", int64(2001), "沪A12345").First(&vehicle).Error)
	require.Equal(t, int64(9101), vehicle.ID)
	require.Equal(t, "BYD", vehicle.Brand)
	require.Equal(t, "秦 PLUS", vehicle.Model)
	require.Equal(t, "白色", vehicle.Color)
	require.Equal(t, 5, vehicle.Seats)
	require.Equal(t, 1, vehicle.Status)
}

func TestHandleVehicleReviewReapproveAfterPlateChangeUpdatesOriginVehicle(t *testing.T) {
	db := newReviewServiceTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.Create(&driverVehicleReviewTestModel{
		ID: 9101, DriverID: 2001, PlateNo: "沪A12345",
		Brand: "BYD", Model: "秦 PLUS", Color: "白色", Seats: 5, Status: 1,
	}).Error)

	updated := carpoolModel.VehicleInfo{
		ID: 9101, DriverID: 2001, PlateNumber: "沪A99999",
		Brand: "BYD", Model: "秦 PLUS", Color: "红色", Seats: 4, Status: 0,
		CreatedAt: time.Now().Add(-time.Hour),
	}
	require.NoError(t, db.Create(&updated).Error)

	err := (&ReviewService{}).HandleVehicleReview(ctx, updated.ID, 88, 1, "")
	require.NoError(t, err)

	var count int64
	require.NoError(t, db.Model(&driverVehicleReviewTestModel{}).Count(&count).Error)
	require.Equal(t, int64(1), count)

	var vehicle driverVehicleReviewTestModel
	require.NoError(t, db.First(&vehicle, int64(9101)).Error)
	require.Equal(t, "沪A99999", vehicle.PlateNo)
	require.Equal(t, "红色", vehicle.Color)
	require.Equal(t, 4, vehicle.Seats)
	require.Equal(t, 1, vehicle.Status)
}

func TestHandleVehicleReviewRejectCreatesDriverMessage(t *testing.T) {
	db := newReviewServiceTestDB(t)
	ctx := context.Background()
	info := carpoolModel.VehicleInfo{
		ID:          9102,
		DriverID:    2002,
		PlateNumber: "沪B12345",
		Brand:       "Tesla",
		Model:       "Model 3",
		Color:       "黑色",
		Status:      0,
		CreatedAt:   time.Now().Add(-time.Hour),
	}
	require.NoError(t, db.Create(&info).Error)

	err := (&ReviewService{}).HandleVehicleReview(ctx, info.ID, 88, 2, "车辆照片不清晰")

	require.NoError(t, err)
	var savedInfo carpoolModel.VehicleInfo
	require.NoError(t, db.First(&savedInfo, info.ID).Error)
	require.Equal(t, 2, savedInfo.Status)
	require.Equal(t, "车辆照片不清晰", savedInfo.RejectReason)

	var msg carpoolModel.RealtimeMessage
	require.NoError(t, db.Where("user_id = ? AND user_role = ? AND topic = ?", uint64(2002), "driver", "vehicle.audit").First(&msg).Error)
	require.Contains(t, msg.Payload, "rejected")
	require.Contains(t, msg.Payload, "车辆照片不清晰")
}
