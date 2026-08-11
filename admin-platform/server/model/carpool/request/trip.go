package request

import commonReq "ride-hailing/admin-server/model/common/request"

type TripListSearch struct {
	commonReq.PageInfo
	Status      *int   `json:"status" form:"status"`
	Keyword     string `json:"keyword" form:"keyword"`
	PublisherID int64  `json:"publisherId" form:"publisherId"`
}

type TripAction struct {
	Reason string `json:"reason"`
}
