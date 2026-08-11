package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ride-hailing/services/trip-service/internal/biz"
)

type tripModel struct {
	ID                   int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID             int64      `gorm:"column:driver_id;type:bigint;not null;index:idx_driver_status"`
	PublisherID          int64      `gorm:"column:publisher_id;type:bigint;not null;index:idx_publisher_status"`
	PublisherRole        int        `gorm:"column:publisher_role;type:tinyint;not null;default:1"`
	TripType             int        `gorm:"column:trip_type;type:tinyint;not null;default:1"`
	Origin               string     `gorm:"column:origin;type:varchar(255);not null;index:idx_origin_destination"`
	OriginName           string     `gorm:"column:origin_name;type:varchar(256);not null"`
	OriginLat            float64    `gorm:"column:origin_lat;type:decimal(10,7);not null"`
	OriginLng            float64    `gorm:"column:origin_lng;type:decimal(10,7);not null"`
	Destination          string     `gorm:"column:destination;type:varchar(255);not null;index:idx_origin_destination"`
	DestName             string     `gorm:"column:dest_name;type:varchar(256);not null"`
	DestLat              float64    `gorm:"column:dest_lat;type:decimal(10,7);not null"`
	DestLng              float64    `gorm:"column:dest_lng;type:decimal(10,7);not null"`
	DepartTime           time.Time  `gorm:"column:depart_time"`
	ArriveTime           time.Time  `gorm:"column:arrive_time"`
	DepartureTime        time.Time  `gorm:"column:departure_time;not null"`
	SeatsTotal           int        `gorm:"column:seats_total;type:int;not null"`
	SeatsAvailable       int        `gorm:"column:seats_available;type:int;not null"`
	Price                float64    `gorm:"column:price;type:decimal(10,2);not null"`
	ShareCost            float64    `gorm:"column:share_cost;type:decimal(8,2);not null"`
	Status               int        `gorm:"column:status;type:tinyint;not null;default:10"`
	RejectReason         string     `gorm:"column:reject_reason;type:varchar(200);default:''"`
	AuditOperatorID      int64      `gorm:"column:audit_operator_id;default:0"`
	AuditTime            *time.Time `gorm:"column:audit_time"`
	RouteDistanceMeters  int        `gorm:"column:route_distance_meters;default:0"`
	RouteDurationSeconds int        `gorm:"column:route_duration_seconds;default:0"`
	IsDeleted            bool       `gorm:"column:is_deleted;not null;default:false"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (tripModel) TableName() string {
	return "carpool_trip"
}

type couponTemplateModel struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	CouponNo        string    `gorm:"column:coupon_no;type:varchar(64);not null;uniqueIndex:uk_marketing_coupon_no"`
	Name            string    `gorm:"column:name;type:varchar(128);not null"`
	CouponType      string    `gorm:"column:coupon_type;type:varchar(16);not null"`
	FaceValue       float64   `gorm:"column:face_value;type:decimal(10,2);not null;default:0"`
	DiscountRate    float64   `gorm:"column:discount_rate;type:decimal(4,2);not null;default:0"`
	ThresholdAmount float64   `gorm:"column:threshold_amount;type:decimal(10,2);not null;default:0"`
	ValidFrom       time.Time `gorm:"column:valid_from;not null"`
	ValidTo         time.Time `gorm:"column:valid_to;not null"`
	ServiceScope    string    `gorm:"column:service_scope;type:varchar(32);not null;default:''"`
	TotalStock      int       `gorm:"column:total_stock;type:int;not null;default:0"`
	IssuedCount     int       `gorm:"column:issued_count;type:int;not null;default:0"`
	Status          string    `gorm:"column:status;type:varchar(16);not null;default:'draft'"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (couponTemplateModel) TableName() string {
	return "marketing_coupon_template"
}

type userCouponModel struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	CouponCode     string    `gorm:"column:coupon_code;type:varchar(64);not null;uniqueIndex:uk_marketing_user_coupon_code"`
	CouponNo       string    `gorm:"column:coupon_no;type:varchar(64);not null;uniqueIndex:uk_mobile_coupon_user"`
	UserID         uint64    `gorm:"column:user_id;type:bigint;not null;uniqueIndex:uk_mobile_coupon_user"`
	UserType       string    `gorm:"column:user_type;type:varchar(16);not null;default:'passenger'"`
	Source         string    `gorm:"column:source;type:varchar(16);not null;default:'mobile'"`
	Status         string    `gorm:"column:status;type:varchar(16);not null;default:'unused'"`
	OrderNo        string    `gorm:"column:order_no;type:varchar(64);not null;default:''"`
	DiscountAmount float64   `gorm:"column:discount_amount;type:decimal(10,2);not null;default:0"`
	Operator       string    `gorm:"column:operator;type:varchar(64);not null;default:''"`
	IssuedAt       time.Time `gorm:"column:issued_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (userCouponModel) TableName() string {
	return "marketing_user_coupon"
}

type tripDemandModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	PassengerID int64     `gorm:"column:passenger_id;type:bigint;not null;index:idx_demand_passenger_status"`
	Origin      string    `gorm:"column:origin;type:varchar(255);not null"`
	Destination string    `gorm:"column:destination;type:varchar(255);not null"`
	DepartTime  time.Time `gorm:"column:depart_time;not null;index:idx_demand_depart_status"`
	Seats       int       `gorm:"column:seats;type:int;not null"`
	Budget      float64   `gorm:"column:budget;type:decimal(10,2);not null;default:0"`
	Remark      string    `gorm:"column:remark;type:varchar(512);not null;default:''"`
	Status      int       `gorm:"column:status;type:tinyint;not null;default:1;index:idx_demand_passenger_status;index:idx_demand_depart_status"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (tripDemandModel) TableName() string {
	return "carpool_trip_demand"
}

type orderModel struct {
	ID     int64 `gorm:"column:id;primaryKey;autoIncrement:false"`
	TripID int64 `gorm:"column:trip_id;type:bigint;not null;index:idx_trip_status"`
	Status int   `gorm:"column:status;type:tinyint;not null;default:0;index:idx_trip_status"`
}

func (orderModel) TableName() string {
	return "carpool_order"
}

const (
	couponStatusEnabled   = "enabled"
	orderStatusCanceled   = 2
	orderStatusRefunded   = 3
	activeOrderStatusPaid = 4
)

type TripRepo struct {
	db    *gorm.DB
	log   *zap.Logger
	redis *redis.Client
}

func NewTripRepo(db *gorm.DB, log *zap.Logger, redisClients ...*redis.Client) *TripRepo {
	var client *redis.Client
	if len(redisClients) > 0 {
		client = redisClients[0]
	}
	return &TripRepo{db: db, log: log, redis: client}
}

func (r *TripRepo) SearchTrips(ctx context.Context, origin, destination, departDate string, page, pageSize int) ([]biz.Trip, int64, error) {
	var models []tripModel
	var total int64
	query := r.db.WithContext(ctx).Model(&tripModel{}).
		Where("status = ? AND is_deleted = ? AND seats_available > 0 AND depart_time > ?", biz.TripStatusApproved, false, time.Now())
	if origin != "" {
		query = query.Where("origin LIKE ?", "%"+origin+"%")
	}
	if destination != "" {
		query = query.Where("destination LIKE ?", "%"+destination+"%")
	}
	if departDate != "" {
		query = query.Where("DATE(depart_time) = ?", departDate)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("depart_time ASC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	result := make([]biz.Trip, len(models))
	for i, m := range models {
		result[i] = *toDomain(&m)
	}
	return result, total, nil
}

func (r *TripRepo) GetByID(ctx context.Context, id int64) (*biz.Trip, error) {
	var m tripModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *TripRepo) Create(ctx context.Context, trip *biz.Trip) error {
	return r.db.WithContext(ctx).Create(toRecord(trip)).Error
}

func (r *TripRepo) AcquirePublishRequest(ctx context.Context, driverID int64, requestID string) (bool, error) {
	if r.redis == nil {
		return false, biz.ErrRedisUnavailable
	}
	if strings.TrimSpace(requestID) == "" {
		return false, biz.ErrInvalidTrip
	}
	key := fmt.Sprintf("app:trip:publish:%d:%s", driverID, requestID)
	ok, err := r.redis.SetNX(ctx, key, "1", 5*time.Minute).Result()
	if err != nil {
		return false, biz.ErrRedisUnavailable
	}
	return ok, nil
}

func (r *TripRepo) ReleasePublishRequest(ctx context.Context, driverID int64, requestID string) {
	if r.redis == nil || strings.TrimSpace(requestID) == "" {
		return
	}
	key := fmt.Sprintf("app:trip:publish:%d:%s", driverID, requestID)
	_ = r.redis.Del(ctx, key).Err()
}

func (r *TripRepo) GetPublishedTripResult(ctx context.Context, driverID int64, requestID string) (*biz.Trip, error) {
	if r.redis == nil {
		return nil, biz.ErrRedisUnavailable
	}
	value, err := r.redis.Get(ctx, fmt.Sprintf("app:trip:publish:result:%d:%s", driverID, requestID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, biz.ErrRedisUnavailable
	}
	var trip biz.Trip
	if err := json.Unmarshal([]byte(value), &trip); err != nil {
		return nil, biz.ErrRedisUnavailable
	}
	return &trip, nil
}

func (r *TripRepo) SavePublishedTripResult(ctx context.Context, driverID int64, requestID string, trip *biz.Trip) error {
	if r.redis == nil {
		return biz.ErrRedisUnavailable
	}
	value, err := json.Marshal(trip)
	if err != nil {
		return err
	}
	if err := r.redis.Set(ctx, fmt.Sprintf("app:trip:publish:result:%d:%s", driverID, requestID), value, 5*time.Minute).Err(); err != nil {
		return biz.ErrRedisUnavailable
	}
	return nil
}

func (r *TripRepo) ListByDriver(ctx context.Context, driverID int64, status int, page, pageSize int) ([]biz.Trip, int64, error) {
	var models []tripModel
	var total int64
	query := r.db.WithContext(ctx).Model(&tripModel{}).Where("driver_id = ? AND is_deleted = ?", driverID, false)
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	result := make([]biz.Trip, len(models))
	for i, m := range models {
		result[i] = *toDomain(&m)
	}
	return result, total, nil
}

func (r *TripRepo) UpdateStatus(ctx context.Context, id int64, status int) error {
	unlock, err := r.acquireLock(ctx, fmt.Sprintf("trip:update:%d", id), 30*time.Second)
	if err != nil {
		return err
	}
	defer unlock()
	return r.db.WithContext(ctx).Model(&tripModel{}).Where("id = ? AND is_deleted = ?", id, false).Update("status", status).Error
}

func (r *TripRepo) acquireLock(ctx context.Context, key string, ttl time.Duration) (func(), error) {
	if r.redis == nil {
		return func() {}, nil
	}
	ok, err := r.redis.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return nil, biz.ErrRedisUnavailable
	}
	if !ok {
		return nil, biz.ErrDuplicateTripRequest
	}
	return func() { _ = r.redis.Del(context.Background(), key).Err() }, nil
}

func (r *TripRepo) HasApprovedTimeConflict(ctx context.Context, driverID int64, departure time.Time) (bool, error) {
	var trips []tripModel
	if err := r.db.WithContext(ctx).Where("driver_id = ? AND status = ? AND is_deleted = ?", driverID, biz.TripStatusApproved, false).Find(&trips).Error; err != nil {
		return false, err
	}
	for _, trip := range trips {
		end := trip.DepartureTime
		if trip.RouteDurationSeconds > 0 {
			end = end.Add(time.Duration(trip.RouteDurationSeconds) * time.Second)
		} else if !trip.ArriveTime.IsZero() && trip.ArriveTime.After(end) {
			end = trip.ArriveTime
		}
		if departure.Before(end.Add(30*time.Minute)) && !departure.Before(trip.DepartureTime) {
			return true, nil
		}
	}
	return false, nil
}

func (r *TripRepo) ReviewTrip(ctx context.Context, id int64, status int, reviewerID int64, reason string) error {
	unlock, err := r.acquireLock(ctx, fmt.Sprintf("trip:review:%d", id), 60*time.Second)
	if err != nil {
		return err
	}
	defer unlock()
	if status != biz.TripStatusApproved && status != biz.TripStatusRejected {
		return biz.ErrInvalidReview
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trip tripModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND is_deleted = ?", id, false).First(&trip).Error; err != nil {
			return err
		}
		if trip.Status != biz.TripStatusPending {
			return biz.ErrTripNotPending
		}
		now := time.Now()
		updates := map[string]interface{}{"status": status, "audit_operator_id": reviewerID, "audit_time": now, "reject_reason": ""}
		if status == biz.TripStatusRejected {
			updates["reject_reason"] = strings.TrimSpace(reason)
		}
		return tx.Model(&tripModel{}).Where("id = ? AND status = ?", id, biz.TripStatusPending).Updates(updates).Error
	})
}

func (r *TripRepo) ListCoupons(ctx context.Context, passengerID int64, page, pageSize int) ([]biz.Coupon, int64, error) {
	var templates []couponTemplateModel
	var total int64
	now := time.Now()
	query := r.db.WithContext(ctx).Model(&couponTemplateModel{}).
		Where("status = ? AND valid_from <= ? AND valid_to >= ? AND (service_scope = ? OR service_scope = ? OR service_scope = '')", couponStatusEnabled, now, now, "carpool", "all")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	items := make([]biz.Coupon, 0, len(templates))
	for _, template := range templates {
		coupon := couponFromTemplate(template)
		var claimed userCouponModel
		if err := r.db.WithContext(ctx).Where("coupon_no = ? AND user_id = ?", template.CouponNo, passengerID).First(&claimed).Error; err == nil {
			coupon.ID = int64(claimed.ID)
			coupon.CouponCode = claimed.CouponCode
			coupon.Status = claimed.Status
			coupon.Claimed = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, err
		}
		items = append(items, coupon)
	}
	return items, total, nil
}

func (r *TripRepo) ClaimCoupon(ctx context.Context, passengerID int64, couponNo string, idempotencyKey string) (*biz.Coupon, bool, error) {
	var out biz.Coupon
	duplicated := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing userCouponModel
		if err := tx.Where("coupon_no = ? AND user_id = ?", couponNo, passengerID).First(&existing).Error; err == nil {
			coupon, err := r.couponFromIssued(ctx, tx, existing)
			if err != nil {
				return err
			}
			out = *coupon
			duplicated = true
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var template couponTemplateModel
		if err := findCouponTemplate(tx, couponNo, &template); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return biz.ErrCouponNotFound
			}
			return err
		}
		now := time.Now()
		if template.ValidFrom.After(now) || template.ValidTo.Before(now) || template.Status != couponStatusEnabled {
			return biz.ErrCouponNotFound
		}
		stock := tx.Model(&couponTemplateModel{}).
			Where("coupon_no = ? AND status = ? AND (total_stock = 0 OR issued_count < total_stock)", template.CouponNo, couponStatusEnabled).
			UpdateColumn("issued_count", gorm.Expr("issued_count + ?", 1))
		if stock.Error != nil {
			return stock.Error
		}
		if stock.RowsAffected != 1 {
			return biz.ErrCouponStockExhausted
		}
		issued := userCouponModel{
			CouponCode: nextCouponCode(passengerID),
			CouponNo:   template.CouponNo,
			UserID:     uint64(passengerID),
			UserType:   "passenger",
			Source:     "mobile",
			Status:     biz.CouponStatusUnused,
			Operator:   "mobile",
			IssuedAt:   now,
		}
		if err := tx.Create(&issued).Error; err != nil {
			if err := tx.Where("coupon_no = ? AND user_id = ?", template.CouponNo, passengerID).First(&issued).Error; err == nil {
				duplicated = true
			} else {
				return err
			}
		}
		out = couponFromTemplate(template)
		out.ID = int64(issued.ID)
		out.CouponCode = issued.CouponCode
		out.Status = issued.Status
		out.Claimed = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return &out, duplicated, nil
}

