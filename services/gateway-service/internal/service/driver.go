package service

import (
	"context"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/gateway-service/internal/biz"
)

type DriverService struct {
	uc *biz.DriverUsecase
}

func NewDriverService(uc *biz.DriverUsecase) *DriverService {
	return &DriverService{uc: uc}
}

func (s *DriverService) EnsureDriver(ctx context.Context, id int64, phone string) (*driverv1.DriverProfileReply, error) {
	return s.uc.EnsureDriver(ctx, id, phone)
}

func (s *DriverService) GetDriver(ctx context.Context, id int64) (*driverv1.DriverProfileReply, error) {
	return s.uc.GetDriver(ctx, id)
}

func (s *DriverService) UpdateDriver(ctx context.Context, req *driverv1.UpdateDriverRequest) (*driverv1.DriverProfileReply, error) {
	return s.uc.UpdateDriver(ctx, req)
}

func (s *DriverService) SubmitCertification(ctx context.Context, req *driverv1.SubmitCertificationRequest) (*driverv1.CertificationReply, error) {
	return s.uc.SubmitCertification(ctx, req)
}

func (s *DriverService) GetCertification(ctx context.Context, id int64) (*driverv1.CertificationReply, error) {
	return s.uc.GetCertification(ctx, id)
}

func (s *DriverService) SaveVehicle(ctx context.Context, req *driverv1.SaveVehicleRequest) (*driverv1.VehicleReply, error) {
	return s.uc.SaveVehicle(ctx, req)
}

func (s *DriverService) UpdateVehicle(ctx context.Context, req *driverv1.UpdateVehicleRequest) (*driverv1.VehicleReply, error) {
	return s.uc.UpdateVehicle(ctx, req)
}

func (s *DriverService) DeleteVehicle(ctx context.Context, req *driverv1.DeleteVehicleRequest) (*driverv1.DeleteVehicleReply, error) {
	return s.uc.DeleteVehicle(ctx, req)
}

func (s *DriverService) ListVehicles(ctx context.Context, id int64) (*driverv1.ListVehiclesReply, error) {
	return s.uc.ListVehicles(ctx, id)
}

func (s *DriverService) ListMessages(ctx context.Context, driverID int64) (*driverv1.ListMessagesReply, error) {
	return s.uc.ListMessages(ctx, driverID)
}

func (s *DriverService) AckMessage(ctx context.Context, req *driverv1.AckMessageRequest) (*driverv1.AckMessageReply, error) {
	return s.uc.AckMessage(ctx, req)
}

func (s *DriverService) ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error) {
	return s.uc.ReportDriverLocation(ctx, req)
}

func (s *DriverService) ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error) {
	return s.uc.ReplayDriverTrack(ctx, req)
}
