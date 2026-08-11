package carpool

import (
	"strconv"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"

	"github.com/gin-gonic/gin"
)

type ShuttleApi struct{}

var shuttleService = service.ServiceGroupApp.CarpoolServiceGroup.ShuttleService

func (a *ShuttleApi) ListLines(c *gin.Context) {
	var search carpoolReq.ShuttleLineSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := shuttleService.ListLines(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("list shuttle lines failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, c)
}

func (a *ShuttleApi) GetLine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("invalid id", c)
		return
	}
	line, err := shuttleService.GetLine(c.Request.Context(), id)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("get shuttle line failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(line, c)
}

func (a *ShuttleApi) CreateLine(c *gin.Context) {
	var payload carpoolReq.ShuttleLinePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	line, err := shuttleService.CreateLine(c.Request.Context(), payload)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("create shuttle line failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.OkWithData(line, c)
}

func (a *ShuttleApi) UpdateLine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("invalid id", c)
		return
	}
	var payload carpoolReq.ShuttleLinePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	line, err := shuttleService.UpdateLine(c.Request.Context(), id, payload)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("update shuttle line failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.OkWithData(line, c)
}

func (a *ShuttleApi) PublishLines(c *gin.Context) {
	var req struct {
		Ids []uint64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if len(req.Ids) == 0 {
		response.FailWithMessage("ids is required", c)
		return
	}
	if err := shuttleService.PublishLines(c.Request.Context(), req.Ids); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("publish shuttle lines failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.Ok(c)
}

func (a *ShuttleApi) ExportLines(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": shuttleService.ExportLines(c.Request.Context())}, c)
}
