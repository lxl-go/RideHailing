package carpool

import (
	"github.com/gin-gonic/gin"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"
)

type PerformanceApi struct{}

var performanceService = service.ServiceGroupApp.CarpoolServiceGroup.PerformanceService

func (a *PerformanceApi) GetPerformanceSummary(c *gin.Context) {
	data, err := performanceService.GetSummary(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("performance").Err(err).Error("get performance summary failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PerformanceApi) ListPerformanceReports(c *gin.Context) {
	var search carpoolReq.PerformanceReportSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := performanceService.ListReports(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("performance").Err(err).Error("list performance reports failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *PerformanceApi) CreatePerformanceReport(c *gin.Context) {
	var req carpoolReq.SavePerformanceReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := performanceService.CreateReport(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("performance").Err(err).Error("create performance report failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PerformanceApi) ListPerformanceScenarios(c *gin.Context) {
	var search carpoolReq.PerformanceScenarioSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := performanceService.ListScenarios(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("performance").Err(err).Error("list performance scenarios failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PerformanceApi) GetRuntimeSnapshot(c *gin.Context) {
	data, err := performanceService.GetRuntimeSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("performance").Err(err).Error("get runtime snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PerformanceApi) ExportPerformance(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": performanceService.ExportTaskID(c.Request.Context())}, c)
}
