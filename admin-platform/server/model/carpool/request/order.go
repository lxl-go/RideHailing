package request

import commonReq "ride-hailing/admin-server/model/common/request"

type OrderSearch struct {
	commonReq.PageInfo
	OrderNo     string `json:"orderNo" form:"orderNo"`
	ServiceType string `json:"serviceType" form:"serviceType"`
	Status      string `json:"status" form:"status"`
	StartDate   string `json:"startDate" form:"startDate"`
	EndDate     string `json:"endDate" form:"endDate"`
}

type OrderRefundSearch struct {
	commonReq.PageInfo
	OrderNo string `json:"orderNo" form:"orderNo"`
	Status  string `json:"status" form:"status"`
}

type OrderRefundApply struct {
	OrderNo         string `json:"orderNo" binding:"required"`
	Reason          string `json:"reason" binding:"required"`
	IdempotentKey   string `json:"idempotentKey" binding:"required"`
	Operator        string `json:"operator"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type OrderRefundReview struct {
	RefundNo string `json:"refundNo" binding:"required"`
	Decision string `json:"decision" binding:"required"`
	Reviewer string `json:"reviewer" binding:"required"`
	Remark   string `json:"remark"`
}

type OrderBatchRefund struct {
	OrderNos       []string `json:"orderNos" binding:"required"`
	Reason         string   `json:"reason" binding:"required"`
	Operator       string   `json:"operator"`
	IdempotentSeed string   `json:"idempotentSeed" binding:"required"`
}
