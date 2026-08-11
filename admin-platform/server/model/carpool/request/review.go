package request

import commonReq "ride-hailing/admin-server/model/common/request"

type CertReviewListSearch struct {
	commonReq.PageInfo
	Status   *int   `json:"status" form:"status"`
	UserID   int64  `json:"userId" form:"userId"`
	CertType *int   `json:"certType" form:"certType"`
	Keyword  string `json:"keyword" form:"keyword"`
}

type CertReviewAction struct {
	RejectReason string `json:"rejectReason"`
}

type VehicleReviewListSearch struct {
	commonReq.PageInfo
	Status   *int   `json:"status" form:"status"`
	DriverID int64  `json:"driverId" form:"driverId"`
	Keyword  string `json:"keyword" form:"keyword"`
}

type VehicleReviewAction struct {
	Status       int    `json:"status" binding:"required,oneof=1 2"`
	RejectReason string `json:"rejectReason"`
}
