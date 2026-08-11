package request

import commonReq "ride-hailing/admin-server/model/common/request"

type ShuttleLineSearch struct {
	commonReq.PageInfo
	Status  *int   `json:"status" form:"status"`
	Keyword string `json:"keyword" form:"keyword"`
}

type ShuttleStation struct {
	Name      string `json:"name"`
	Time      string `json:"time"`
	OffsetMin int    `json:"offsetMin"`
	Type      string `json:"type"`
	Seats     int    `json:"seats"`
}

type ShuttleLinePayload struct {
	LineCode    string           `json:"lineCode"`
	LineName    string           `json:"lineName"`
	Route       string           `json:"route"`
	DepartTime  string           `json:"departTime"`
	ArriveTime  string           `json:"arriveTime"`
	VehicleType string           `json:"vehicleType"`
	TotalSeats  int              `json:"totalSeats"`
	RemainSeats int              `json:"remainSeats"`
	SortNo      int              `json:"sortNo"`
	Stations    []ShuttleStation `json:"stations"`
	Notice      string           `json:"notice"`
}
