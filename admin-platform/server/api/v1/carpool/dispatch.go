package carpool

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"
)

type DispatchApi struct{}

var dispatchService = service.ServiceGroupApp.CarpoolServiceGroup.DispatchService

func (a *DispatchApi) ListDispatchOrders(c *gin.Context) {
	var search carpoolReq.DispatchOrderSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := dispatchService.ListOrders(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("list dispatch orders failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *DispatchApi) GetDispatchOrderDetail(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := dispatchService.GetOrderDetail(c.Request.Context(), id)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("get dispatch order detail failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DispatchApi) CancelDispatchOrder(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var req carpoolReq.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	req.OrderID = strconv.FormatUint(id, 10)
	if err := dispatchService.CancelOrder(c.Request.Context(), req); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("cancel dispatch order failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("cancelled", c)
}

func (a *DispatchApi) ReassignDispatchOrder(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var req carpoolReq.ReassignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	req.OrderID = strconv.FormatUint(id, 10)
	data, err := dispatchService.ReassignOrder(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("reassign dispatch order failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DispatchApi) ScoreDrivers(c *gin.Context) {
	var req carpoolReq.DispatchScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := dispatchService.ScoreDrivers(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("score dispatch drivers failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DispatchApi) ListDispatchConfigs(c *gin.Context) {
	data, err := dispatchService.ListConfigs(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("list dispatch configs failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DispatchApi) SaveDispatchConfig(c *gin.Context) {
	var req carpoolReq.DispatchConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := dispatchService.SaveConfig(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("save dispatch config failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DispatchApi) PublishDispatchConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dispatchService.PublishConfig(c.Request.Context(), id, "admin"); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("publish dispatch config failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("published", c)
}

func (a *DispatchApi) RollbackDispatchConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dispatchService.RollbackConfig(c.Request.Context(), id, "admin"); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("rollback dispatch config failed")
		response.FailWithMessage("operation failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("rolled back", c)
}

func (a *DispatchApi) ListDispatchAudits(c *gin.Context) {
	var search carpoolReq.DispatchAuditSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := dispatchService.ListDispatchAudits(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("list dispatch audits failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *DispatchApi) ReplayTrack(c *gin.Context) {
	var search carpoolReq.TrackReplaySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := dispatchService.ReplayTrack(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("dispatch").Err(err).Error("replay dispatch track failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *DispatchApi) ExportDispatch(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": dispatchService.ExportTaskID(c.Request.Context())}, c)
}

func parseUintParam(c *gin.Context, name string) (uint64, error) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New(name + "必须是正整数")
	}
	return id, nil
}
