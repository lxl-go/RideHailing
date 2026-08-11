package biz

import "context"

type DriverRepo interface {
	GetProfileByID(ctx context.Context, id int64) (*DriverProfile, error)
	CreateProfile(ctx context.Context, profile *DriverProfile) error
	UpdateProfile(ctx context.Context, profile *DriverProfile) error
	SaveCertification(ctx context.Context, cert *DriverCertification) error
	SaveCertificationAudit(ctx context.Context, audit *CertificationAudit) error
	GetCertification(ctx context.Context, driverID int64) (*DriverCertification, error)
	SaveVehicle(ctx context.Context, vehicle *DriverVehicle) error
	SaveVehicleAudit(ctx context.Context, audit *DriverVehicleAudit) error
	MarkVehicleAuditDriverDeleted(ctx context.Context, driverID int64, plateNo string) error
	GetVehicleByID(ctx context.Context, id int64) (*DriverVehicle, error)
	UpdateVehicle(ctx context.Context, vehicle *DriverVehicle) error
	ListVehicles(ctx context.Context, driverID int64) ([]DriverVehicle, error)
	ListVehicleAudits(ctx context.Context, driverID int64) ([]DriverVehicleAudit, error)
	ListDriverMessages(ctx context.Context, driverID int64) ([]DriverMessage, error)
	AckDriverMessage(ctx context.Context, driverID int64, messageID int64) error
	SaveLocation(ctx context.Context, point *DriverLocationPoint) error
	ReplayTrack(ctx context.Context, query TrackReplayQuery) ([]DriverLocationPoint, int64, error)
}
