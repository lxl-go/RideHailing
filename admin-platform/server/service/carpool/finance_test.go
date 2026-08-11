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

func TestFinanceServiceUsesRealCarpoolPaymentsAndOrderRefunds(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&testCarpoolTrip{},
		&testCarpoolOrder{},
		&testCarpoolPayment{},
		&carpoolModel.OrderRefund{},
	))
	global.GVA_DB = db

	now := time.Date(2026, 8, 4, 11, 0, 0, 0, time.Local)
	previousNow := financeNow
	financeNow = func() time.Time { return now }
	t.Cleanup(func() {
		financeNow = previousNow
	})
	require.NoError(t, db.Create(&testCarpoolTrip{ID: 7101, DriverID: 9101, StartLocation: "A", EndLocation: "B", DepartureTime: now, Status: "published"}).Error)
	require.NoError(t, db.Create(&testCarpoolOrder{ID: 8101, TripID: 7101, PassengerID: 9201, SeatsBooked: 1, TotalPrice: 66.66, Status: "accepted", AcceptedAt: &now, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&testCarpoolPayment{ID: 9101, OrderID: 8101, OutTradeNo: "PAY-REAL-001", TotalAmount: 66.66, Status: "paid", PaidAt: now, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&carpoolModel.OrderRefund{RefundNo: "RF-REAL-001", OrderNo: "8101", ServiceType: "carpool", PassengerID: 9201, RefundAmount: 16.66, Status: "approved", IdempotentKey: "refund-real-001"}).Error)

	service := FinanceService{}
	transactions, total, err := service.ListTransactions(context.Background(), carpoolReq.FinanceSearch{OrderNo: "8101"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "8101", transactions[0].OrderNo)
	require.Equal(t, 66.66, transactions[0].Amount)
	require.Equal(t, "alipay", transactions[0].PaymentMethod)

	refunds, refundTotal, err := service.ListRefunds(context.Background(), carpoolReq.FinanceSearch{OrderNo: "8101"})
	require.NoError(t, err)
	require.EqualValues(t, 1, refundTotal)
	require.Equal(t, "RF-REAL-001", refunds[0].RefundNo)

	summary, err := service.GetSummary(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 1, summary.TransactionCount)
	require.Equal(t, 66.66, summary.TotalAmount)
	require.Equal(t, 16.66, summary.RefundAmount)
	require.EqualValues(t, 0, summary.AbnormalCount)
	require.Equal(t, 66.66, summary.DriverIncomeDay)
	require.Equal(t, 66.66, summary.DriverIncomeWeek)
	require.Equal(t, 66.66, summary.DriverIncomeMonth)
	require.Equal(t, 66.66, summary.DriverIncomeYear)

	abnormal, abnormalTotal, err := service.ListAbnormalTransactions(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 0, abnormalTotal)
	require.Empty(t, abnormal)
}
