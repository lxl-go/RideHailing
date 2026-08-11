package carpool

import (
	"time"

	"gorm.io/gorm"
)

type Trip struct {
	ID             int64          `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:主键ID"`
	PublisherID    int64          `json:"publisherId" gorm:"column:publisher_id;type:bigint;not null;index:idx_publisher_status;comment:发布人ID"`
	PublisherRole  int            `json:"publisherRole" gorm:"column:publisher_role;type:tinyint;not null;default:1;comment:发布角色 1车主 2乘客"`
	TripType       int            `json:"tripType" gorm:"column:trip_type;type:tinyint;not null;default:1;index:idx_trip_type;comment:行程类型 1跨城 2市内"`
	OriginName     string         `json:"originName" gorm:"column:origin_name;type:varchar(256);not null;index:idx_origin_dest;comment:起点名称"`
	OriginLat      float64        `json:"originLat" gorm:"column:origin_lat;type:decimal(10,7);not null;comment:起点纬度"`
	OriginLng      float64        `json:"originLng" gorm:"column:origin_lng;type:decimal(10,7);not null;comment:起点经度"`
	DestName       string         `json:"destName" gorm:"column:dest_name;type:varchar(256);not null;index:idx_origin_dest;comment:终点名称"`
	DestLat        float64        `json:"destLat" gorm:"column:dest_lat;type:decimal(10,7);not null;comment:终点纬度"`
	DestLng        float64        `json:"destLng" gorm:"column:dest_lng;type:decimal(10,7);not null;comment:终点经度"`
	DepartureTime  time.Time      `json:"departureTime" gorm:"column:departure_time;not null;index:idx_departure_status;comment:出发时间"`
	ArriveTime     time.Time      `json:"arriveTime" gorm:"column:arrive_time;default null;comment:预计到达时间"`
	SeatsTotal     int            `json:"seatsTotal" gorm:"column:seats_total;type:int;not null;comment:总座位数"`
	SeatsAvailable int            `json:"seatsAvailable" gorm:"column:seats_available;type:int;not null;comment:剩余座位数"`
	ShareCost      float64        `json:"shareCost" gorm:"column:share_cost;type:decimal(8,2);not null;comment:分摊费用"`
	BaggageInfo    string         `json:"baggageInfo" gorm:"column:baggage_info;type:varchar(128);default:'';comment:行李说明"`
	PetAllowed     int            `json:"petAllowed" gorm:"column:pet_allowed;type:tinyint;not null;default:0;comment:允许宠物 0不允许 1允许"`
	Remarks        string         `json:"remarks" gorm:"column:remarks;type:varchar(512);default:'';comment:备注"`
	Status               int            `json:"status" gorm:"column:status;type:tinyint;not null;default:10;index:idx_publisher_status;index:idx_departure_status;comment:10待审核 20已通过 30已驳回"`
	RejectReason         string         `json:"rejectReason" gorm:"column:reject_reason;type:varchar(200);default:'';comment:驳回原因"`
	AuditOperatorID      int64          `json:"auditOperatorId" gorm:"column:audit_operator_id;default:0;comment:审核人"`
	AuditTime            *time.Time     `json:"auditTime" gorm:"column:audit_time;default:null;comment:审核时间"`
	RouteDistanceMeters  int            `json:"routeDistanceMeters" gorm:"column:route_distance_meters;default:0;comment:路线距离米"`
	RouteDurationSeconds int            `json:"routeDurationSeconds" gorm:"column:route_duration_seconds;default:0;comment:路线时长秒"`
	IsDeleted            bool           `json:"isDeleted" gorm:"column:is_deleted;not null;default:false;index;comment:是否软删除"`
	MatchedOrderID int64          `json:"matchedOrderId" gorm:"column:matched_order_id;type:bigint;default:0;comment:匹配订单ID"`
	CreatedAt      time.Time      `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:软删除时间"`
}

func (Trip) TableName() string {
	return "carpool_trip"
}
