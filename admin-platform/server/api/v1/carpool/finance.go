package carpool

import (
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"

	"github.com/gin-gonic/gin"
)

type FinanceApi struct{}

var financeService = service.ServiceGroupApp.CarpoolServiceGroup.FinanceService

func (a *FinanceApi) ListTransactions(c *gin.Context) {
	var search carpoolReq.FinanceSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := financeService.ListTransactions(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("finance").Err(err).Error("list transactions failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *FinanceApi) ListRefunds(c *gin.Context) {
	var search carpoolReq.FinanceSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	list, total, err := financeService.ListRefunds(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("finance").Err(err).Error("list refunds failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *FinanceApi) GetSummary(c *gin.Context) {
	summary, err := financeService.GetSummary(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("finance").Err(err).Error("get summary failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(summary, c)
}

func (a *FinanceApi) ListAbnormalTransactions(c *gin.Context) {
	list, total, err := financeService.ListAbnormalTransactions(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("finance").Err(err).Error("list abnormal failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(gin.H{"list": list, "total": total}, c)
}

func (a *FinanceApi) ExportFinance(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": financeService.ExportTaskID(c.Request.Context())}, c)
}
