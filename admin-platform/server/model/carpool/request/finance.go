package request

import commonReq "ride-hailing/admin-server/model/common/request"

type FinanceSearch struct {
	commonReq.PageInfo
	OrderNo string `json:"orderNo" form:"orderNo"`
	Status  string `json:"status" form:"status"`
}
