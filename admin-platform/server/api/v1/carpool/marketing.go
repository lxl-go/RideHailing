package carpool

import (
	"github.com/gin-gonic/gin"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"
)

type MarketingApi struct{}

var marketingService = service.ServiceGroupApp.CarpoolServiceGroup.MarketingService

func (a *MarketingApi) CreateCouponTemplate(c *gin.Context) {
	var req carpoolReq.SaveCouponTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := marketingService.CreateCouponTemplate(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("create coupon template failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *MarketingApi) ListCouponTemplates(c *gin.Context) {
	var search carpoolReq.CouponTemplateSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := marketingService.ListCouponTemplates(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("list coupon templates failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *MarketingApi) ListUserCoupons(c *gin.Context) {
	var search carpoolReq.UserCouponSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := marketingService.ListUserCoupons(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("list user coupons failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *MarketingApi) IssueCoupon(c *gin.Context) {
	var req carpoolReq.IssueCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := marketingService.IssueCoupon(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("issue coupon failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *MarketingApi) RedeemCoupon(c *gin.Context) {
	var req carpoolReq.RedeemCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := marketingService.RedeemCoupon(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("redeem coupon failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *MarketingApi) DeleteCouponTemplate(c *gin.Context) {
	if err := marketingService.DeleteCouponTemplate(c.Request.Context(), c.Param("couponNo")); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("delete coupon template failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("deleted", c)
}

func (a *MarketingApi) ListCampaigns(c *gin.Context) {
	var search carpoolReq.MarketingCampaignSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := marketingService.ListCampaigns(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("list campaigns failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *MarketingApi) GetReferralSummary(c *gin.Context) {
	data, err := marketingService.GetReferralSummary(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("marketing").Err(err).Error("get referral summary failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *MarketingApi) ExportMarketing(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": marketingService.ExportTaskID(c.Request.Context())}, c)
}
