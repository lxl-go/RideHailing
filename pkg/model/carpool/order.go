package carpool

import "time"

type Order struct {
	ID          int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:主键ID"`
	TripID      int64     `json:"tripId" gorm:"column:trip_id;type:bigint;not null;index:idx_trip_status;comment:行程ID"`
	PassengerID int64     `json:"passengerId" gorm:"column:passenger_id;type:bigint;not null;index:idx_passenger_status;comment:乘客ID"`
	SeatsBooked int       `json:"seatsBooked" gorm:"column:seats_booked;type:int;not null;comment:预订座位数"`
	TotalPrice  float64   `json:"totalPrice" gorm:"column:total_price;type:decimal(10,2);not null;comment:订单总价"`
	Status      int       `json:"status" gorm:"column:status;type:tinyint;not null;default:0;index:idx_trip_status;index:idx_passenger_status;comment:0待确认 1已确认 2已完成 3已取消"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (Order) TableName() string {
	return "carpool_order"
}

const (
	OrderStatusPending   = 0
	OrderStatusConfirmed = 1
	OrderStatusCompleted = 2
	OrderStatusCancelled = 3
)
