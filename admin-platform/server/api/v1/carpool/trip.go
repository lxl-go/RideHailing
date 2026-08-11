package carpool

import (
	"strconv"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	carpoolResp "ride-hailing/admin-server/model/carpool/response"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"

	"github.com/gin-gonic/gin"
)

type TripApi struct{}

var tripService = service.ServiceGroupApp.CarpoolServiceGroup.TripService

func (a *TripApi) ListTrips(c *gin.Context) {
	var search carpoolReq.TripListSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := tripService.ListTrips(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("list trips failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{
		List:     carpoolResp.NewTripResponses(list),
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, c)
}

func (a *TripApi) GetTrip(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("invalid id", c)
		return
	}
	trip, err := tripService.GetTrip(c.Request.Context(), id)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("get trip failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(carpoolResp.NewTripResponse(*trip), c)
}

func (a *TripApi) ReviewTrip(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("行程ID格式不正确", c)
		return
	}
	var action struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&action); err != nil {
		response.FailWithMessage("请求参数不正确: "+err.Error(), c)
		return
	}
	if err := tripService.ReviewTrip(c.Request.Context(), id, int64(utils.GetUserID(c)), action.Approved, action.Reason); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("review trip failed")
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("审核完成", c)
}

func (a *TripApi) DeactivateTrip(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("invalid id", c)
		return
	}
	var act carpoolReq.TripAction
	if err := c.ShouldBindJSON(&act); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if act.Reason == "" {
		response.FailWithMessage("reason is required", c)
		return
	}
	if err := tripService.DeactivateTrip(c.Request.Context(), id, act.Reason); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("deactivate trip failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.OkWithMessage("deactivated", c)
}
