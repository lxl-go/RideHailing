package carpool

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"gorm.io/gorm"
)

const (
	orderStatusPaid      = "paid"
	orderStatusCompleted = "completed"
	orderStatusRefunding = "refunding"
	orderStatusRefunded  = "refunded"

	refundStatusPending  = "pending"
	refundStatusApproved = "approved"
	refundStatusRejected = "rejected"
	refundTypeAuto       = "auto"
	refundTypeManual     = "manual"
)

type OrderService struct{}

type OrderDetail struct {
	Order   carpoolModel.AdminOrderView       `json:"order"`
	Refunds []carpoolModel.OrderRefund        `json:"refunds"`
	History []carpoolModel.OrderStatusHistory `json:"history"`
}

type OrderOverview struct {
	TotalOrders     int64   `json:"totalOrders"`
	PaidOrders      int64   `json:"paidOrders"`
	CompletedOrders int64   `json:"completedOrders"`
	RefundingOrders int64   `json:"refundingOrders"`
	RefundedOrders  int64   `json:"refundedOrders"`
	CancelledOrders int64   `json:"cancelledOrders"`
	Revenue         float64 `json:"revenue"`
}

type OrderBatchRefundResult struct {
	Items []OrderBatchRefundItem `json:"items"`
}

type OrderBatchRefundItem struct {
	OrderNo string                    `json:"orderNo"`
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Refund  *carpoolModel.OrderRefund `json:"refund,omitempty"`
}

