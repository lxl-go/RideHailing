package biz

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ride-hailing/pkg/realname"
)

type fakeDriverRepo struct {
	profiles            map[int64]*DriverProfile
	certifications      map[int64]*DriverCertification
	certificationAudits map[int64]*CertificationAudit
	vehicles            map[int64][]DriverVehicle
	vehicleAudits       map[int64][]DriverVehicleAudit
	messages            map[int64][]DriverMessage
	locations           []DriverLocationPoint
}

func newFakeDriverRepo() *fakeDriverRepo {
	return &fakeDriverRepo{
		profiles:            map[int64]*DriverProfile{},
		certifications:      map[int64]*DriverCertification{},
		certificationAudits: map[int64]*CertificationAudit{},
		vehicles:            map[int64][]DriverVehicle{},
		vehicleAudits:       map[int64][]DriverVehicleAudit{},
		messages:            map[int64][]DriverMessage{},
	}
}

func (r *fakeDriverRepo) GetProfileByID(_ context.Context, id int64) (*DriverProfile, error) {
	item, ok := r.profiles[id]
	if !ok {
		return nil, ErrDriverNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeDriverRepo) CreateProfile(_ context.Context, profile *DriverProfile) error {
	copy := *profile
	r.profiles[profile.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) UpdateProfile(_ context.Context, profile *DriverProfile) error {
	copy := *profile
	r.profiles[profile.ID] = &copy
	return nil
}

func (r *fakeDriverRepo) SaveCertification(_ context.Context, cert *DriverCertification) error {
	copy := *cert
	r.certifications[cert.DriverID] = &copy
	return nil
}

func (r *fakeDriverRepo) SaveCertificationAudit(_ context.Context, audit *CertificationAudit) error {
	copy := *audit
	r.certificationAudits[audit.UserID] = &copy
	return nil
}

func (r *fakeDriverRepo) GetCertification(_ context.Context, driverID int64) (*DriverCertification, error) {
	item, ok := r.certifications[driverID]
	if !ok {
		return nil, ErrCertificationNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeDriverRepo) SaveVehicle(_ context.Context, vehicle *DriverVehicle) error {
	copy := *vehicle
	r.vehicles[vehicle.DriverID] = append(r.vehicles[vehicle.DriverID], copy)
	return nil
}

func (r *fakeDriverRepo) SaveVehicleAudit(_ context.Context, audit *DriverVehicleAudit) error {
	copy := *audit
	for driverID, items := range r.vehicleAudits {
		for i := range items {
			if items[i].ID == audit.ID || (items[i].DriverID == audit.DriverID && items[i].PlateNumber == audit.PlateNumber) {
				r.vehicleAudits[driverID][i] = copy
				return nil
			}
		}
	}
	r.vehicleAudits[audit.DriverID] = append(r.vehicleAudits[audit.DriverID], copy)
	return nil
}

func (r *fakeDriverRepo) MarkVehicleAuditDriverDeleted(_ context.Context, driverID int64, plateNo string) error {
	for i := range r.vehicleAudits[driverID] {
		if r.vehicleAudits[driverID][i].PlateNumber == plateNo {
			r.vehicleAudits[driverID][i].Status = VehicleAuditStatusDriverDeleted
			return nil
		}
	}
	return nil
}

func (r *fakeDriverRepo) GetVehicleByID(_ context.Context, id int64) (*DriverVehicle, error) {
	for _, items := range r.vehicles {
		for _, item := range items {
			if item.ID == id {
				copy := item
				return &copy, nil
			}
		}
	}
	return nil, ErrVehicleNotFound
}

func (r *fakeDriverRepo) UpdateVehicle(_ context.Context, vehicle *DriverVehicle) error {
	for driverID, items := range r.vehicles {
		for i := range items {
			if items[i].ID == vehicle.ID {
				copy := *vehicle
				r.vehicles[driverID][i] = copy
				return nil
			}
		}
	}
	return ErrVehicleNotFound
}

func (r *fakeDriverRepo) ListVehicles(_ context.Context, driverID int64) ([]DriverVehicle, error) {
	var active []DriverVehicle
	for _, item := range r.vehicles[driverID] {
		if item.Status == VehicleStatusActive {
			active = append(active, item)
		}
	}
	return active, nil
}

func (r *fakeDriverRepo) ListVehicleAudits(_ context.Context, driverID int64) ([]DriverVehicleAudit, error) {
	return append([]DriverVehicleAudit(nil), r.vehicleAudits[driverID]...), nil
}

func (r *fakeDriverRepo) ListDriverMessages(_ context.Context, driverID int64) ([]DriverMessage, error) {
	var items []DriverMessage
	for _, item := range r.messages[driverID] {
		if !item.Delivered {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeDriverRepo) AckDriverMessage(_ context.Context, driverID int64, messageID int64) error {
	for i := range r.messages[driverID] {
		if r.messages[driverID][i].ID == messageID {
			r.messages[driverID][i].Delivered = true
			return nil
		}
	}
	return nil
}

func (r *fakeDriverRepo) SaveLocation(_ context.Context, point *DriverLocationPoint) error {
	copy := *point
	r.locations = append(r.locations, copy)
	return nil
}

func (r *fakeDriverRepo) ReplayTrack(_ context.Context, query TrackReplayQuery) ([]DriverLocationPoint, int64, error) {
	var items []DriverLocationPoint
	for _, item := range r.locations {
		if item.DriverID == query.DriverID && (query.OrderID == 0 || item.OrderID == query.OrderID) {
			items = append(items, item)
		}
	}
	return items, int64(len(items)), nil
}

func TestSubmitCertificationSetsPendingStatus(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	uc := NewDriverUsecase(node, zap.NewNop(), repo, stubRealNameVerifier{
		result: realname.Result{ErrorCode: 0, Matched: true, City: "上海市"},
	})

	cert, err := uc.SubmitCertification(context.Background(), SubmitCertificationCommand{
		DriverID:         2001,
		RealName:         "  Bob  ",
		IDCardNo:         " 110101199001011111 ",
		LicenseNo:        "  DL001  ",
		LicenseType:      " C1 ",
		VehicleLicenseNo: "  VL001  ",
	})

	require.NoError(t, err)
	require.Equal(t, CertificationStatusPending, cert.Status)
	require.Equal(t, "Bob", cert.RealName)
	require.Equal(t, "110101199001011111", cert.IDCardNo)
	require.Equal(t, "DL001", cert.LicenseNo)
	require.Contains(t, repo.certifications, int64(2001))
}

func TestSaveVehicleCreatesPendingAuditInsteadOfActiveVehicle(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	vehicle, err := uc.SaveVehicle(context.Background(), SaveVehicleCommand{
		DriverID: 2001,
		PlateNo:  "  沪A12345  ",
		Brand:    "  BYD  ",
		Seats:    4,
	})

	require.NoError(t, err)
	require.NotZero(t, vehicle.ID)
	require.Equal(t, VehicleAuditStatusPending, vehicle.Status)
	require.Equal(t, "沪A12345", vehicle.PlateNo)
	require.Equal(t, "BYD", vehicle.Brand)
	require.Empty(t, repo.vehicles[2001])
	require.Len(t, repo.vehicleAudits[2001], 1)
	require.Equal(t, VehicleAuditStatusPending, repo.vehicleAudits[2001][0].Status)
	require.Equal(t, "audit", vehicle.Source)
	require.Equal(t, vehicle.ID, vehicle.AuditID)
	require.Equal(t, VehicleAuditStatusPending, vehicle.ReviewStatus)
	require.False(t, vehicle.CanDelete)
	require.False(t, vehicle.CanEdit)
}

func TestListVehiclesMarksPendingAuditsAsNonDeletableAndHidesRejectedAudits(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.vehicleAudits[2001] = []DriverVehicleAudit{
		{ID: 5001, DriverID: 2001, PlateNumber: "沪A12345", Brand: "BYD", Model: "秦", Color: "白色", Status: VehicleAuditStatusPending},
		{ID: 5002, DriverID: 2001, PlateNumber: "沪A54321", Brand: "Tesla", Model: "Model 3", Color: "黑色", Status: VehicleAuditStatusRejected, RejectReason: "资料不清晰"},
	}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	vehicles, err := uc.ListVehicles(context.Background(), 2001)

	require.NoError(t, err)
	require.Len(t, vehicles, 1)
	require.Equal(t, int64(5001), vehicles[0].AuditID)
	require.Equal(t, "audit", vehicles[0].Source)
	require.Equal(t, VehicleAuditStatusPending, vehicles[0].ReviewStatus)
	require.False(t, vehicles[0].CanDelete)
	require.False(t, vehicles[0].CanEdit)
}

func TestListVehiclesPendingAuditOverridesActiveVehicleForSamePlate(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.vehicles[2001] = []DriverVehicle{{
		ID:       4001,
		DriverID: 2001,
		PlateNo:  "娌狝12345",
		Brand:    "BYD",
		Model:    "绉?PLUS",
		Color:    "鐧借壊",
		Seats:    4,
		Status:   VehicleStatusActive,
	}}
	repo.vehicleAudits[2001] = []DriverVehicleAudit{{
		ID:          5001,
		DriverID:    2001,
		PlateNumber: "娌狝12345",
		Brand:       "BYD",
		Model:       "绉?PLUS",
		Color:       "鐧借壊",
		Seats:       4,
		Status:      VehicleAuditStatusPending,
	}}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	vehicles, err := uc.ListVehicles(context.Background(), 2001)

	require.NoError(t, err)
	require.Len(t, vehicles, 1)
	require.Equal(t, int64(5001), vehicles[0].ID)
	require.Equal(t, int64(5001), vehicles[0].AuditID)
	require.Equal(t, "audit", vehicles[0].Source)
	require.Equal(t, VehicleAuditStatusPending, vehicles[0].Status)
	require.Equal(t, VehicleAuditStatusPending, vehicles[0].ReviewStatus)
	require.False(t, vehicles[0].CanEdit)
	require.False(t, vehicles[0].CanDelete)
}

func TestListVehiclesRejectedAuditHidesActiveVehicleForResubmit(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.vehicles[2001] = []DriverVehicle{{
		ID:       4001,
		DriverID: 2001,
		PlateNo:  "TEST-OLD",
		Brand:    "BYD",
		Model:    "Qin",
		Color:    "white",
		Seats:    4,
		Status:   VehicleStatusActive,
	}}
	repo.vehicleAudits[2001] = []DriverVehicleAudit{{
		ID:           4001,
		DriverID:     2001,
		PlateNumber:  "TEST-NEW",
		Brand:        "Tesla",
		Model:        "Model 3",
		Color:        "black",
		Seats:        5,
		Status:       VehicleAuditStatusRejected,
		RejectReason: "资料不清晰",
	}}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	vehicles, err := uc.ListVehicles(context.Background(), 2001)

	require.NoError(t, err)
	require.Empty(t, vehicles)
}

func TestSaveVehicleOnlyValidatesBasicLegality(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	vehicle, err := uc.SaveVehicle(context.Background(), SaveVehicleCommand{
		DriverID:    2001,
		PlateNo:     " 沪A12345 ",
		Brand:       " 测试品牌 ",
		Model:       " 测试车型 ",
		Color:       " 测试颜色 ",
		VehicleType: " 试驾车型 ",
		Seats:       5,
	})

	require.NoError(t, err)
	require.Equal(t, "沪A12345", vehicle.PlateNo)
	require.Equal(t, "测试品牌", vehicle.Brand)
	require.Equal(t, "测试车型", vehicle.Model)
	require.Equal(t, 5, vehicle.Seats)

	_, err = uc.SaveVehicle(context.Background(), SaveVehicleCommand{
		DriverID: 2001,
		PlateNo:  "TEST123",
		Brand:    "测试品牌",
		Seats:    5,
	})
	require.ErrorIs(t, err, ErrInvalidDriver)

	_, err = uc.SaveVehicle(context.Background(), SaveVehicleCommand{
		DriverID: 2001,
		PlateNo:  "沪A12345",
		Brand:    "测试品牌",
		Seats:    12,
	})
	require.ErrorIs(t, err, ErrInvalidDriver)
}

func TestUpdateVehicleCreatesPendingAuditForApprovedVehicle(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	originalPlate := "沪A11111"
	updatedPlate := "沪A22222"
	repo.vehicles[2001] = []DriverVehicle{{
		ID:       4001,
		DriverID: 2001,
		PlateNo:  originalPlate,
		Brand:    "BYD",
		Seats:    4,
		Status:   VehicleStatusActive,
	}}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	vehicle, err := uc.UpdateVehicle(context.Background(), UpdateVehicleCommand{
		DriverID:  2001,
		VehicleID: 4001,
		PlateNo:   " " + updatedPlate + " ",
		Brand:     " Tesla ",
		Seats:     5,
	})

	require.NoError(t, err)
	require.Equal(t, updatedPlate, vehicle.PlateNo)
	require.Equal(t, "Tesla", vehicle.Brand)
	require.Equal(t, 5, vehicle.Seats)
	require.Equal(t, VehicleAuditStatusPending, vehicle.Status)
	require.Equal(t, VehicleAuditStatusPending, vehicle.ReviewStatus)
	require.Equal(t, "audit", vehicle.Source)
	require.False(t, vehicle.CanDelete)
	require.Equal(t, originalPlate, repo.vehicles[2001][0].PlateNo)
	require.Equal(t, "BYD", repo.vehicles[2001][0].Brand)
	require.Len(t, repo.vehicleAudits[2001], 1)
	require.Equal(t, int64(4001), repo.vehicleAudits[2001][0].ID)
	require.Equal(t, updatedPlate, repo.vehicleAudits[2001][0].PlateNumber)
	require.Equal(t, 5, repo.vehicleAudits[2001][0].Seats)

	_, err = uc.UpdateVehicle(context.Background(), UpdateVehicleCommand{
		DriverID:  2002,
		VehicleID: 4001,
		PlateNo:   "沪A33333",
		Brand:     "Hack",
		Seats:     4,
	})
	require.ErrorIs(t, err, ErrVehicleNotFound)
	require.Equal(t, originalPlate, repo.vehicles[2001][0].PlateNo)
}

func TestDeleteVehicleRequiresDriverOwnershipAndSyncsAuditDeletedStatus(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.vehicles[2001] = []DriverVehicle{{
		ID:       4001,
		DriverID: 2001,
		PlateNo:  "沪A44444",
		Brand:    "BYD",
		Seats:    4,
		Status:   VehicleStatusActive,
	}}
	repo.vehicleAudits[2001] = []DriverVehicleAudit{{
		ID:          4001,
		DriverID:    2001,
		PlateNumber: "沪A44444",
		Brand:       "BYD",
		Status:      VehicleAuditStatusApproved,
	}}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	err = uc.DeleteVehicle(context.Background(), DeleteVehicleCommand{DriverID: 2001, VehicleID: 4001})

	require.NoError(t, err)
	require.Equal(t, VehicleStatusInactive, repo.vehicles[2001][0].Status)
	require.Equal(t, VehicleAuditStatusDriverDeleted, repo.vehicleAudits[2001][0].Status)
	vehicles, err := uc.ListVehicles(context.Background(), 2001)
	require.NoError(t, err)
	require.Empty(t, vehicles)

	err = uc.DeleteVehicle(context.Background(), DeleteVehicleCommand{DriverID: 2002, VehicleID: 4001})
	require.ErrorIs(t, err, ErrVehicleNotFound)
}

func TestAckMessageMarksDriverMessageDelivered(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.messages[2001] = []DriverMessage{
		{ID: 9001, Topic: "vehicle.audit", Title: "车辆认证已驳回"},
	}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	items, err := uc.ListMessages(context.Background(), ListMessagesQuery{DriverID: 2001})
	require.NoError(t, err)
	require.Len(t, items, 1)

	err = uc.AckMessage(context.Background(), AckMessageCommand{DriverID: 2001, MessageID: 9001})
	require.NoError(t, err)
	items, err = uc.ListMessages(context.Background(), ListMessagesQuery{DriverID: 2001})
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestUpdateDriverCanChangeServiceStatus(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.profiles[2001] = &DriverProfile{ID: 2001, Name: "Driver", ServiceStatus: DriverServiceStatusOffline}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	profile, err := uc.UpdateDriver(context.Background(), UpdateDriverCommand{
		ID:            2001,
		Name:          "  Driver Lee  ",
		ServiceStatus: DriverServiceStatusOnline,
	})

	require.NoError(t, err)
	require.Equal(t, "Driver Lee", profile.Name)
	require.Equal(t, DriverServiceStatusOnline, profile.ServiceStatus)
	require.Equal(t, DriverServiceStatusOnline, repo.profiles[2001].ServiceStatus)
}

func TestUpdateDriverRejectsInvalidServiceStatus(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	repo.profiles[2001] = &DriverProfile{ID: 2001, Name: "Driver", ServiceStatus: DriverServiceStatusOffline}
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	_, err = uc.UpdateDriver(context.Background(), UpdateDriverCommand{
		ID:            2001,
		ServiceStatus: 99,
	})

	require.ErrorIs(t, err, ErrInvalidDriver)
	require.Equal(t, DriverServiceStatusOffline, repo.profiles[2001].ServiceStatus)
}

func TestReportLocationRejectsInvalidCoordinates(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	uc := NewDriverUsecase(node, zap.NewNop(), newFakeDriverRepo())

	_, err = uc.ReportLocation(context.Background(), ReportLocationCommand{
		DriverID:   2001,
		OrderID:    3001,
		Latitude:   91,
		Longitude:  120.1,
		ReportedAt: time.Now(),
	})

	require.ErrorIs(t, err, ErrInvalidDriverLocation)
}

func TestReportLocationRejectsMissingOrderID(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	_, err = uc.ReportLocation(context.Background(), ReportLocationCommand{
		DriverID:  2001,
		Latitude:  30.22,
		Longitude: 120.11,
	})

	require.ErrorIs(t, err, ErrInvalidDriverLocation)
	require.Empty(t, repo.locations)
}

func TestReportLocationDefaultsReportedAt(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	repo := newFakeDriverRepo()
	uc := NewDriverUsecase(node, zap.NewNop(), repo)

	point, err := uc.ReportLocation(context.Background(), ReportLocationCommand{
		DriverID:  2001,
		OrderID:   3001,
		Latitude:  30.22,
		Longitude: 120.11,
	})

	require.NoError(t, err)
	require.NotZero(t, point.ID)
	require.False(t, point.ReportedAt.IsZero())
	require.Len(t, repo.locations, 1)
}

func TestReplayTrackRejectsMissingOrderID(t *testing.T) {
	node, err := snowflake.NewNode(4)
	require.NoError(t, err)
	uc := NewDriverUsecase(node, zap.NewNop(), newFakeDriverRepo())

	_, err = uc.ReplayTrack(context.Background(), TrackReplayQuery{DriverID: 2001})

	require.ErrorIs(t, err, ErrInvalidDriverLocation)
}
