package carpool

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"
)

type AIApi struct{}

var aiService = service.ServiceGroupApp.CarpoolServiceGroup.AIService

func (a *AIApi) GetAISummary(c *gin.Context) {
	data, err := aiService.GetSummary(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("get ai summary failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AIApi) Chat(c *gin.Context) {
	var req carpoolReq.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := aiService.Chat(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("ai chat failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AIApi) PlanRainRoute(c *gin.Context) {
	var req carpoolReq.RainRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := aiService.PlanRainRoute(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("plan rain route failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AIApi) ChatWithRainRoute(c *gin.Context) {
	var req carpoolReq.ChatWithRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := aiService.ChatWithRainRoute(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("chat with rain route failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AIApi) ReportFlooding(c *gin.Context) {
	var req carpoolReq.FloodReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := aiService.ReportFlooding(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("report flooding failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AIApi) ListConversationLogs(c *gin.Context) {
	var search carpoolReq.AIConversationLogSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := aiService.ListConversationLogs(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("list ai conversation logs failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *AIApi) ListRoutePlanLogs(c *gin.Context) {
	var search carpoolReq.AIRoutePlanLogSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := aiService.ListRoutePlanLogs(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("list ai route plan logs failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *AIApi) ListFloodReports(c *gin.Context) {
	var search carpoolReq.AIFloodReportSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := aiService.ListFloodReports(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("list ai flood reports failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *AIApi) AuditFloodReport(c *gin.Context) {
	var req carpoolReq.AIFloodReportAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := aiService.AuditFloodReport(c.Request.Context(), req); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ai").Err(err).Error("audit ai flood report failed")
		if err == gorm.ErrRecordNotFound {
			response.FailWithMessage("report not found", c)
			return
		}
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("audited", c)
}

func (a *AIApi) ExportAI(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": aiService.ExportTaskID(c.Request.Context())}, c)
}

func (a *AIApi) GetTravelRouteInfo(c *gin.Context) {
	var req carpoolReq.RainRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if strings.TrimSpace(req.UserRole) == "" {
		req.UserRole = "passenger"
	}
	if strings.TrimSpace(req.Origin) == "" || strings.TrimSpace(req.Destination) == "" || strings.TrimSpace(req.City) == "" {
		response.FailWithMessage("origin, destination and city are required", c)
		return
	}
	response.OkWithData(gin.H{
		"origin":      strings.TrimSpace(req.Origin),
		"destination": strings.TrimSpace(req.Destination),
		"city":        strings.TrimSpace(req.City),
		"weather":     strings.TrimSpace(req.Weather),
		"available":   true,
		"source":      "admin-platform",
		"riskHints": []string{
			"avoid flooded roads when heavy rain is reported",
			"prefer main roads with recent traffic updates",
		},
	}, c)
}
