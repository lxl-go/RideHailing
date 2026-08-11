package carpool

import "time"

type OrderDispatchAudit struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:dispatch audit id" json:"id"`
	AuditNo        string    `gorm:"column:audit_no;type:varchar(64);not null;uniqueIndex:uk_dispatch_audit_no;comment:audit no" json:"auditNo"`
	OrderID        uint64    `gorm:"column:order_id;not null;index:idx_dispatch_audit_order;comment:order id" json:"orderId"`
	OrderNo        string    `gorm:"column:order_no;type:varchar(64);not null;index:idx_dispatch_audit_order_no;comment:order no" json:"orderNo"`
	Action         string    `gorm:"column:action;type:varchar(32);not null;index:idx_dispatch_audit_action;comment:score cancel reassign" json:"action"`
	DriverID       uint64    `gorm:"column:driver_id;not null;default:0;index:idx_dispatch_audit_driver;comment:driver id" json:"driverId"`
	DriverName     string    `gorm:"column:driver_name;type:varchar(64);not null;default:'';comment:driver name" json:"driverName"`
	Decision       string    `gorm:"column:decision;type:varchar(32);not null;index:idx_dispatch_audit_decision;comment:selected cancelled reassigned skipped" json:"decision"`
	DecisionReason string    `gorm:"column:decision_reason;type:varchar(512);not null;default:'';comment:decision reason" json:"decisionReason"`
	Score          float64   `gorm:"column:score;type:decimal(8,2);not null;default:0;comment:driver score" json:"score"`
	ScoreDetail    string    `gorm:"column:score_detail;type:text;comment:score detail json" json:"scoreDetail"`
	IdempotencyKey string    `gorm:"column:idempotency_key;type:varchar(128);not null;uniqueIndex:uk_dispatch_audit_idem;comment:idempotency key" json:"idempotencyKey"`
	Operator       string    `gorm:"column:operator;type:varchar(64);not null;default:'';comment:operator" json:"operator"`
	TraceID        string    `gorm:"column:trace_id;type:varchar(64);not null;default:'';index:idx_dispatch_audit_trace;comment:trace id" json:"traceId"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (OrderDispatchAudit) TableName() string {
	return "order_dispatch_audit"
}

type DispatchConfig struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:dispatch config id" json:"id"`
	ConfigNo            string    `gorm:"column:config_no;type:varchar(64);not null;uniqueIndex:uk_dispatch_config_no;comment:config no" json:"configNo"`
	City                string    `gorm:"column:city;type:varchar(64);not null;index:idx_dispatch_config_scope;comment:city" json:"city"`
	FleetID             string    `gorm:"column:fleet_id;type:varchar(64);not null;default:'';index:idx_dispatch_config_scope;comment:fleet id" json:"fleetId"`
	HotZone             string    `gorm:"column:hot_zone;type:varchar(64);not null;default:'';index:idx_dispatch_config_scope;comment:hot zone" json:"hotZone"`
	DayDistanceWeight   float64   `gorm:"column:day_distance_weight;type:decimal(6,4);not null;default:0.65;comment:day distance weight" json:"dayDistanceWeight"`
	DayRatingWeight     float64   `gorm:"column:day_rating_weight;type:decimal(6,4);not null;default:0.25;comment:day rating weight" json:"dayRatingWeight"`
	DayIdleWeight       float64   `gorm:"column:day_idle_weight;type:decimal(6,4);not null;default:0.10;comment:day idle weight" json:"dayIdleWeight"`
	NightDistanceWeight float64   `gorm:"column:night_distance_weight;type:decimal(6,4);not null;default:0.15;comment:night distance weight" json:"nightDistanceWeight"`
	NightRatingWeight   float64   `gorm:"column:night_rating_weight;type:decimal(6,4);not null;default:0.75;comment:night rating weight" json:"nightRatingWeight"`
	NightIdleWeight     float64   `gorm:"column:night_idle_weight;type:decimal(6,4);not null;default:0.10;comment:night idle weight" json:"nightIdleWeight"`
	Enabled             bool      `gorm:"column:enabled;type:tinyint;not null;default:1;index:idx_dispatch_config_enabled;comment:enabled" json:"enabled"`
	Published           bool      `gorm:"column:published;type:tinyint;not null;default:0;index:idx_dispatch_config_published;comment:published" json:"published"`
	Version             int       `gorm:"column:version;type:int;not null;default:1;comment:version" json:"version"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (DispatchConfig) TableName() string {
	return "dispatch_config"
}

type DispatchConfigVersion struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:dispatch config version id" json:"id"`
	ConfigNo     string    `gorm:"column:config_no;type:varchar(64);not null;index:idx_dispatch_config_version_no;comment:config no" json:"configNo"`
	Version      int       `gorm:"column:version;type:int;not null;index:idx_dispatch_config_version;comment:version" json:"version"`
	SnapshotJSON string    `gorm:"column:snapshot_json;type:text;comment:snapshot json" json:"snapshotJson"`
	Operator     string    `gorm:"column:operator;type:varchar(64);not null;default:'';comment:operator" json:"operator"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (DispatchConfigVersion) TableName() string {
	return "dispatch_config_version"
}

type DriverLocationPoint struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:driver location point id" json:"id"`
	DriverID   uint64    `gorm:"column:driver_id;not null;index:idx_driver_location_driver;comment:driver id" json:"driverId"`
	City       string    `gorm:"column:city;type:varchar(64);not null;index:idx_driver_location_pool;comment:city" json:"city"`
	FleetID    string    `gorm:"column:fleet_id;type:varchar(64);not null;default:'';index:idx_driver_location_pool;comment:fleet id" json:"fleetId"`
	HotZone    string    `gorm:"column:hot_zone;type:varchar(64);not null;default:'';index:idx_driver_location_pool;comment:hot zone" json:"hotZone"`
	Lat        float64   `gorm:"column:lat;type:decimal(10,7);not null;default:0;comment:latitude" json:"lat"`
	Lng        float64   `gorm:"column:lng;type:decimal(10,7);not null;default:0;comment:longitude" json:"lng"`
	Online     bool      `gorm:"column:online;type:tinyint;not null;default:1;index:idx_driver_location_online;comment:online" json:"online"`
	ReportedAt time.Time `gorm:"column:reported_at;not null;index:idx_driver_location_reported;comment:reported time" json:"reportedAt"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (DriverLocationPoint) TableName() string {
	return "driver_location_point"
}

type RealtimeMessage struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:realtime message id" json:"id"`
	Topic     string    `gorm:"column:topic;type:varchar(64);not null;index:idx_realtime_message_topic;comment:topic" json:"topic"`
	UserID    uint64    `gorm:"column:user_id;not null;default:0;index:idx_realtime_message_user;comment:user id" json:"userId"`
	UserRole  string    `gorm:"column:user_role;type:varchar(32);not null;default:'';index:idx_realtime_message_role;comment:user role" json:"userRole"`
	Payload   string    `gorm:"column:payload;type:text;comment:payload json" json:"payload"`
	Delivered bool      `gorm:"column:delivered;type:tinyint;not null;default:0;index:idx_realtime_message_delivered;comment:delivered" json:"delivered"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (RealtimeMessage) TableName() string {
	return "realtime_message"
}
