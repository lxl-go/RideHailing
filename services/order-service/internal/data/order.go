package data

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ride-hailing/services/order-service/internal/biz"
)

type orderModel struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	TripID       int64      `gorm:"column:trip_id;type:bigint;not null;index:idx_trip_status"`
	PassengerID  int64      `gorm:"column:passenger_id;type:bigint;not null;index:idx_passenger_status"`
	SeatsBooked  int        `gorm:"column:seats_booked;type:int;not null"`
	TotalPrice   float64    `gorm:"column:total_price;type:decimal(10,2);not null"`
	Status       int        `gorm:"column:status;type:tinyint;not null;default:0"`
	AcceptedAt   *time.Time `gorm:"column:accepted_at;index:idx_order_accepted_at"`
	RejectReason string     `gorm:"column:reject_reason;type:varchar(255);not null;default:''"`
	RejectedAt   *time.Time `gorm:"column:rejected_at"`
	RefundAmount float64    `gorm:"column:refund_amount;type:decimal(10,2);not null;default:0"`
	RefundedAt   *time.Time `gorm:"column:refunded_at"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (orderModel) TableName() string {
	return "carpool_order"
}

type paymentModel struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	OrderID       int64      `gorm:"column:order_id;type:bigint;not null;uniqueIndex:uk_order_payment"`
	PassengerID   int64      `gorm:"column:passenger_id;type:bigint;not null;index:idx_payment_passenger"`
	OutTradeNo    string     `gorm:"column:out_trade_no;type:varchar(64);not null;uniqueIndex:uk_payment_out_trade_no"`
	AlipayTradeNo string     `gorm:"column:alipay_trade_no;type:varchar(64)"`
	Channel       string     `gorm:"column:channel;type:varchar(32);not null"`
	TotalAmount   string     `gorm:"column:total_amount;type:decimal(10,2);not null"`
	Status        int        `gorm:"column:status;type:tinyint;not null;default:0;index:idx_payment_status"`
	PaidAt        *time.Time `gorm:"column:paid_at"`
	NotifyPayload string     `gorm:"column:notify_payload;type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (paymentModel) TableName() string {
	return "carpool_payment"
}

