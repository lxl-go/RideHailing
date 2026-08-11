package carpool

import "time"

type AdminOrderView struct {
	ID                     string    `json:"id"`
	OrderNo                string    `json:"orderNo"`
	ServiceType            string    `json:"serviceType"`
	PassengerID            string    `json:"passengerId"`
	PassengerName          string    `json:"passengerName"`
	PassengerPhone         string    `json:"passengerPhone"`
	DriverID               string    `json:"driverId"`
	DriverName             string    `json:"driverName"`
	DriverPhone            string    `json:"driverPhone"`
	VehicleNo              string    `json:"vehicleNo"`
	AIContextID            string    `json:"aiContextId"`
	AIRiskLevel            string    `json:"aiRiskLevel"`
	AIRouteSummary         string    `json:"aiRouteSummary"`
	RecommendedVehicleType string    `json:"recommendedVehicleType"`
	RouteName              string    `json:"routeName"`
	DepartTime             time.Time `json:"departTime"`
	ArrivalTime            time.Time `json:"arrivalTime"`
	Status                 string    `json:"status"`
	PayAmount              float64   `json:"payAmount"`
	RefundAmount           float64   `json:"refundAmount"`
	CancelFee              float64   `json:"cancelFee"`
	CancelReason           string    `json:"cancelReason"`
	Version                int       `json:"version"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type CarpoolOrderRecord struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TripID       uint64     `gorm:"column:trip_id;type:bigint;not null;index" json:"tripId"`
	PassengerID  uint64     `gorm:"column:passenger_id;type:bigint;not null;index" json:"passengerId"`
	SeatsBooked   int        `gorm:"column:seats_booked;type:int;not null;default:1" json:"seatsBooked"`
	TotalPrice    float64    `gorm:"column:total_price;type:decimal(10,2);not null;default:0" json:"totalPrice"`
	Status        string     `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	AcceptedAt    *time.Time `gorm:"column:accepted_at;index" json:"acceptedAt"`
	RejectReason  string     `gorm:"column:reject_reason;type:varchar(255);not null;default:''" json:"rejectReason"`
	RejectedAt    *time.Time `gorm:"column:rejected_at" json:"rejectedAt"`
	RefundAmount  float64    `gorm:"column:refund_amount;type:decimal(10,2);not null;default:0" json:"refundAmount"`
	RefundedAt    *time.Time `gorm:"column:refunded_at" json:"refundedAt"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (CarpoolOrderRecord) TableName() string {
	return "carpool_order"
}

type CarpoolTripRecord struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DriverID       uint64    `gorm:"column:driver_id;type:bigint;not null;default:0;index" json:"driverId"`
	StartLocation  string    `gorm:"column:start_location;type:varchar(255);not null;default:''" json:"startLocation"`
	EndLocation    string    `gorm:"column:end_location;type:varchar(255);not null;default:''" json:"endLocation"`
	DepartureTime  time.Time `gorm:"column:departure_time;not null;index" json:"departureTime"`
	AvailableSeats int       `gorm:"column:available_seats;type:int;not null;default:0" json:"availableSeats"`
	PricePerSeat   float64   `gorm:"column:price_per_seat;type:decimal(10,2);not null;default:0" json:"pricePerSeat"`
	Status         string    `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (CarpoolTripRecord) TableName() string {
	return "carpool_trip"
}

type CarpoolPaymentRecord struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID     uint64    `gorm:"column:order_id;type:bigint;not null;index" json:"orderId"`
	OutTradeNo  string    `gorm:"column:out_trade_no;type:varchar(64);not null;default:'';index" json:"outTradeNo"`
	TradeNo     string    `gorm:"column:trade_no;type:varchar(64);not null;default:'';index" json:"tradeNo"`
	TotalAmount float64   `gorm:"column:total_amount;type:decimal(10,2);not null;default:0" json:"totalAmount"`
	Status      string    `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	PaidAt      time.Time `gorm:"column:paid_at" json:"paidAt"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (CarpoolPaymentRecord) TableName() string {
	return "carpool_payment"
}

type OrderMain struct {
	ID                     uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:order id" json:"id"`
	OrderNo                string    `gorm:"column:order_no;type:varchar(64);not null;uniqueIndex:uk_order_main_no;comment:order number" json:"orderNo"`
	ServiceType            string    `gorm:"column:service_type;type:varchar(16);not null;index:idx_order_main_service;comment:carpool or shuttle" json:"serviceType"`
	PassengerID            uint64    `gorm:"column:passenger_id;type:bigint;not null;index:idx_order_main_passenger;comment:passenger id" json:"passengerId"`
	PassengerName          string    `gorm:"column:passenger_name;type:varchar(64);not null;default:'';comment:passenger name" json:"passengerName"`
	PassengerPhone         string    `gorm:"column:passenger_phone;type:varchar(32);not null;default:'';comment:passenger phone" json:"passengerPhone"`
	DriverID               uint64    `gorm:"column:driver_id;type:bigint;not null;index:idx_order_main_driver;comment:driver id" json:"driverId"`
	DriverName             string    `gorm:"column:driver_name;type:varchar(64);not null;default:'';comment:driver name" json:"driverName"`
	DriverPhone            string    `gorm:"column:driver_phone;type:varchar(32);not null;default:'';comment:driver phone" json:"driverPhone"`
	VehicleNo              string    `gorm:"column:vehicle_no;type:varchar(32);not null;default:'';comment:vehicle number" json:"vehicleNo"`
	AIContextID            string    `gorm:"column:ai_context_id;type:varchar(64);not null;default:'';index:idx_order_main_ai_context;comment:ai context id" json:"aiContextId"`
	AIRiskLevel            string    `gorm:"column:ai_risk_level;type:varchar(32);not null;default:'';index:idx_order_main_ai_risk;comment:ai risk level" json:"aiRiskLevel"`
	AIRouteSummary         string    `gorm:"column:ai_route_summary;type:varchar(512);not null;default:'';comment:ai route summary" json:"aiRouteSummary"`
	RecommendedVehicleType string    `gorm:"column:recommended_vehicle_type;type:varchar(32);not null;default:'';comment:recommended vehicle type" json:"recommendedVehicleType"`
	RouteName              string    `gorm:"column:route_name;type:varchar(128);not null;default:'';index:idx_order_main_route;comment:route name" json:"routeName"`
	DepartTime             time.Time `gorm:"column:depart_time;not null;index:idx_order_main_depart;comment:departure time" json:"departTime"`
	ArrivalTime            time.Time `gorm:"column:arrival_time;not null;comment:arrival time" json:"arrivalTime"`
	Status                 string    `gorm:"column:status;type:varchar(16);not null;default:'pending';index:idx_order_main_status;comment:order status" json:"status"`
	PayAmount              float64   `gorm:"column:pay_amount;type:decimal(10,2);not null;default:0;comment:paid amount" json:"payAmount"`
	RefundAmount           float64   `gorm:"column:refund_amount;type:decimal(10,2);not null;default:0;comment:refund amount" json:"refundAmount"`
	CancelFee              float64   `gorm:"column:cancel_fee;type:decimal(10,2);not null;default:0;comment:cancel fee" json:"cancelFee"`
	CancelReason           string    `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:cancel reason" json:"cancelReason"`
	Version                int       `gorm:"column:version;type:int;not null;default:1;comment:optimistic lock version" json:"version"`
	CreatedAt              time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt              time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (OrderMain) TableName() string {
	return "order_main"
}

type OrderRefund struct {
	ID                uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:refund id" json:"id"`
	RefundNo          string    `gorm:"column:refund_no;type:varchar(64);not null;uniqueIndex:uk_order_refund_no;comment:refund number" json:"refundNo"`
	OrderNo           string    `gorm:"column:order_no;type:varchar(64);not null;index:idx_order_refund_order;comment:order number" json:"orderNo"`
	ServiceType       string    `gorm:"column:service_type;type:varchar(16);not null;comment:carpool or shuttle" json:"serviceType"`
	PassengerID       uint64    `gorm:"column:passenger_id;type:bigint;not null;index:idx_order_refund_passenger;comment:passenger id" json:"passengerId"`
	RefundAmount      float64   `gorm:"column:refund_amount;type:decimal(10,2);not null;default:0;comment:refund amount" json:"refundAmount"`
	CancelFee         float64   `gorm:"column:cancel_fee;type:decimal(10,2);not null;default:0;comment:cancel fee" json:"cancelFee"`
	Reason            string    `gorm:"column:reason;type:varchar(255);not null;default:'';comment:refund reason" json:"reason"`
	ReviewType        string    `gorm:"column:review_type;type:varchar(16);not null;default:'auto';comment:auto or manual" json:"reviewType"`
	Status            string    `gorm:"column:status;type:varchar(16);not null;default:'pending';index:idx_order_refund_status;comment:refund status" json:"status"`
	IdempotentKey     string    `gorm:"column:idempotent_key;type:varchar(128);not null;uniqueIndex:uk_order_refund_idem;comment:idempotent key" json:"idempotentKey"`
	Reviewer          string    `gorm:"column:reviewer;type:varchar(64);not null;default:'';comment:reviewer" json:"reviewer"`
	ReviewRemark      string    `gorm:"column:review_remark;type:varchar(255);not null;default:'';comment:review remark" json:"reviewRemark"`
	EstimatedFinishAt time.Time `gorm:"column:estimated_finish_at;comment:estimated finish time" json:"estimatedFinishAt"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (OrderRefund) TableName() string {
	return "order_refund"
}

type OrderStatusHistory struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:history id" json:"id"`
	OrderNo    string    `gorm:"column:order_no;type:varchar(64);not null;index:idx_order_history_order;comment:order number" json:"orderNo"`
	FromStatus string    `gorm:"column:from_status;type:varchar(16);not null;default:'';comment:from status" json:"fromStatus"`
	ToStatus   string    `gorm:"column:to_status;type:varchar(16);not null;default:'';comment:to status" json:"toStatus"`
	Operator   string    `gorm:"column:operator;type:varchar(64);not null;default:'';comment:operator" json:"operator"`
	Reason     string    `gorm:"column:reason;type:varchar(255);not null;default:'';comment:reason" json:"reason"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}
