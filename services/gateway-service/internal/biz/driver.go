package biz

import (
	"context"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/gateway-service/internal/data"
)

type DriverUsecase struct {
	client data.DriverClient
}

func NewDriverUsecase(client data.DriverClient) *DriverUsecase {
	return &DriverUsecase{client: client}
}

func (uc *DriverUsecase) EnsureDriver(ctx context.Context, id int64, phone string) (*driverv1.DriverProfileReply, error) {
	return uc.client.EnsureDriver(ctx, id, phone)
}

func (uc *DriverUsecase) GetDriver(ctx context.Context, id int64) (*driverv1.DriverProfileReply, error) {
	return uc.client.GetDriver(ctx, id)
}

func (uc *DriverUsecase) UpdateDriver(ctx context.Context, req *driverv1.UpdateDriverRequest) (*driverv1.DriverProfileReply, error) {
	return uc.client.UpdateDriver(ctx, req)
}

func (uc *DriverUsecase) SubmitCertification(ctx context.Context, req *driverv1.SubmitCertificationRequest) (*driverv1.CertificationReply, error) {
	return uc.client.SubmitCertification(ctx, req)
}

func (uc *DriverUsecase) GetCertification(ctx context.Context, id int64) (*driverv1.CertificationReply, error) {
	return uc.client.GetCertification(ctx, id)
}

func (uc *DriverUsecase) SaveVehicle(ctx context.Context, req *driverv1.SaveVehicleRequest) (*driverv1.VehicleReply, error) {
	return uc.client.SaveVehicle(ctx, req)
}

func (uc *DriverUsecase) UpdateVehicle(ctx context.Context, req *driverv1.UpdateVehicleRequest) (*driverv1.VehicleReply, error) {
	return uc.client.UpdateVehicle(ctx, req)
}

func (uc *DriverUsecase) DeleteVehicle(ctx context.Context, req *driverv1.DeleteVehicleRequest) (*driverv1.DeleteVehicleReply, error) {
	return uc.client.DeleteVehicle(ctx, req)
}

func (uc *DriverUsecase) ListVehicles(ctx context.Context, id int64) (*driverv1.ListVehiclesReply, error) {
	return uc.client.ListVehicles(ctx, id)
}

func (uc *DriverUsecase) ListMessages(ctx context.Context, driverID int64) (*driverv1.ListMessagesReply, error) {
	return uc.client.ListMessages(ctx, driverID)
}

func (uc *DriverUsecase) AckMessage(ctx context.Context, req *driverv1.AckMessageRequest) (*driverv1.AckMessageReply, error) {
	return uc.client.AckMessage(ctx, req)
}

func (uc *DriverUsecase) ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error) {
	return uc.client.ReportDriverLocation(ctx, req)
}

func (uc *DriverUsecase) ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error) {
	return uc.client.ReplayDriverTrack(ctx, req)
}
