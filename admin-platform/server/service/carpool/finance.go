package carpool

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"gorm.io/gorm"
)

type FinanceService struct{}

type FinanceSummary struct {
	TransactionCount   int64   `json:"transactionCount"`
	TotalAmount        float64 `json:"totalAmount"`
	RefundAmount       float64 `json:"refundAmount"`
	AbnormalCount      int64   `json:"abnormalCount"`
	DriverIncomeDay    float64 `json:"driverIncomeDay"`
	DriverIncomeWeek   float64 `json:"driverIncomeWeek"`
	DriverIncomeMonth  float64 `json:"driverIncomeMonth"`
	DriverIncomeYear   float64 `json:"driverIncomeYear"`
}

var financeNow = time.Now

func (s *FinanceService) ListTransactions(ctx context.Context, search carpoolReq.FinanceSearch) ([]carpoolModel.FinanceTransaction, int64, error) {
	db := realFinanceTransactionQuery(ctx, search)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.FinanceTransaction
	if err := realFinanceTransactionQuery(ctx, search).Order("cp.created_at DESC, cp.id DESC").Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *FinanceService) ListRefunds(ctx context.Context, search carpoolReq.FinanceSearch) ([]carpoolModel.FinanceRefund, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.OrderRefund{})
	if search.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+search.OrderNo+"%")
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var refunds []carpoolModel.OrderRefund
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&refunds).Error; err != nil {
		return nil, 0, err
	}
	return financeRefundsFromOrderRefunds(refunds), total, nil
}

func (s *FinanceService) GetSummary(ctx context.Context) (*FinanceSummary, error) {
	var summary FinanceSummary
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.CarpoolPaymentRecord{}).Count(&summary.TransactionCount).Error; err != nil {
		return nil, err
	}
	if err := scanFinanceAmount(global.GVA_DB.WithContext(ctx).Model(&carpoolModel.CarpoolPaymentRecord{}), "COALESCE(SUM(total_amount),0)", &summary.TotalAmount); err != nil {
		return nil, err
	}
	if err := scanFinanceAmount(global.GVA_DB.WithContext(ctx).Model(&carpoolModel.OrderRefund{}), "COALESCE(SUM(refund_amount),0)", &summary.RefundAmount); err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.CarpoolPaymentRecord{}).Where("status <> ?", orderStatusPaid).Count(&summary.AbnormalCount).Error; err != nil {
		return nil, err
	}
	now := financeNow()
	ranges := map[string]*float64{
		"day":   &summary.DriverIncomeDay,
		"week":  &summary.DriverIncomeWeek,
		"month": &summary.DriverIncomeMonth,
		"year":  &summary.DriverIncomeYear,
	}
	for label, target := range ranges {
		start, end := driverIncomeRange(now, label)
		if err := driverIncomeAmount(ctx, start, end, target); err != nil {
			return nil, err
		}
	}
	summary.TotalAmount = roundFinance(summary.TotalAmount)
	summary.RefundAmount = roundFinance(summary.RefundAmount)
	summary.DriverIncomeDay = roundFinance(summary.DriverIncomeDay)
	summary.DriverIncomeWeek = roundFinance(summary.DriverIncomeWeek)
	summary.DriverIncomeMonth = roundFinance(summary.DriverIncomeMonth)
	summary.DriverIncomeYear = roundFinance(summary.DriverIncomeYear)
	return &summary, nil
}

func (s *FinanceService) ListAbnormalTransactions(ctx context.Context) ([]carpoolModel.FinanceTransaction, int64, error) {
	var list []carpoolModel.FinanceTransaction
	var total int64
	db := realFinanceTransactionQuery(ctx, carpoolReq.FinanceSearch{}).Where("cp.status <> ?", orderStatusPaid)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("cp.created_at DESC, cp.id DESC").Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *FinanceService) ExportTaskID(ctx context.Context) string {
	return fmt.Sprintf("FIN-EXP-%d", time.Now().UnixNano())
}

func roundFinance(v float64) float64 {
	return math.Round(v*100) / 100
}

func realFinanceTransactionQuery(ctx context.Context, search carpoolReq.FinanceSearch) *gorm.DB {
	db := global.GVA_DB.WithContext(ctx).
		Table("carpool_payment AS cp").
		Joins("JOIN carpool_order AS co ON co.id = cp.order_id").
		Joins("LEFT JOIN carpool_trip AS ct ON ct.id = co.trip_id").
		Select(`
			cp.id AS id,
			CAST(co.id AS CHAR) AS order_no,
			COALESCE(ct.driver_id, 0) AS driver_id,
			co.passenger_id AS passenger_id,
			cp.total_amount AS amount,
			'alipay' AS payment_method,
			CASE WHEN cp.status = 'paid' THEN 'success' ELSE cp.status END AS status,
			CASE WHEN cp.status = 'paid' THEN 0 ELSE 1 END AS abnormal,
			CASE WHEN cp.status = 'paid' THEN '' ELSE '支付状态未成功' END AS abnormal_reason,
			cp.created_at AS created_at,
			cp.updated_at AS updated_at`)
	if search.OrderNo != "" {
		keyword := "%" + strings.TrimSpace(search.OrderNo) + "%"
		db = db.Where("CAST(co.id AS CHAR) LIKE ? OR cp.out_trade_no LIKE ?", keyword, keyword)
	}
	if search.Status != "" {
		status := strings.TrimSpace(search.Status)
		if status == "success" {
			status = orderStatusPaid
		}
		db = db.Where("cp.status = ?", status)
	}
	return db
}

func scanFinanceAmount(db *gorm.DB, expression string, target *float64) error {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	if err := db.Select(expression + " AS amount").Scan(&row).Error; err != nil {
		return err
	}
	*target = row.Amount
	return nil
}

func driverIncomeAmount(ctx context.Context, start, end time.Time, target *float64) error {
	db := global.GVA_DB.WithContext(ctx).Table("carpool_order AS co").
		Joins("JOIN carpool_trip AS ct ON ct.id = co.trip_id").
		Where("co.accepted_at >= ? AND co.accepted_at < ?", start, end).
		Where("(co.status IN ('accepted','picking_up','delivering','completed') OR CAST(co.status AS CHAR) IN ('1','5','6','2'))")
	return scanFinanceAmount(db, "COALESCE(SUM(co.total_price),0)", target)
}

func driverIncomeRange(now time.Time, label string) (time.Time, time.Time) {
	loc := now.Location()
	switch label {
	case "day":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 0, 1)
	case "week":
		offset := int(now.Weekday())
		if offset == 0 {
			offset = 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-offset+1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 0, 7)
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, 0)
	default:
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(1, 0, 0)
	}
}

func financeRefundsFromOrderRefunds(refunds []carpoolModel.OrderRefund) []carpoolModel.FinanceRefund {
	list := make([]carpoolModel.FinanceRefund, 0, len(refunds))
	for _, refund := range refunds {
		list = append(list, carpoolModel.FinanceRefund{
			ID:           refund.ID,
			OrderNo:      refund.OrderNo,
			RefundNo:     refund.RefundNo,
			RefundAmount: refund.RefundAmount,
			Status:       refund.Status,
			CreatedAt:    refund.CreatedAt,
			UpdatedAt:    refund.UpdatedAt,
		})
	}
	return list
}
