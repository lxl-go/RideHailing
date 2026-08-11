package carpool

import "time"

type Trip struct {
	ID             int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:主键ID"`
	PublisherID    int64     `json:"publisherId" gorm:"column:publisher_id;type:bigint;not null;index:idx_publisher_status;comment:发布人ID"`
	PublisherRole  int       `json:"publisherRole" gorm:"column:publisher_role;type:tinyint;not null;default:1;comment:发布角色 1车主 2乘客"`
	TripType       int       `json:"tripType" gorm:"column:trip_type;type:tinyint;not null;default:1;index:idx_trip_type;comment:行程类型 1跨城 2市内"`
	Origin         string    `json:"origin" gorm:"column:origin;type:varchar(255);not null;index:idx_origin_destination;comment:起点名称"`
	OriginLat      float64   `json:"originLat" gorm:"column:origin_lat;type:decimal(10,7);not null;comment:起点纬度"`
	OriginLng      float64   `json:"originLng" gorm:"column:origin_lng;type:decimal(10,7);not null;comment:起点经度"`
	Destination    string    `json:"destination" gorm:"column:destination;type:varchar(255);not null;index:idx_origin_destination;comment:终点名称"`
	DestLat        float64   `json:"destLat" gorm:"column:dest_lat;type:decimal(10,7);not null;comment:终点纬度"`
	DestLng        float64   `json:"destLng" gorm:"column:dest_lng;type:decimal(10,7);not null;comment:终点经度"`
	DepartTime     time.Time `json:"departTime" gorm:"column:depart_time;not null;index:idx_depart_status;comment:出发时间"`
	ArriveTime     time.Time `json:"arriveTime" gorm:"column:arrive_time;default null;comment:预计到达时间"`
	SeatsTotal     int       `json:"seatsTotal" gorm:"column:seats_total;type:int;not null;comment:总座位数"`
	SeatsAvailable int       `json:"seatsAvailable" gorm:"column:seats_available;type:int;not null;comment:剩余座位数"`
	Price          float64   `json:"price" gorm:"column:price;type:decimal(10,2);not null;comment:分摊费用"`
	BaggageInfo    string    `json:"baggageInfo" gorm:"column:baggage_info;type:varchar(128);default:'';comment:行李说明"`
	PetAllowed     int       `json:"petAllowed" gorm:"column:pet_allowed;type:tinyint;not null;default:0;comment:允许宠物 0否 1允许"`
	Remarks        string    `json:"remarks" gorm:"column:remarks;type:varchar(512);default:'';comment:备注"`
	Status         int       `json:"status" gorm:"column:status;type:tinyint;not null;default:0;index:idx_publisher_status;index:idx_depart_status;comment:0草稿 1招募中 2已满员 3进行中 4已完成 5已取消"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (Trip) TableName() string {
	return "carpool_trip"
}

const (
	TripStatusDraft      = 0
	TripStatusRecruiting = 1
	TripStatusFull       = 2
	TripStatusInProgress = 3
	TripStatusCompleted  = 4
	TripStatusCancelled  = 5
)
