package request

import (
	"time"

	commonReq "ride-hailing/admin-server/model/common/request"
)

type DispatchOrderSearch struct {
	commonReq.PageInfo
	OrderNo     string `json:"orderNo" form:"orderNo"`
	ServiceType string `json:"serviceType" form:"serviceType"`
	Status      string `json:"status" form:"status"`
	City        string `json:"city" form:"city"`
	FleetID     string `json:"fleetId" form:"fleetId"`
	HotZone     string `json:"hotZone" form:"hotZone"`
	Plate       string `json:"plate" form:"plate"`
	Phone       string `json:"phone" form:"phone"`
}

type CancelOrderRequest struct {
	OrderID        string `json:"orderId" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	Operator       string `json:"operator"`
	IdempotencyKey string `json:"idempotencyKey" binding:"required"`
}

type ReassignOrderRequest struct {
	OrderID        string `json:"orderId" binding:"required"`
	DriverID       string `json:"driverId" binding:"required"`
	DriverName     string `json:"driverName"`
	DriverPhone    string `json:"driverPhone"`
	VehicleNo      string `json:"vehicleNo"`
	Reason         string `json:"reason" binding:"required"`
	Operator       string `json:"operator"`
	IdempotencyKey string `json:"idempotencyKey" binding:"required"`
}

type DispatchConfigRequest struct {
	ID                  uint64  `json:"id"`
	ConfigNo            string  `json:"configNo"`
	City                string  `json:"city" binding:"required"`
	FleetID             string  `json:"fleetId"`
	HotZone             string  `json:"hotZone"`
	DayDistanceWeight   float64 `json:"dayDistanceWeight"`
	DayRatingWeight     float64 `json:"dayRatingWeight"`
	DayIdleWeight       float64 `json:"dayIdleWeight"`
	NightDistanceWeight float64 `json:"nightDistanceWeight"`
	NightRatingWeight   float64 `json:"nightRatingWeight"`
	NightIdleWeight     float64 `json:"nightIdleWeight"`
	Enabled             bool    `json:"enabled"`
	Operator            string  `json:"operator"`
}

type DispatchScoreRequest struct {
	OrderID        string            `json:"orderId" binding:"required"`
	City           string            `json:"city"`
	FleetID        string            `json:"fleetId"`
	HotZone        string            `json:"hotZone"`
	Candidates     []DriverCandidate `json:"candidates" binding:"required"`
	IdempotencyKey string            `json:"idempotencyKey"`
	Operator       string            `json:"operator"`
}

type DriverCandidate struct {
	DriverID       string          `json:"driverId"`
	DriverName     string          `json:"driverName"`
	City           string          `json:"city"`
	FleetID        string          `json:"fleetId"`
	HotZone        string          `json:"hotZone"`
	DistanceKM     float64         `json:"distanceKm"`
	Rating         float64         `json:"rating"`
	IdleMinutes    float64         `json:"idleMinutes"`
	VehicleType    string          `json:"vehicleType"`
	Online         bool            `json:"online"`
	ServiceWindows []ServiceWindow `json:"serviceWindows"`
}

type ServiceWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type DispatchAuditSearch struct {
	commonReq.PageInfo
	OrderNo  string `json:"orderNo" form:"orderNo"`
	Action   string `json:"action" form:"action"`
	DriverID uint64 `json:"driverId" form:"driverId"`
}

type TrackReplaySearch struct {
	commonReq.PageInfo
	DriverID  string `json:"driverId" form:"driverId"`
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
}
