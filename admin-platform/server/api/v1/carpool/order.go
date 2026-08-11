package carpool

import (
	"strconv"

	"github.com/gin-gonic/gin"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"
)

type OrderApi struct{}

var orderService = service.ServiceGroupApp.CarpoolServiceGroup.OrderService

func (a *OrderApi) ListOrders(c *gin.Context) {
	var search carpoolReq.OrderSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := orderService.ListOrders(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("list orders failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *OrderApi) GetOverview(c *gin.Context) {
	var search carpoolReq.OrderSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	data, err := orderService.GetOverview(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("get order overview failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *OrderApi) GetOrderDetail(c *gin.Context) {
	detail, err := orderService.GetOrderDetail(c.Request.Context(), c.Param("orderNo"))
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("get order detail failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(detail, c)
}

func (a *OrderApi) GetOrderHistory(c *gin.Context) {
	history, err := orderService.GetStatusHistory(c.Request.Context(), c.Param("orderNo"))
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("get order history failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(gin.H{"list": history}, c)
}

func (a *OrderApi) ListRefunds(c *gin.Context) {
	var search carpoolReq.OrderRefundSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := orderService.ListRefunds(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("list refunds failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *OrderApi) ApplyRefund(c *gin.Context) {
	var req carpoolReq.OrderRefundApply
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	refund, err := orderService.ApplyRefund(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("apply refund failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.OkWithData(refund, c)
}

func (a *OrderApi) ReviewRefund(c *gin.Context) {
	var req carpoolReq.OrderRefundReview
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	refund, err := orderService.ReviewRefund(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("review refund failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.OkWithData(refund, c)
}

func (a *OrderApi) BatchRefund(c *gin.Context) {
	var req carpoolReq.OrderBatchRefund
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if len(req.OrderNos) == 0 || len(req.OrderNos) > 100 {
		response.FailWithMessage("orderNos size must be 1-100", c)
		return
	}
	result, err := orderService.BatchRefund(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("order").Err(err).Error("batch refund failed")
		response.FailWithMessage("operation failed", c)
		return
	}
	response.OkWithData(result, c)
}

func (a *OrderApi) ExportOrders(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "0"))
	if pageSize > 50000 {
		response.FailWithMessage("export limit exceeded", c)
		return
	}
	response.OkWithData(gin.H{"taskId": orderService.ExportOrders(c.Request.Context())}, c)
}
