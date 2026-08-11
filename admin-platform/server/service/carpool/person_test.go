package carpool

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

func TestPersonServiceRulesRolesAndImport(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.PersonProfile{},
		&carpoolModel.PersonRole{},
		&carpoolModel.PersonImportBatch{},
		&carpoolModel.PersonImportError{},
	))
	global.GVA_DB = db

	service := PersonService{}
	ctx := context.Background()

	person, err := service.CreatePerson(ctx, carpoolReq.PersonPayload{
		PersonType:      "driver",
		Name:            "Driver A",
		Phone:           "13800000001",
		Email:           "driver@example.com",
		IDCardNo:        "110101199001011234",
		DriverLicenseNo: "DL10001",
		VehicleNo:       "京A10001",
		VehicleType:     "SUV",
		RegisterDate:    time.Now().Format("2006-01-02"),
		Roles:           []string{"carpool_driver", "shuttle_driver"},
	})
	require.NoError(t, err)
	require.Equal(t, "138****0001", person.PhoneMasked)
	require.Equal(t, "110101********1234", person.IDCardMasked)
	require.Len(t, person.Roles, 2)

	_, err = service.CreatePerson(ctx, carpoolReq.PersonPayload{
		PersonType:   "driver",
		Name:         "Duplicate",
		Phone:        "13800000001",
		IDCardNo:     "110101199001019999",
		RegisterDate: time.Now().Format("2006-01-02"),
		Roles:        []string{"carpool_driver"},
	})
	require.Error(t, err)

	detail, err := service.AssignRoles(ctx, carpoolReq.PersonRoleAssign{
		PersonID: person.ID,
		Roles:    []string{"dispatcher", "ticket_checker"},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"dispatcher", "ticket_checker"}, roleCodes(detail.Roles))

	require.NoError(t, service.BatchUpdateStatus(ctx, carpoolReq.PersonBatchStatus{
		IDs:    []uint64{person.ID},
		Status: "disabled",
		Reason: "risk control",
	}))
	disabled, err := service.GetPersonDetail(ctx, person.ID)
	require.NoError(t, err)
	require.Equal(t, "disabled", disabled.Status)

	preview, err := service.PreviewImport(ctx, carpoolReq.PersonImportPayload{
		SourceType: "json",
		Operator:   "admin",
		Rows: []carpoolReq.PersonPayload{
			{PersonType: "passenger", Name: "Passenger A", Phone: "13900000001", IDCardNo: "110101199201011234", RegisterDate: "2026-07-29", Roles: []string{"passenger"}},
			{PersonType: "passenger", Name: "Bad Phone", Phone: "123", IDCardNo: "110101199201019999", RegisterDate: "2026-07-29", Roles: []string{"passenger"}},
			{PersonType: "driver", Name: "Duplicate Existing", Phone: "13800000001", IDCardNo: "110101199301011234", RegisterDate: "2026-07-29", Roles: []string{"carpool_driver"}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 3, preview.Total)
	require.Equal(t, 1, preview.SuccessCount)
	require.Equal(t, 2, preview.ErrorCount)

	_, err = service.CommitImport(ctx, carpoolReq.PersonImportPayload{
		SourceType: "json",
		Operator:   "admin",
		Rows: []carpoolReq.PersonPayload{
			{PersonType: "passenger", Name: "Passenger B", Phone: "13900000002", IDCardNo: "110101199401011234", RegisterDate: "2026-07-29", Roles: []string{"passenger"}},
			{PersonType: "passenger", Name: "Invalid", Phone: "13900000003", IDCardNo: "bad-card", RegisterDate: "2026-07-29", Roles: []string{"passenger"}},
		},
	})
	require.Error(t, err)

	list, total, err := service.ListPersons(ctx, carpoolReq.PersonSearch{PersonType: "passenger"})
	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, list)
}

func TestPersonServiceBatchDeleteDriversSoftDeletesOnlyDrivers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.PersonProfile{},
		&carpoolModel.PersonRole{},
		&carpoolModel.PersonImportBatch{},
		&carpoolModel.PersonImportError{},
	))
	global.GVA_DB = db

	service := PersonService{}
	ctx := context.Background()

	driver, err := service.CreatePerson(ctx, carpoolReq.PersonPayload{
		PersonType:   "driver",
		Name:         "Driver Delete",
		Phone:        "13800000011",
		IDCardNo:     "110101199001011235",
		RegisterDate: time.Now().Format("2006-01-02"),
		Roles:        []string{"carpool_driver"},
	})
	require.NoError(t, err)
	passenger, err := service.CreatePerson(ctx, carpoolReq.PersonPayload{
		PersonType:   "passenger",
		Name:         "Passenger Keep",
		Phone:        "13900000011",
		IDCardNo:     "110101199201011235",
		RegisterDate: time.Now().Format("2006-01-02"),
		Roles:        []string{"passenger"},
	})
	require.NoError(t, err)

	err = service.BatchDeleteDrivers(ctx, carpoolReq.PersonBatchDelete{
		IDs:    []uint64{driver.ID, passenger.ID},
		Reason: "offboard",
	})
	require.ErrorContains(t, err, "only driver can be batch deleted")

	err = service.BatchDeleteDrivers(ctx, carpoolReq.PersonBatchDelete{
		IDs:    []uint64{driver.ID},
		Reason: "offboard",
	})
	require.NoError(t, err)

	var deleted carpoolModel.PersonProfile
	require.NoError(t, db.First(&deleted, driver.ID).Error)
	require.Equal(t, "deleted", deleted.Status)

	list, total, err := service.ListPersons(ctx, carpoolReq.PersonSearch{PersonType: "driver"})
	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, list)

	stillThere, err := service.GetPersonDetail(ctx, passenger.ID)
	require.NoError(t, err)
	require.Equal(t, "passenger", stillThere.PersonType)
}

func TestPersonServiceStatsCountsPeopleByTypeAndStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&carpoolModel.PersonProfile{}, &carpoolModel.PersonRole{}))
	global.GVA_DB = db

	service := PersonService{}
	ctx := context.Background()
	registerDate := time.Now()
	require.NoError(t, db.Create(&[]carpoolModel.PersonProfile{
		{PersonNo: "STAT-D-1", PersonType: "driver", Name: "Driver Enabled", PhoneHash: "pd1", IDCardHash: "id1", RegisterDate: registerDate, Status: "enabled"},
		{PersonNo: "STAT-D-2", PersonType: "driver", Name: "Driver Disabled", PhoneHash: "pd2", IDCardHash: "id2", RegisterDate: registerDate, Status: "disabled"},
		{PersonNo: "STAT-P-1", PersonType: "passenger", Name: "Passenger Enabled", PhoneHash: "pp1", IDCardHash: "ip1", RegisterDate: registerDate, Status: "enabled"},
	}).Error)

	driverStats, err := service.GetStats(ctx, "driver")
	require.NoError(t, err)
	require.EqualValues(t, 2, driverStats.Total)
	require.EqualValues(t, 1, driverStats.Enabled)
	require.EqualValues(t, 1, driverStats.Disabled)

	passengerStats, err := service.GetStats(ctx, "passenger")
	require.NoError(t, err)
	require.EqualValues(t, 1, passengerStats.Total)
	require.EqualValues(t, 1, passengerStats.Enabled)
	require.EqualValues(t, 0, passengerStats.Disabled)
}

func roleCodes(roles []carpoolModel.PersonRole) []string {
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.RoleCode)
	}
	return codes
}