func (r *TripRepo) CreateDemand(ctx context.Context, demand *biz.TripDemand) error {
	return r.db.WithContext(ctx).Create(demandToRecord(demand)).Error
}

func (r *TripRepo) ListDemandsByPassenger(ctx context.Context, passengerID int64, status int, page, pageSize int) ([]biz.TripDemand, int64, error) {
	var models []tripDemandModel
	var total int64
	query := r.db.WithContext(ctx).Model(&tripDemandModel{}).Where("passenger_id = ?", passengerID)
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	result := make([]biz.TripDemand, len(models))
	for i := range models {
		result[i] = *demandToDomain(&models[i])
	}
	return result, total, nil
}

func (r *TripRepo) CancelDemand(ctx context.Context, id int64, passengerID int64) error {
	result := r.db.WithContext(ctx).Model(&tripDemandModel{}).
		Where("id = ? AND passenger_id = ? AND status = ?", id, passengerID, biz.DemandStatusPending).
		Update("status", biz.DemandStatusCancelled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrDemandCannotCancel
	}
	return nil
}

func (r *TripRepo) DeleteDriverTrip(ctx context.Context, id int64, driverID int64) error {
	unlock, err := r.acquireLock(ctx, fmt.Sprintf("trip:delete:%d", id), 30*time.Second)
	if err != nil {
		return err
	}
	defer unlock()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trip tripModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&trip).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return biz.ErrTripNotFound
			}
			return err
		}
		if trip.DriverID != driverID {
			return biz.ErrTripNotFound
		}
		if trip.IsDeleted || trip.Status == biz.TripStatusCancelled {
			return nil
		}
		active, err := hasActiveOrders(tx, id)
		if err != nil {
			return err
		}
		if active {
			return biz.ErrTripHasActiveOrders
		}
		result := tx.Model(&tripModel{}).Where("id = ? AND driver_id = ? AND is_deleted = ?", id, driverID, false).Updates(map[string]interface{}{
			"is_deleted": true,
			"status":     biz.TripStatusCancelled,
			"updated_at": time.Now(),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return biz.ErrTripNotFound
		}
		return nil
	})
}

