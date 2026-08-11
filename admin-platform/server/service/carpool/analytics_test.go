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

func TestAnalyticsServiceDashboardVolumeConversionAndRepurchase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&carpoolModel.OrderMain{}, &carpoolModel.PersonProfile{}))
	global.GVA_DB = db

	base := time.Date(2026, 7, 29, 10, 0, 0, 0, time.Local)
	require.NoError(t, db.Create(&[]carpoolModel.PersonProfile{
		{ID: 1001, PersonNo: "P1001", PersonType: "passenger", Name: "Passenger A", PhoneHash: "p1", IDCardHash: "i1", RegisterDate: base.AddDate(0, 0, -20), Status: "enabled"},
		{ID: 1002, PersonNo: "P1002", PersonType: "passenger", Name: "Passenger B", PhoneHash: "p2", IDCardHash: "i2", RegisterDate: base.AddDate(0, 0, -18), Status: "enabled"},
		{ID: 1003, PersonNo: "P1003", PersonType: "passenger", Name: "Passenger C", PhoneHash: "p3", IDCardHash: "i3", RegisterDate: base.AddDate(0, 0, -5), Status: "enabled"},
		{ID: 2001, PersonNo: "D2001", PersonType: "driver", Name: "Driver A", PhoneHash: "p4", IDCardHash: "i4", RegisterDate: base.AddDate(0, 0, -30), Status: "enabled"},
		{ID: 2002, PersonNo: "D2002", PersonType: "driver", Name: "Driver B", PhoneHash: "p5", IDCardHash: "i5", RegisterDate: base.AddDate(0, 0, -30), Status: "enabled"},
	}).Error)
	require.NoError(t, db.Create(&[]carpoolModel.OrderMain{
		{OrderNo: "A-1", ServiceType: "carpool", PassengerID: 1001, DriverID: 2001, RouteName: "R1", DepartTime: base.AddDate(0, 0, -12), ArrivalTime: base.AddDate(0, 0, -12).Add(time.Hour), Status: "completed", PayAmount: 80, CreatedAt: base.AddDate(0, 0, -12), UpdatedAt: base.AddDate(0, 0, -12)},
		{OrderNo: "A-2", ServiceType: "carpool", PassengerID: 1001, DriverID: 2001, RouteName: "R1", DepartTime: base.AddDate(0, 0, -2), ArrivalTime: base.AddDate(0, 0, -2).Add(time.Hour), Status: "completed", PayAmount: 120, CreatedAt: base.AddDate(0, 0, -2), UpdatedAt: base.AddDate(0, 0, -2)},
		{OrderNo: "B-1", ServiceType: "shuttle", PassengerID: 1002, DriverID: 2002, RouteName: "R2", DepartTime: base, ArrivalTime: base.Add(time.Hour), Status: "paid", PayAmount: 60, CreatedAt: base, UpdatedAt: base},
		{OrderNo: "C-1", ServiceType: "carpool", PassengerID: 1003, DriverID: 2002, RouteName: "R3", DepartTime: base, ArrivalTime: base.Add(time.Hour), Status: "cancelled", PayAmount: 50, CreatedAt: base, UpdatedAt: base},
	}).Error)

	service := AnalyticsService{}
	ctx := context.Background()

	dashboard, err := service.GetDashboard(ctx, carpoolReq.AnalyticsSearch{StartDate: "2026-07-01", EndDate: "2026-07-31"})
	require.NoError(t, err)
	require.EqualValues(t, 3, dashboard.MonthOrderCount)
	require.EqualValues(t, 2, dashboard.ActivePassengers)
	require.EqualValues(t, 2, dashboard.ActiveDrivers)
	require.Equal(t, 260.0, dashboard.MonthRevenue)
	require.InDelta(t, 66.67, dashboard.ConversionRate, 0.01)

	volume, err := service.GetOrderVolume(ctx, carpoolReq.AnalyticsSearch{Period: "day", StartDate: "2026-07-27", EndDate: "2026-07-29"})
	require.NoError(t, err)
	require.Equal(t, []string{"2026-07-27", "2026-07-28", "2026-07-29"}, volume.Categories)
	require.Equal(t, []int64{1, 0, 2}, volume.TotalOrders)
	require.Equal(t, []int64{1, 0, 1}, volume.ValidOrders)

	classification, err := service.GetOrderClassification(ctx, carpoolReq.AnalyticsSearch{StartDate: "2026-07-01", EndDate: "2026-07-31"})
	require.NoError(t, err)
	require.EqualValues(t, 3, classification.ValidOrders)
	require.EqualValues(t, 1, classification.InvalidOrders)
	require.EqualValues(t, 0, classification.CouponOrders)

	repurchase, err := service.GetRepurchase(ctx, carpoolReq.AnalyticsSearch{StartDate: "2026-07-01", EndDate: "2026-07-31"})
	require.NoError(t, err)
	require.EqualValues(t, 2, repurchase.FirstOrderUsers)
	require.EqualValues(t, 1, repurchase.RepurchaseUsers)
	require.Equal(t, 50.0, repurchase.RepurchaseRate)
	require.Equal(t, 10.0, repurchase.AvgRepurchaseDays)
}
