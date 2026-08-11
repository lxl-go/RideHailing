package service

import (
	"context"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/driver-service/internal/biz"
)

type fakeDriverRepo struct {
	profiles map[int64]*biz.DriverProfile
	vehicles map[int64]*biz.DriverVehicle
	audits   map[int64]*biz.DriverVehicleAudit
}

func (r *fakeDriverRepo) GetProfileByID(_ context.Context, id int64) (*biz.DriverProfile, error) {
	item, ok := r.profiles[id]
	if !ok {
		return nil, biz.ErrDriverNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeDriverRepo) CreateProfile(_ context.Context, profile *biz.DriverProfile) error {
	copy := *profile
	r.profiles[profile.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) UpdateProfile(_ context.Context, profile *biz.DriverProfile) error {
	copy := *profile
	r.profiles[profile.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) SaveCertification(_ context.Context, _ *biz.DriverCertification) error {
	return nil
}

func (r *fakeDriverRepo) SaveCertificationAudit(_ context.Context, _ *biz.CertificationAudit) error {
	return nil
}

func (r *fakeDriverRepo) GetCertification(_ context.Context, _ int64) (*biz.DriverCertification, error) {
	return nil, biz.ErrCertificationNotFound
}

func (r *fakeDriverRepo) SaveVehicle(_ context.Context, vehicle *biz.DriverVehicle) error {
	copy := *vehicle
	r.vehicles[vehicle.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) SaveVehicleAudit(_ context.Context, audit *biz.DriverVehicleAudit) error {
	if r.audits == nil {
		r.audits = map[int64]*biz.DriverVehicleAudit{}
	}
	copy := *audit
	r.audits[audit.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) MarkVehicleAuditDriverDeleted(_ context.Context, _ int64, _ string) error {
	return nil
}

func (r *fakeDriverRepo) GetVehicleByID(_ context.Context, id int64) (*biz.DriverVehicle, error) {
	item, ok := r.vehicles[id]
	if !ok {
		return nil, biz.ErrVehicleNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeDriverRepo) UpdateVehicle(_ context.Context, vehicle *biz.DriverVehicle) error {
	copy := *vehicle
	r.vehicles[vehicle.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) ListVehicles(_ context.Context, _ int64) ([]biz.DriverVehicle, error) {
	return nil, nil
}

func (r *fakeDriverRepo) ListVehicleAudits(_ context.Context, _ int64) ([]biz.DriverVehicleAudit, error) {
	return nil, nil
}

func (r *fakeDriverRepo) ListDriverMessages(_ context.Context, _ int64) ([]biz.DriverMessage, error) {
	return nil, nil
}

func (r *fakeDriverRepo) AckDriverMessage(_ context.Context, _ int64, _ int64) error {
	return nil
}

func (r *fakeDriverRepo) SaveLocation(_ context.Context, _ *biz.DriverLocationPoint) error {
	return nil
}

func (r *fakeDriverRepo) ReplayTrack(_ context.Context, _ biz.TrackReplayQuery) ([]biz.DriverLocationPoint, int64, error) {
	return nil, 0, nil
}

func TestUpdateDriverAcceptsServiceStatus(t *testing.T) {
	node, err := snowflake.NewNode(5)
	require.NoError(t, err)
	repo := &fakeDriverRepo{
		profiles: map[int64]*biz.DriverProfile{
			2001: {ID: 2001, Name: "Driver", ServiceStatus: biz.DriverServiceStatusOffline},
		},
		vehicles: map[int64]*biz.DriverVehicle{},
	}
	svc := NewDriverService(biz.NewDriverUsecase(node, zap.NewNop(), repo))

	reply, err := svc.UpdateDriver(context.Background(), &driverv1.UpdateDriverRequest{
		Id:            2001,
		Name:          "Driver",
		ServiceStatus: int32(biz.DriverServiceStatusOnline),
	})

	require.NoError(t, err)
	require.Equal(t, int32(biz.DriverServiceStatusOnline), reply.Driver.ServiceStatus)
	require.Equal(t, biz.DriverServiceStatusOnline, repo.profiles[2001].ServiceStatus)
}

func TestVehicleRPCsRequireDriverOwnership(t *testing.T) {
	node, err := snowflake.NewNode(5)
	require.NoError(t, err)
	repo := &fakeDriverRepo{
		profiles: map[int64]*biz.DriverProfile{},
		vehicles: map[int64]*biz.DriverVehicle{
			4001: {ID: 4001, DriverID: 2001, PlateNo: "沪A11111", Brand: "BYD", Seats: 4, Status: biz.VehicleStatusActive},
		},
	}
	svc := NewDriverService(biz.NewDriverUsecase(node, zap.NewNop(), repo))

	reply, err := svc.UpdateVehicle(context.Background(), &driverv1.UpdateVehicleRequest{
		DriverId:  2001,
		VehicleId: 4001,
		PlateNo:   "沪A22222",
		Brand:     "Tesla",
		Seats:     5,
	})
	require.NoError(t, err)
	require.Equal(t, "沪A22222", reply.Vehicle.PlateNo)
	require.Equal(t, int32(biz.VehicleAuditStatusPending), reply.Vehicle.Status)
	require.Equal(t, "audit", reply.Vehicle.Source)
	require.False(t, reply.Vehicle.CanDelete)

	_, err = svc.UpdateVehicle(context.Background(), &driverv1.UpdateVehicleRequest{
		DriverId:  2002,
		VehicleId: 4001,
		PlateNo:   "沪A33333",
		Brand:     "Hack",
		Seats:     4,
	})
	require.Error(t, err)

	_, err = svc.DeleteVehicle(context.Background(), &driverv1.DeleteVehicleRequest{
		DriverId:  2001,
		VehicleId: 4001,
	})
	require.NoError(t, err)
	require.Equal(t, biz.VehicleStatusInactive, repo.vehicles[4001].Status)
}

func TestMapErrorUsesChineseRealNameMessages(t *testing.T) {
	err := mapError(biz.ErrRealNameNotMatched)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.Equal(t, "实名认证未通过：真实姓名和身份证号不匹配，或身份证信息库无记录，请核对后重试", status.Convert(err).Message())

	err = mapError(biz.ErrRealNameUnavailable)
	require.Equal(t, codes.Unavailable, status.Code(err))
	require.Equal(t, "实名认证服务暂不可用，请稍后重试", status.Convert(err).Message())

	err = mapError(biz.ErrCertificationNotFound)
	require.Equal(t, codes.NotFound, status.Code(err))
	require.Equal(t, "暂未提交司机认证资料", status.Convert(err).Message())
}