type orderIdempotencyModel struct {
	ID        uint   `gorm:"column:id;primaryKey;autoIncrement"`
	BizKey    string `gorm:"column:biz_key;type:varchar(128);not null;uniqueIndex:uk_order_idempotency_biz_key"`
	OrderID   int64  `gorm:"column:order_id;type:bigint;not null;index:idx_order_idempotency_order"`
	ActorID   int64  `gorm:"column:actor_id;type:bigint;not null;index:idx_order_idempotency_actor"`
	Action    string `gorm:"column:action;type:varchar(32);not null"`
	Status    int    `gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (orderIdempotencyModel) TableName() string {
	return "carpool_order_idempotency"
}

type tripModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID       int64     `gorm:"column:driver_id;type:bigint;not null;index:idx_driver_status"`
	Origin         string    `gorm:"column:origin;type:varchar(128);not null"`
	Destination    string    `gorm:"column:destination;type:varchar(128);not null"`
	DepartTime     time.Time `gorm:"column:depart_time;not null;index:idx_depart_status"`
	SeatsAvailable int       `gorm:"column:seats_available;type:int;not null"`
	Price          float64   `gorm:"column:price;type:decimal(10,2);not null"`
	Status         int       `gorm:"column:status;type:tinyint;not null;default:1"`
	IsDeleted      bool      `gorm:"column:is_deleted;not null;default:false"`
}

func (tripModel) TableName() string {
	return "carpool_trip"
}

type OrderRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewOrderRepo(db *gorm.DB, log *zap.Logger) *OrderRepo {
	return &OrderRepo{db: db, log: log}
}

func (r *OrderRepo) GetTripForOrder(ctx context.Context, id int64) (*biz.TripSnapshot, error) {
	var m tripModel
	if err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", id, false).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrTripNotFound
		}
		return nil, err
	}
	return tripToDomain(&m), nil
}

func (r *OrderRepo) CreateAtomic(ctx context.Context, order *biz.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&tripModel{}).
			Where("id = ? AND status = ? AND is_deleted = ? AND depart_time > ? AND seats_available >= ?", order.TripID, biz.TripStatusRecruiting, false, time.Now(), order.SeatsBooked).
			UpdateColumn("seats_available", gorm.Expr("seats_available - ?", order.SeatsBooked))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return biz.ErrInsufficientSeats
		}
		return tx.Create(orderToRecord(order)).Error
	})
}

func (r *OrderRepo) GetByID(ctx context.Context, id int64) (*biz.Order, error) {
	var m orderModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, err
	}
	return orderToDomain(&m), nil
}

func (r *OrderRepo) ListByPassenger(ctx context.Context, passengerID int64, status int, page, pageSize int) ([]biz.Order, int64, error) {
	var models []orderModel
	var total int64
	query := r.db.WithContext(ctx).Model(&orderModel{}).Where("passenger_id = ?", passengerID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return ordersToDomain(models), total, nil
}

func (r *OrderRepo) ListByDriver(ctx context.Context, driverID int64, status int, page, pageSize int) ([]biz.Order, int64, error) {
	var rows []driverOrderRow
	var total int64
	activeStatuses := []int{biz.OrderStatusAccepted, biz.OrderStatusPickingUp, biz.OrderStatusDelivering}
	query := r.db.WithContext(ctx).Table("carpool_order AS co").
		Joins("JOIN carpool_trip AS ct ON ct.id = co.trip_id").
		Where("ct.driver_id = ?", driverID)
	if status >= 0 {
		query = query.Where("co.status = ?", status)
	} else {
		query = query.Where("co.status IN ?", activeStatuses)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Select(`
			co.id AS id,
			co.trip_id AS trip_id,
			co.passenger_id AS passenger_id,
			ct.driver_id AS driver_id,
			ct.origin AS origin,
			ct.destination AS destination,
			ct.depart_time AS depart_time,
			co.seats_booked AS seats_booked,
			co.total_price AS total_price,
			co.status AS status,
			co.accepted_at AS accepted_at,
			co.reject_reason AS reject_reason,
			co.rejected_at AS rejected_at,
			co.refund_amount AS refund_amount,
			co.refunded_at AS refunded_at,
			co.created_at AS created_at,
			co.updated_at AS updated_at`).
		Order("co.accepted_at DESC, co.created_at DESC, co.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return driverOrderRowsToDomain(rows), total, nil
}

func (r *OrderRepo) ListPendingByTrip(ctx context.Context, tripID int64, page, pageSize int) ([]biz.Order, int64, error) {
	var models []orderModel
	var total int64
	query := r.db.WithContext(ctx).Model(&orderModel{}).Where("trip_id = ? AND status = ?", tripID, biz.OrderStatusPaid)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return ordersToDomain(models), total, nil
}

func (r *OrderRepo) ListPendingByDriver(ctx context.Context, driverID int64, page, pageSize int) ([]biz.Order, int64, error) {
	var models []orderModel
	var total int64
	query := r.db.WithContext(ctx).Table("carpool_order").
		Joins("JOIN carpool_trip ON carpool_trip.id = carpool_order.trip_id").
		Where("carpool_trip.driver_id = ? AND carpool_order.status = ?", driverID, biz.OrderStatusPaid)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("carpool_order.created_at ASC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return ordersToDomain(models), total, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int64, status int) error {
	result := r.db.WithContext(ctx).Model(&orderModel{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepo) IncrementTripSeats(ctx context.Context, tripID int64, seats int) error {
	result := r.db.WithContext(ctx).Model(&tripModel{}).Where("id = ?", tripID).
		UpdateColumn("seats_available", gorm.Expr("seats_available + ?", seats))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrTripNotFound
	}
	return nil
}

func (r *OrderRepo) ApplyOrderTransition(ctx context.Context, transition biz.OrderTransition) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		bizKey := orderTransitionBizKey(transition)
		if bizKey != "" {
			if err := tx.Create(&orderIdempotencyModel{
				BizKey:  bizKey,
				OrderID: transition.OrderID,
				ActorID: transition.ActorID,
				Action:  transition.Action,
				Status:  1,
			}).Error; err != nil {
				var existing orderIdempotencyModel
				if queryErr := tx.Where("biz_key = ?", bizKey).First(&existing).Error; queryErr == nil {
					return nil
				}
				return err
			}
		}

		var locked orderModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", transition.OrderID).First(&locked).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return biz.ErrOrderNotFound
			}
			return err
		}
		if !bizStatusIn(locked.Status, transition.FromStatuses) {
			return biz.TransitionStatusError(transition.Action)
		}
		updates := map[string]any{"status": transition.ToStatus}
		if transition.AcceptedAt != nil {
			updates["accepted_at"] = transition.AcceptedAt
		}
		if transition.RejectReason != "" {
			updates["reject_reason"] = transition.RejectReason
		}
		if transition.RejectedAt != nil {
			updates["rejected_at"] = transition.RejectedAt
		}
		if transition.RefundAmount > 0 {
			updates["refund_amount"] = transition.RefundAmount
		}
		if transition.RefundedAt != nil {
			updates["refunded_at"] = transition.RefundedAt
		}
		result := tx.Model(&orderModel{}).Where("id = ?", transition.OrderID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return biz.TransitionStatusError(transition.Action)
		}
		if transition.RestoreTripID > 0 && transition.RestoreSeats > 0 {
			seatResult := tx.Model(&tripModel{}).Where("id = ?", transition.RestoreTripID).
				UpdateColumn("seats_available", gorm.Expr("seats_available + ?", transition.RestoreSeats))
			if seatResult.Error != nil {
				return seatResult.Error
			}
			if seatResult.RowsAffected == 0 {
				return biz.ErrTripNotFound
			}
		}
		if transition.RefundPayment {
			paymentResult := tx.Model(&paymentModel{}).
				Where("order_id = ? AND status = ?", transition.OrderID, biz.PaymentStatusPaid).
				Update("status", biz.PaymentStatusRefunded)
			if paymentResult.Error != nil {
				return paymentResult.Error
			}
		}
		return nil
	})
}

func (r *OrderRepo) ListDriverIncome(ctx context.Context, driverID int64, start, end time.Time, page, pageSize int) ([]biz.DriverIncomeRecord, int64, float64, error) {
	var total int64
	var totalAmount float64
	base := r.db.WithContext(ctx).Table("carpool_order AS co").
		Joins("JOIN carpool_trip AS ct ON ct.id = co.trip_id").
		Where("ct.driver_id = ? AND co.accepted_at >= ? AND co.accepted_at < ?", driverID, start, end).
		Where("co.status IN ?", []int{biz.OrderStatusAccepted, biz.OrderStatusPickingUp, biz.OrderStatusDelivering, biz.OrderStatusCompleted})
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}
	if err := base.Select("COALESCE(SUM(co.total_price),0)").Scan(&totalAmount).Error; err != nil {
		return nil, 0, 0, err
	}
	var records []driverIncomeRow
	if err := base.Select(`
			co.id AS order_id,
			co.passenger_id AS passenger_id,
			co.trip_id AS trip_id,
			ct.origin AS origin,
			ct.destination AS destination,
			co.total_price AS amount,
			co.status AS status,
			co.accepted_at AS accepted_at`).
		Order("co.accepted_at DESC, co.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&records).Error; err != nil {
		return nil, 0, 0, err
	}
	return driverIncomeRowsToDomain(records), total, totalAmount, nil
}

func (r *OrderRepo) CreatePayment(ctx context.Context, payment *biz.Payment) error {
	return r.db.WithContext(ctx).Create(paymentToRecord(payment)).Error
}

func (r *OrderRepo) GetPaymentByOrderID(ctx context.Context, orderID int64) (*biz.Payment, error) {
	var m paymentModel
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPaymentNotFound
		}
		return nil, err
	}
	return paymentToDomain(&m), nil
}

func (r *OrderRepo) GetPaymentByOutTradeNo(ctx context.Context, outTradeNo string) (*biz.Payment, error) {
	var m paymentModel
	if err := r.db.WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPaymentNotFound
		}
		return nil, err
	}
	return paymentToDomain(&m), nil
}

func (r *OrderRepo) MarkPaymentPaid(ctx context.Context, outTradeNo, alipayTradeNo string, notifyPayload string) (*biz.Payment, bool, error) {
	var updated paymentModel
	duplicated := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payment paymentModel
		if err := tx.Where("out_trade_no = ?", outTradeNo).First(&payment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return biz.ErrPaymentNotFound
			}
			return err
		}
		if payment.Status == biz.PaymentStatusPaid {
			duplicated = true
			updated = payment
			return nil
		}

		now := time.Now()
		paymentUpdate := tx.Model(&paymentModel{}).
			Where("out_trade_no = ? AND status = ?", outTradeNo, biz.PaymentStatusPending).
			Updates(map[string]any{
				"alipay_trade_no": alipayTradeNo,
				"status":          biz.PaymentStatusPaid,
				"paid_at":         &now,
				"notify_payload":  notifyPayload,
			})
		if paymentUpdate.Error != nil {
			return paymentUpdate.Error
		}
		if paymentUpdate.RowsAffected == 0 {
			if err := tx.Where("out_trade_no = ?", outTradeNo).First(&updated).Error; err != nil {
				return err
			}
			if updated.Status == biz.PaymentStatusPaid {
				duplicated = true
				return nil
			}
			return biz.ErrInvalidPayment
		}
		orderUpdate := tx.Model(&orderModel{}).
			Where("id = ?", payment.OrderID).
			Update("status", biz.OrderStatusPaid)
		if orderUpdate.Error != nil {
			return orderUpdate.Error
		}
		if orderUpdate.RowsAffected == 0 {
			return biz.ErrOrderNotFound
		}
		if err := tx.Where("out_trade_no = ?", outTradeNo).First(&updated).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return paymentToDomain(&updated), duplicated, nil
}

func ordersToDomain(models []orderModel) []biz.Order {
	items := make([]biz.Order, len(models))
	for i := range models {
		items[i] = *orderToDomain(&models[i])
	}
	return items
}

type driverIncomeRow struct {
	OrderID     int64
	PassengerID int64
	TripID      int64
	Origin      string
	Destination string
	Amount      float64
	Status      int
	AcceptedAt  time.Time
}

type driverOrderRow struct {
	ID           int64
	TripID       int64
	PassengerID  int64
	DriverID     int64
	Origin       string
	Destination  string
	DepartTime   time.Time
	SeatsBooked  int
	TotalPrice   float64
	Status       int
	AcceptedAt   *time.Time
	RejectReason string
	RejectedAt   *time.Time
	RefundAmount float64
	RefundedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func driverOrderRowsToDomain(rows []driverOrderRow) []biz.Order {
	items := make([]biz.Order, len(rows))
	for i := range rows {
		items[i] = biz.Order{
			ID:           rows[i].ID,
			TripID:       rows[i].TripID,
			PassengerID:  rows[i].PassengerID,
			DriverID:     rows[i].DriverID,
			Origin:       rows[i].Origin,
			Destination:  rows[i].Destination,
			DepartTime:   rows[i].DepartTime,
			SeatsBooked:  rows[i].SeatsBooked,
			TotalPrice:   rows[i].TotalPrice,
			Status:       rows[i].Status,
			AcceptedAt:   rows[i].AcceptedAt,
			RejectReason: rows[i].RejectReason,
			RejectedAt:   rows[i].RejectedAt,
			RefundAmount: rows[i].RefundAmount,
			RefundedAt:   rows[i].RefundedAt,
			CreatedAt:    rows[i].CreatedAt,
			UpdatedAt:    rows[i].UpdatedAt,
		}
	}
	return items
}

func driverIncomeRowsToDomain(rows []driverIncomeRow) []biz.DriverIncomeRecord {
	items := make([]biz.DriverIncomeRecord, len(rows))
	for i := range rows {
		items[i] = biz.DriverIncomeRecord{
			OrderID:     rows[i].OrderID,
			PassengerID: rows[i].PassengerID,
			TripID:      rows[i].TripID,
			Origin:      rows[i].Origin,
			Destination: rows[i].Destination,
			Amount:      rows[i].Amount,
			Status:      rows[i].Status,
			AcceptedAt:  rows[i].AcceptedAt,
		}
	}
	return items
}

func paymentToDomain(m *paymentModel) *biz.Payment {
	return &biz.Payment{
		ID:            m.ID,
		OrderID:       m.OrderID,
		PassengerID:   m.PassengerID,
		OutTradeNo:    m.OutTradeNo,
		AlipayTradeNo: m.AlipayTradeNo,
		Channel:       m.Channel,
		TotalAmount:   m.TotalAmount,
		Status:        m.Status,
		PaidAt:        m.PaidAt,
		NotifyPayload: m.NotifyPayload,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func paymentToRecord(e *biz.Payment) *paymentModel {
	return &paymentModel{
		ID:            e.ID,
		OrderID:       e.OrderID,
		PassengerID:   e.PassengerID,
		OutTradeNo:    e.OutTradeNo,
		AlipayTradeNo: e.AlipayTradeNo,
		Channel:       e.Channel,
		TotalAmount:   e.TotalAmount,
		Status:        e.Status,
		PaidAt:        e.PaidAt,
		NotifyPayload: e.NotifyPayload,
	}
}

func orderToDomain(m *orderModel) *biz.Order {
	return &biz.Order{
		ID:           m.ID,
		TripID:       m.TripID,
		PassengerID:  m.PassengerID,
		SeatsBooked:  m.SeatsBooked,
		TotalPrice:   m.TotalPrice,
		Status:       m.Status,
		AcceptedAt:   m.AcceptedAt,
		RejectReason: m.RejectReason,
		RejectedAt:   m.RejectedAt,
		RefundAmount: m.RefundAmount,
		RefundedAt:   m.RefundedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func orderToRecord(e *biz.Order) *orderModel {
	return &orderModel{
		ID:           e.ID,
		TripID:       e.TripID,
		PassengerID:  e.PassengerID,
		SeatsBooked:  e.SeatsBooked,
		TotalPrice:   e.TotalPrice,
		Status:       e.Status,
		AcceptedAt:   e.AcceptedAt,
		RejectReason: e.RejectReason,
		RejectedAt:   e.RejectedAt,
		RefundAmount: e.RefundAmount,
		RefundedAt:   e.RefundedAt,
	}
}

func tripToDomain(m *tripModel) *biz.TripSnapshot {
	return &biz.TripSnapshot{
		ID:             m.ID,
		DriverID:       m.DriverID,
		Origin:         m.Origin,
		Destination:    m.Destination,
		DepartTime:     m.DepartTime,
		SeatsAvailable: m.SeatsAvailable,
		Price:          m.Price,
		Status:         m.Status,
	}
}

func orderTransitionBizKey(transition biz.OrderTransition) string {
	key := strings.TrimSpace(transition.IdempotencyKey)
	if key == "" {
		return ""
	}
	action := strings.TrimSpace(transition.Action)
	if action == "" {
		action = "order"
	}
	return action + ":" + key
}

func bizStatusIn(status int, expected []int) bool {
	for _, item := range expected {
		if status == item {
			return true
		}
	}
	return false
}