func (s *OrderService) ListOrders(ctx context.Context, search carpoolReq.OrderSearch) ([]carpoolModel.AdminOrderView, int64, error) {
	db := realOrderSearchQuery(ctx, search)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := limitOffset(search.Page, search.PageSize)
	var rows []adminOrderRow
	if err := realOrderRowsQuery(ctx, search).Order("ct.departure_time DESC, co.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return adminOrderViewsFromRows(rows), total, nil
}

func (s *OrderService) GetOverview(ctx context.Context, search carpoolReq.OrderSearch) (*OrderOverview, error) {
	base := realOrderSearchQuery(ctx, search)
	result := &OrderOverview{}

	if err := base.Count(&result.TotalOrders).Error; err != nil {
		return nil, err
	}
	if err := realOrderSearchQuery(ctx, search).Where("co.status = ?", orderStatusPaid).Count(&result.PaidOrders).Error; err != nil {
		return nil, err
	}
	if err := realOrderSearchQuery(ctx, search).Where("co.status = ?", orderStatusCompleted).Count(&result.CompletedOrders).Error; err != nil {
		return nil, err
	}
	if err := realOrderSearchQuery(ctx, search).Where("co.status = ?", orderStatusRefunding).Count(&result.RefundingOrders).Error; err != nil {
		return nil, err
	}
	if err := realOrderSearchQuery(ctx, search).Where("co.status = ?", orderStatusRefunded).Count(&result.RefundedOrders).Error; err != nil {
		return nil, err
	}
	if err := realOrderSearchQuery(ctx, search).Where("co.status IN ?", invalidOrderStatuses()).Count(&result.CancelledOrders).Error; err != nil {
		return nil, err
	}
	if err := realOrderSearchQuery(ctx, search).Where("co.status IN ?", validOrderStatuses()).Select("COALESCE(SUM(co.total_price),0)").Scan(&result.Revenue).Error; err != nil {
		return nil, err
	}
	result.Revenue = roundOrderAmount(result.Revenue)
	return result, nil
}

func (s *OrderService) GetOrderDetail(ctx context.Context, orderNo string) (*OrderDetail, error) {
	order, err := getRealAdminOrderByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	refunds, _, err := s.ListRefunds(ctx, carpoolReq.OrderRefundSearch{OrderNo: orderNo})
	if err != nil {
		return nil, err
	}
	history, err := s.GetStatusHistory(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: order, Refunds: refunds, History: history}, nil
}

func (s *OrderService) GetStatusHistory(ctx context.Context, orderNo string) ([]carpoolModel.OrderStatusHistory, error) {
	var history []carpoolModel.OrderStatusHistory
	err := global.GVA_DB.WithContext(ctx).Where("order_no = ?", orderNo).Order("created_at ASC, id ASC").Find(&history).Error
	return history, err
}

func (s *OrderService) ListRefunds(ctx context.Context, search carpoolReq.OrderRefundSearch) ([]carpoolModel.OrderRefund, int64, error) {
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
	limit, offset := limitOffset(search.Page, search.PageSize)
	var list []carpoolModel.OrderRefund
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *OrderService) ApplyRefund(ctx context.Context, req carpoolReq.OrderRefundApply) (*carpoolModel.OrderRefund, error) {
	var existing carpoolModel.OrderRefund
	if err := global.GVA_DB.WithContext(ctx).Where("idempotent_key = ?", req.IdempotentKey).First(&existing).Error; err == nil {
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var created carpoolModel.OrderRefund
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orderID, err := parsePositiveUintString(req.OrderNo, "orderNo")
		if err != nil {
			return err
		}
		var order carpoolModel.CarpoolOrderRecord
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
		}

		var trip carpoolModel.CarpoolTripRecord
		if err := tx.Where("id = ?", order.TripID).First(&trip).Error; err != nil {
			return err
		}
		rule := calculateRefund(order, trip.DepartureTime)
		refund := carpoolModel.OrderRefund{
			RefundNo:          nextRefundNo(),
			OrderNo:           strconv.FormatUint(order.ID, 10),
			ServiceType:       "carpool",
			PassengerID:       order.PassengerID,
			RefundAmount:      rule.refundAmount,
			CancelFee:         rule.cancelFee,
			Reason:            req.Reason,
			ReviewType:        rule.reviewType,
			Status:            rule.status,
			IdempotentKey:     req.IdempotentKey,
			EstimatedFinishAt: time.Now().Add(24 * time.Hour),
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}
		created = refund

		newStatus := orderStatusRefunding
		if rule.status == refundStatusApproved && rule.reviewType == refundTypeAuto {
			newStatus = orderStatusRefunded
		}
		if err := tx.Model(&carpoolModel.CarpoolOrderRecord{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"status": newStatus,
			}).Error; err != nil {
			return err
		}
		return tx.Create(&carpoolModel.OrderStatusHistory{
			OrderNo: refund.OrderNo, FromStatus: order.Status, ToStatus: newStatus, Operator: fallback(req.Operator, "system"), Reason: req.Reason,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *OrderService) ReviewRefund(ctx context.Context, req carpoolReq.OrderRefundReview) (*carpoolModel.OrderRefund, error) {
	var updated carpoolModel.OrderRefund
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var refund carpoolModel.OrderRefund
		if err := tx.Where("refund_no = ?", req.RefundNo).First(&refund).Error; err != nil {
			return err
		}
		orderID, err := parsePositiveUintString(refund.OrderNo, "orderNo")
		if err != nil {
			return err
		}
		var order carpoolModel.CarpoolOrderRecord
		if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
		}
		status := refundStatusRejected
		orderStatus := order.Status
		if req.Decision == "approved" {
			status = refundStatusApproved
			orderStatus = orderStatusRefunded
		}
		if err := tx.Model(&refund).Updates(map[string]interface{}{
			"status":        status,
			"reviewer":      req.Reviewer,
			"review_remark": req.Remark,
		}).Error; err != nil {
			return err
		}
		if orderStatus != order.Status {
			if err := tx.Model(&carpoolModel.CarpoolOrderRecord{}).Where("id = ?", order.ID).
				Updates(map[string]interface{}{"status": orderStatus}).Error; err != nil {
				return err
			}
			if err := tx.Create(&carpoolModel.OrderStatusHistory{
				OrderNo: refund.OrderNo, FromStatus: order.Status, ToStatus: orderStatus, Operator: req.Reviewer, Reason: req.Remark,
			}).Error; err != nil {
				return err
			}
		}
		updated = refund
		updated.Status = status
		updated.Reviewer = req.Reviewer
		updated.ReviewRemark = req.Remark
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *OrderService) BatchRefund(ctx context.Context, req carpoolReq.OrderBatchRefund) (*OrderBatchRefundResult, error) {
	if len(req.OrderNos) == 0 {
		return nil, errors.New("orderNos is required")
	}
	if len(req.OrderNos) > 100 {
		return nil, errors.New("batch refund limit exceeded")
	}
	result := &OrderBatchRefundResult{Items: make([]OrderBatchRefundItem, 0, len(req.OrderNos))}
	for i, orderNo := range req.OrderNos {
		refund, err := s.ApplyRefund(ctx, carpoolReq.OrderRefundApply{
			OrderNo: orderNo, Reason: req.Reason, Operator: req.Operator, IdempotentKey: fmt.Sprintf("%s:%d:%s", req.IdempotentSeed, i, orderNo),
		})
		item := OrderBatchRefundItem{OrderNo: orderNo, Success: err == nil, Refund: refund}
		if err != nil {
			item.Message = err.Error()
		} else {
			item.Message = "ok"
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *OrderService) ExportOrders(ctx context.Context) string {
	return fmt.Sprintf("ORDER-EXP-%d", time.Now().UnixNano())
}

func SeedOrderDefaults(db *gorm.DB) error {
	var count int64
	if err := db.Model(&carpoolModel.OrderMain{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	return db.Create(&[]carpoolModel.OrderMain{
		{ID: 41001, OrderNo: "OD202607290001", ServiceType: "shuttle", PassengerID: 1001, PassengerName: "Passenger A", PassengerPhone: "13800000001", DriverID: 2001, DriverName: "Driver A", DriverPhone: "13900000001", VehicleNo: "京A10001", RouteName: "Morning Shuttle", DepartTime: now.Add(3 * time.Hour), ArrivalTime: now.Add(4 * time.Hour), Status: orderStatusPaid, PayAmount: 100, Version: 1},
		{ID: 41002, OrderNo: "OD202607290002", ServiceType: "carpool", PassengerID: 1002, PassengerName: "Passenger B", PassengerPhone: "13800000002", DriverID: 2002, DriverName: "Driver B", DriverPhone: "13900000002", VehicleNo: "京A10002", RouteName: "Airport Carpool", DepartTime: now.Add(60 * time.Minute), ArrivalTime: now.Add(2 * time.Hour), Status: orderStatusPaid, PayAmount: 80, Version: 1},
		{ID: 41003, OrderNo: "OD202607290003", ServiceType: "carpool", PassengerID: 1003, PassengerName: "Passenger C", PassengerPhone: "13800000003", DriverID: 2003, DriverName: "Driver C", DriverPhone: "13900000003", VehicleNo: "京A10003", RouteName: "Railway Carpool", DepartTime: now.Add(-2 * time.Hour), ArrivalTime: now.Add(-1 * time.Hour), Status: orderStatusCompleted, PayAmount: 120, Version: 1},
	}).Error
}

type refundRule struct {
	refundAmount float64
	cancelFee    float64
	reviewType   string
	status       string
}

func calculateRefund(order carpoolModel.CarpoolOrderRecord, departTime time.Time) refundRule {
	payAmount := roundOrderAmount(order.TotalPrice)
	if order.Status == orderStatusCompleted {
		return refundRule{refundAmount: payAmount, reviewType: refundTypeManual, status: refundStatusPending}
	}
	minutes := time.Until(departTime).Minutes()
	if minutes > 120 {
		return refundRule{refundAmount: payAmount, reviewType: refundTypeAuto, status: refundStatusApproved}
	}
	if minutes >= 30 {
		fee := roundOrderAmount(payAmount * 0.1)
		return refundRule{refundAmount: roundOrderAmount(payAmount - fee), cancelFee: fee, reviewType: refundTypeAuto, status: refundStatusApproved}
	}
	return refundRule{refundAmount: payAmount, reviewType: refundTypeManual, status: refundStatusPending}
}

func nextRefundNo() string {
	return fmt.Sprintf("RF%d", time.Now().UnixNano())
}

func limitOffset(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return pageSize, (page - 1) * pageSize
}

func roundOrderAmount(v float64) float64 {
	return math.Round(v*100) / 100
}

func orderSearchQuery(ctx context.Context, search carpoolReq.OrderSearch) *gorm.DB {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.OrderMain{})
	if search.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+search.OrderNo+"%")
	}
	if search.ServiceType != "" {
		db = db.Where("service_type = ?", search.ServiceType)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if search.StartDate != "" {
		db = db.Where("depart_time >= ?", search.StartDate)
	}
	if search.EndDate != "" {
		db = db.Where("depart_time <= ?", search.EndDate)
	}
	return db
}

type adminOrderRow struct {
	ID            uint64
	PassengerID   uint64
	DriverID      uint64
	StartLocation string
	EndLocation   string
	DepartureTime time.Time
	PayAmount     float64
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func realOrderSearchQuery(ctx context.Context, search carpoolReq.OrderSearch) *gorm.DB {
	db := global.GVA_DB.WithContext(ctx).
		Table("carpool_order AS co").
		Joins("LEFT JOIN carpool_trip AS ct ON ct.id = co.trip_id").
		Joins("LEFT JOIN carpool_payment AS cp ON cp.order_id = co.id AND cp.status = ?", orderStatusPaid)
	if search.OrderNo != "" {
		keyword := "%" + strings.TrimSpace(search.OrderNo) + "%"
		db = db.Where("CAST(co.id AS CHAR) LIKE ? OR cp.out_trade_no LIKE ?", keyword, keyword)
	}
	if search.ServiceType != "" && search.ServiceType != "carpool" {
		db = db.Where("1 = 0")
	}
	if search.Status != "" {
		db = db.Where("co.status = ?", search.Status)
	}
	if search.StartDate != "" {
		db = db.Where("ct.departure_time >= ?", search.StartDate)
	}
	if search.EndDate != "" {
		db = db.Where("ct.departure_time <= ?", search.EndDate)
	}
	return db
}

func realOrderRowsQuery(ctx context.Context, search carpoolReq.OrderSearch) *gorm.DB {
	return realOrderSearchQuery(ctx, search).Select(`
		co.id AS id,
		co.passenger_id AS passenger_id,
		COALESCE(ct.driver_id, 0) AS driver_id,
		COALESCE(ct.start_location, '') AS start_location,
		COALESCE(ct.end_location, '') AS end_location,
		ct.departure_time AS departure_time,
		COALESCE(NULLIF(cp.total_amount, 0), co.total_price) AS pay_amount,
		co.status AS status,
		co.created_at AS created_at,
		co.updated_at AS updated_at`)
}

func getRealAdminOrderByNo(ctx context.Context, orderNo string) (carpoolModel.AdminOrderView, error) {
	orderID, err := parsePositiveUintString(orderNo, "orderNo")
	if err != nil {
		return carpoolModel.AdminOrderView{}, err
	}
	var row adminOrderRow
	err = realOrderRowsQuery(ctx, carpoolReq.OrderSearch{}).Where("co.id = ?", orderID).Scan(&row).Error
	if err != nil {
		return carpoolModel.AdminOrderView{}, err
	}
	if row.ID == 0 {
		return carpoolModel.AdminOrderView{}, gorm.ErrRecordNotFound
	}
	return adminOrderViewFromRow(row), nil
}

func adminOrderViewsFromRows(rows []adminOrderRow) []carpoolModel.AdminOrderView {
	list := make([]carpoolModel.AdminOrderView, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminOrderViewFromRow(row))
	}
	return list
}

func adminOrderViewFromRow(row adminOrderRow) carpoolModel.AdminOrderView {
	orderID := strconv.FormatUint(row.ID, 10)
	driverID := ""
	if row.DriverID > 0 {
		driverID = strconv.FormatUint(row.DriverID, 10)
	}
	return carpoolModel.AdminOrderView{
		ID:          orderID,
		OrderNo:     orderID,
		ServiceType: "carpool",
		PassengerID: strconv.FormatUint(row.PassengerID, 10),
		DriverID:    driverID,
		RouteName:   strings.Trim(strings.TrimSpace(row.StartLocation)+" - "+strings.TrimSpace(row.EndLocation), " -"),
		DepartTime:  row.DepartureTime,
		ArrivalTime: row.DepartureTime,
		Status:      row.Status,
		PayAmount:   roundOrderAmount(row.PayAmount),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Version:     1,
	}
}

func fallback(v, defaultValue string) string {
	if v == "" {
		return defaultValue
	}
	return v
}
