package data

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/driver-service/internal/biz"
)

func TestDriverRepoPersistsProfileCertificationAndVehicles(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:driver_repo_core?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&driverProfileModel{}, &driverCertificationModel{}, &certificationAuditModel{}, &driverVehicleModel{}, &vehicleAuditModel{}))

	repo := NewDriverRepo(db, zap.NewNop())
	err = repo.CreateProfile(context.Background(), &biz.DriverProfile{
		ID:                  2001,
		Name:                "Driver 2001",
		ServiceStatus:       biz.DriverServiceStatusOffline,
		CertificationStatus: biz.CertificationStatusDraft,
	})
	require.NoError(t, err)

	profile, err := repo.GetProfileByID(context.Background(), 2001)
	require.NoError(t, err)
	require.Equal(t, "Driver 2001", profile.Name)

	require.NoError(t, repo.SaveCertification(context.Background(), &biz.DriverCertification{
		ID:          3001,
		DriverID:    2001,
		RealName:    "Bob",
		IDCardNo:    "110101199001011111",
		LicenseNo:   "DL001",
		LicenseType: "C1",
		City:        "上海市",
		Status:      biz.CertificationStatusPending,
	}))
	cert, err := repo.GetCertification(context.Background(), 2001)
	require.NoError(t, err)
	require.Equal(t, "Bob", cert.RealName)
	require.Equal(t, "110101199001011111", cert.IDCardNo)

	require.NoError(t, repo.SaveCertificationAudit(context.Background(), &biz.CertificationAudit{
		ID:              5001,
		UserID:          2001,
		UserRole:        2,
		CertType:        1,
		CertNumber:      "110101199001011111",
		RealName:        "Bob",
		DriverLicenseNo: "DL001",
		LicenseType:     "C1",
		City:            "上海市",
		Status:          0,
		SubmitCount:     1,
	}))

	require.NoError(t, repo.SaveVehicle(context.Background(), &biz.DriverVehicle{
		ID:       4001,
		DriverID: 2001,
		PlateNo:  "沪A12345",
		Brand:    "BYD",
		Seats:    4,
		Status:   biz.VehicleStatusActive,
	}))
	vehicles, err := repo.ListVehicles(context.Background(), 2001)
	require.NoError(t, err)
	require.Len(t, vehicles, 1)
	require.Equal(t, "沪A12345", vehicles[0].PlateNo)
	vehicle, err := repo.GetVehicleByID(context.Background(), 4001)
	require.NoError(t, err)
	vehicle.PlateNo = "ZJA54321"
	vehicle.Seats = 5
	require.NoError(t, repo.UpdateVehicle(context.Background(), vehicle))

	updated, err := repo.GetVehicleByID(context.Background(), 4001)
	require.NoError(t, err)
	require.Equal(t, "ZJA54321", updated.PlateNo)
	require.Equal(t, 5, updated.Seats)

	updated.Status = biz.VehicleStatusInactive
	require.NoError(t, repo.UpdateVehicle(context.Background(), updated))
	vehicles, err = repo.ListVehicles(context.Background(), 2001)
	require.NoError(t, err)
	require.Empty(t, vehicles)
}

func TestDriverRepoSaveVehicleAuditAllowsSameDriverResubmitAndRejectsOtherDriverPlate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:vehicle_audit_resubmit?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&vehicleAuditModel{}))
	repo := NewDriverRepo(db, zap.NewNop())

	require.NoError(t, repo.SaveVehicleAudit(context.Background(), &biz.DriverVehicleAudit{
		ID:          7001,
		DriverID:    2001,
		PlateNumber: "沪A12345",
		Brand:       "BYD",
		Model:       "秦 PLUS",
		Color:       "白色",
		Seats:       5,
		Status:      biz.VehicleAuditStatusPending,
	}))
	audits, err := repo.ListVehicleAudits(context.Background(), 2001)
	require.NoError(t, err)
	require.Len(t, audits, 1)
	require.Equal(t, biz.VehicleAuditStatusPending, audits[0].Status)
	require.Equal(t, 5, audits[0].Seats)

	require.NoError(t, db.Model(&vehicleAuditModel{}).Where("id = ?", int64(7001)).Updates(map[string]interface{}{
		"status":        biz.VehicleAuditStatusRejected,
		"reject_reason": "照片不清晰",
	}).Error)
	resubmit := &biz.DriverVehicleAudit{
		ID:          7002,
		DriverID:    2001,
		PlateNumber: "沪A12345",
		Brand:       "BYD",
		Model:       "汉",
		Color:       "黑色",
		Seats:       6,
		Status:      biz.VehicleAuditStatusPending,
	}
	require.NoError(t, repo.SaveVehicleAudit(context.Background(), resubmit))
	require.Equal(t, int64(7001), resubmit.ID)

	var saved vehicleAuditModel
	require.NoError(t, db.First(&saved, int64(7001)).Error)
	require.Equal(t, biz.VehicleAuditStatusPending, saved.Status)
	require.Equal(t, 6, saved.Seats)
	require.Equal(t, "汉", saved.Model)
	require.Empty(t, saved.RejectReason)

	err = repo.SaveVehicleAudit(context.Background(), &biz.DriverVehicleAudit{
		ID:          7003,
		DriverID:    2002,
		PlateNumber: "沪A12345",
		Brand:       "Tesla",
		Model:       "Model 3",
		Status:      biz.VehicleAuditStatusPending,
	})
	require.ErrorIs(t, err, biz.ErrVehiclePlateInUse)
}

func TestDriverRepoPersistsAndReplaysTrackPoints(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&driverLocationPointModel{}))

	repo := NewDriverRepo(db, zap.NewNop())
	base := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)

	require.NoError(t, repo.SaveLocation(context.Background(), &biz.DriverLocationPoint{
		ID:         5002,
		DriverID:   2001,
		OrderID:    3001,
		Latitude:   30.223,
		Longitude:  120.112,
		Speed:      12.3,
		Heading:    95,
		ReportedAt: base.Add(2 * time.Minute),
	}))
	require.NoError(t, repo.SaveLocation(context.Background(), &biz.DriverLocationPoint{
		ID:         5001,
		DriverID:   2001,
		OrderID:    3001,
		Latitude:   30.221,
		Longitude:  120.11,
		Speed:      10.5,
		Heading:    90,
		ReportedAt: base,
	}))
	require.NoError(t, repo.SaveLocation(context.Background(), &biz.DriverLocationPoint{
		ID:         6001,
		DriverID:   2002,
		OrderID:    3001,
		Latitude:   31,
		Longitude:  121,
		ReportedAt: base,
	}))

	points, total, err := repo.ReplayTrack(context.Background(), biz.TrackReplayQuery{
		DriverID: 2001,
		OrderID:  3001,
		Page:     1,
		PageSize: 1,
	})

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, points, 1)
	require.Equal(t, int64(5001), points[0].ID)
	require.Equal(t, base, points[0].ReportedAt)
}
