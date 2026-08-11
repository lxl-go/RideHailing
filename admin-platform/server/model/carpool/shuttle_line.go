package carpool

import "time"

type ShuttleLine struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:班车线路ID" json:"id"`
	LineCode    string    `gorm:"column:line_code;type:varchar(32);not null;uniqueIndex:uk_carpool_shuttle_line_code;comment:线路编号" json:"lineCode"`
	LineName    string    `gorm:"column:line_name;type:varchar(64);not null;index:idx_carpool_shuttle_keyword;comment:线路名称" json:"lineName"`
	Route       string    `gorm:"column:route;type:varchar(255);not null;index:idx_carpool_shuttle_keyword;comment:站点时序" json:"route"`
	DepartTime  string    `gorm:"column:depart_time;type:varchar(5);not null;comment:发车时间" json:"departTime"`
	ArriveTime  string    `gorm:"column:arrive_time;type:varchar(5);not null;comment:到达时间" json:"arriveTime"`
	VehicleType string    `gorm:"column:vehicle_type;type:varchar(32);not null;comment:车型" json:"vehicleType"`
	TotalSeats  int       `gorm:"column:total_seats;type:int;not null;default:0;comment:总座位数" json:"totalSeats"`
	RemainSeats int       `gorm:"column:remain_seats;type:int;not null;default:0;comment:剩余座位数" json:"remainSeats"`
	SortNo      int       `gorm:"column:sort_no;type:int;not null;default:1;index:idx_carpool_shuttle_sort;comment:排序号" json:"sortNo"`
	Status      int8      `gorm:"column:status;type:tinyint;not null;default:0;index:idx_carpool_shuttle_status;comment:0草稿 1已发布 2已停运" json:"status"`
	Stations    string    `gorm:"column:stations;type:json;comment:站点JSON" json:"stations"`
	Notice      string    `gorm:"column:notice;type:varchar(512);not null;default:'';comment:运营提示" json:"notice"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

func (ShuttleLine) TableName() string {
	return "carpool_shuttle_line"
}
