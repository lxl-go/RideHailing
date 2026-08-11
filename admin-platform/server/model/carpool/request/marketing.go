package request

import commonReq "ride-hailing/admin-server/model/common/request"

type CouponTemplateSearch struct {
	commonReq.PageInfo
	Keyword     string `json:"keyword" form:"keyword"`
	CouponType  string `json:"couponType" form:"couponType"`
	Status      string `json:"status" form:"status"`
	CityScope   string `json:"cityScope" form:"cityScope"`
	ServiceType string `json:"serviceType" form:"serviceType"`
}

type SaveCouponTemplateRequest struct {
	Name            string  `json:"name" binding:"required"`
	CouponType      string  `json:"couponType" binding:"required"`
	FaceValue       float64 `json:"faceValue"`
	DiscountRate    float64 `json:"discountRate"`
	ThresholdAmount float64 `json:"thresholdAmount"`
	ValidFrom       string  `json:"validFrom" binding:"required"`
	ValidTo         string  `json:"validTo" binding:"required"`
	CityScope       string  `json:"cityScope"`
	ServiceScope    string  `json:"serviceScope"`
	TimeScope       string  `json:"timeScope"`
	Stackable       bool    `json:"stackable"`
	TotalStock      int     `json:"totalStock"`
	Status          string  `json:"status"`
}

type IssueCouponRequest struct {
	CouponNo string `json:"couponNo" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	UserType string `json:"userType"`
	Source   string `json:"source"`
	Operator string `json:"operator"`
}

type RedeemCouponRequest struct {
	CouponCode  string  `json:"couponCode" binding:"required"`
	OrderNo     string  `json:"orderNo" binding:"required"`
	OrderAmount float64 `json:"orderAmount" binding:"required"`
}

type UserCouponSearch struct {
	commonReq.PageInfo
	CouponNo   string `json:"couponNo" form:"couponNo"`
	CouponCode string `json:"couponCode" form:"couponCode"`
	UserID     string `json:"userId" form:"userId"`
	Status     string `json:"status" form:"status"`
	Source     string `json:"source" form:"source"`
}

type MarketingCampaignSearch struct {
	commonReq.PageInfo
	Keyword string `json:"keyword" form:"keyword"`
	Channel string `json:"channel" form:"channel"`
	Status  string `json:"status" form:"status"`
}
