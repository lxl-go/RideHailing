package request

import commonReq "ride-hailing/admin-server/model/common/request"

type PersonSearch struct {
	commonReq.PageInfo
	PersonType string `json:"personType" form:"personType"`
	Status     string `json:"status" form:"status"`
	RoleCode   string `json:"roleCode" form:"roleCode"`
	Keyword    string `json:"keyword" form:"keyword"`
}

type PersonPayload struct {
	PersonType        string   `json:"personType" binding:"required"`
	Name              string   `json:"name" binding:"required"`
	Phone             string   `json:"phone" binding:"required"`
	Email             string   `json:"email"`
	IDCardNo          string   `json:"idCardNo" binding:"required"`
	DriverLicenseNo   string   `json:"driverLicenseNo"`
	VehicleNo         string   `json:"vehicleNo"`
	VehicleType       string   `json:"vehicleType"`
	CommonAddress     string   `json:"commonAddress"`
	PaymentPreference string   `json:"paymentPreference"`
	Rating            float64  `json:"rating"`
	Status            string   `json:"status"`
	RegisterDate      string   `json:"registerDate" binding:"required"`
	Roles             []string `json:"roles" binding:"required"`
}

type PersonRoleAssign struct {
	PersonID uint64   `json:"personId" binding:"required"`
	Roles    []string `json:"roles" binding:"required"`
}

type PersonBatchStatus struct {
	IDs    []uint64 `json:"ids" binding:"required"`
	Status string   `json:"status" binding:"required"`
	Reason string   `json:"reason"`
}

type PersonBatchDelete struct {
	IDs    []uint64 `json:"ids" binding:"required"`
	Reason string   `json:"reason"`
}

type PersonImportPayload struct {
	SourceType string          `json:"sourceType"`
	Operator   string          `json:"operator"`
	Rows       []PersonPayload `json:"rows" binding:"required"`
}

type PersonImportErrorSearch struct {
	commonReq.PageInfo
	BatchNo string `json:"batchNo" form:"batchNo"`
}
