package carpool

import (
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils/logger"

	"github.com/gin-gonic/gin"
)

type AnalyticsApi struct{}

var analyticsService = service.ServiceGroupApp.CarpoolServiceGroup.AnalyticsService

func (a *AnalyticsApi) GetDashboard(c *gin.Context) {
	var search carpoolReq.AnalyticsSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := analyticsService.GetDashboard(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("analytics").Err(err).Error("get dashboard failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AnalyticsApi) GetOrderVolume(c *gin.Context) {
	var search carpoolReq.AnalyticsSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := analyticsService.GetOrderVolume(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("analytics").Err(err).Error("get order volume failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AnalyticsApi) GetOrderClassification(c *gin.Context) {
	var search carpoolReq.AnalyticsSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := analyticsService.GetOrderClassification(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("analytics").Err(err).Error("get order classification failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AnalyticsApi) GetConversion(c *gin.Context) {
	var search carpoolReq.AnalyticsSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := analyticsService.GetConversion(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("analytics").Err(err).Error("get conversion failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AnalyticsApi) GetRepurchase(c *gin.Context) {
	var search carpoolReq.AnalyticsSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := analyticsService.GetRepurchase(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("analytics").Err(err).Error("get repurchase failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AnalyticsApi) ExportAnalytics(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": analyticsService.ExportTaskID(c.Request.Context())}, c)
}