func hasActiveOrders(tx *gorm.DB, tripID int64) (bool, error) {
	var active int64
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&orderModel{}).
		Where("trip_id = ? AND status NOT IN ?", tripID, []int{orderStatusCanceled, orderStatusRefunded}).
		Count(&active).Error
	return active > 0, err
}

func (r *TripRepo) couponFromIssued(ctx context.Context, tx *gorm.DB, issued userCouponModel) (*biz.Coupon, error) {
	var template couponTemplateModel
	if err := tx.WithContext(ctx).Where("coupon_no = ?", issued.CouponNo).First(&template).Error; err != nil {
		return nil, err
	}
	coupon := couponFromTemplate(template)
	coupon.ID = int64(issued.ID)
	coupon.CouponCode = issued.CouponCode
	coupon.Status = issued.Status
	coupon.Claimed = true
	return &coupon, nil
}

func toDomain(m *tripModel) *biz.Trip {
	return &biz.Trip{
		ID:                   m.ID,
		DriverID:             m.DriverID,
		PublisherID:          m.PublisherID,
		PublisherRole:        m.PublisherRole,
		TripType:             m.TripType,
		Origin:               m.Origin,
		OriginName:           m.OriginName,
		OriginLat:            m.OriginLat,
		OriginLng:            m.OriginLng,
		Destination:          m.Destination,
		DestName:             m.DestName,
		DestLat:              m.DestLat,
		DestLng:              m.DestLng,
		DepartTime:           m.DepartTime,
		ArriveTime:           m.ArriveTime,
		DepartureTime:        m.DepartureTime,
		SeatsTotal:           m.SeatsTotal,
		SeatsAvailable:       m.SeatsAvailable,
		Price:                m.Price,
		ShareCost:            m.ShareCost,
		Status:               m.Status,
		RejectReason:         m.RejectReason,
		AuditOperatorID:      m.AuditOperatorID,
		AuditTime:            m.AuditTime,
		RouteDistanceMeters:  m.RouteDistanceMeters,
		RouteDurationSeconds: m.RouteDurationSeconds,
		IsDeleted:            m.IsDeleted,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

func toRecord(e *biz.Trip) *tripModel {
	return &tripModel{
		ID:                   e.ID,
		DriverID:             e.DriverID,
		PublisherID:          e.PublisherID,
		PublisherRole:        e.PublisherRole,
		TripType:             e.TripType,
		Origin:               e.Origin,
		OriginName:           e.OriginName,
		OriginLat:            e.OriginLat,
		OriginLng:            e.OriginLng,
		Destination:          e.Destination,
		DestName:             e.DestName,
		DestLat:              e.DestLat,
		DestLng:              e.DestLng,
		DepartTime:           e.DepartTime,
		ArriveTime:           e.ArriveTime,
		DepartureTime:        e.DepartureTime,
		SeatsTotal:           e.SeatsTotal,
		SeatsAvailable:       e.SeatsAvailable,
		Price:                e.Price,
		ShareCost:            e.ShareCost,
		Status:               e.Status,
		RejectReason:         e.RejectReason,
		AuditOperatorID:      e.AuditOperatorID,
		AuditTime:            e.AuditTime,
		RouteDistanceMeters:  e.RouteDistanceMeters,
		RouteDurationSeconds: e.RouteDurationSeconds,
		IsDeleted:            e.IsDeleted,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
	}
}

func couponFromTemplate(m couponTemplateModel) biz.Coupon {
	return biz.Coupon{
		ID:              int64(m.ID),
		CouponNo:        m.CouponNo,
		Name:            m.Name,
		CouponType:      m.CouponType,
		FaceValue:       m.FaceValue,
		DiscountRate:    m.DiscountRate,
		ThresholdAmount: m.ThresholdAmount,
		ValidFrom:       m.ValidFrom,
		ValidTo:         m.ValidTo,
		Status:          m.Status,
	}
}

func demandToDomain(m *tripDemandModel) *biz.TripDemand {
	return &biz.TripDemand{
		ID:          m.ID,
		PassengerID: m.PassengerID,
		Origin:      m.Origin,
		Destination: m.Destination,
		DepartTime:  m.DepartTime,
		Seats:       m.Seats,
		Budget:      m.Budget,
		Remark:      m.Remark,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func demandToRecord(e *biz.TripDemand) *tripDemandModel {
	return &tripDemandModel{
		ID:          e.ID,
		PassengerID: e.PassengerID,
		Origin:      e.Origin,
		Destination: e.Destination,
		DepartTime:  e.DepartTime,
		Seats:       e.Seats,
		Budget:      e.Budget,
		Remark:      e.Remark,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
	}
}

func findCouponTemplate(tx *gorm.DB, couponNo string, out *couponTemplateModel) error {
	if id, err := strconv.ParseUint(couponNo, 10, 64); err == nil {
		return tx.Where("id = ? OR coupon_no = ?", id, couponNo).First(out).Error
	}
	return tx.Where("coupon_no = ?", couponNo).First(out).Error
}

func nextCouponCode(passengerID int64) string {
	return fmt.Sprintf("UC%d%d", passengerID, time.Now().UnixNano())
}
