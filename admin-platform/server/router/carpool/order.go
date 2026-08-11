package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type OrderRouter struct{}

func (r *OrderRouter) InitOrderRouter(Router *gin.RouterGroup) {
	orderRouter := Router.Group("carpool/order").Use(middleware.OperationRecord())
	orderRouterWithoutRecord := Router.Group("carpool/order")
	{
		orderRouter.POST("refund/apply", orderApi.ApplyRefund)
		orderRouter.POST("refund/review", orderApi.ReviewRefund)
		orderRouter.POST("refund/batch", orderApi.BatchRefund)
		orderRouter.POST("export", orderApi.ExportOrders)
	}
	{
		orderRouterWithoutRecord.GET("overview", orderApi.GetOverview)
		orderRouterWithoutRecord.GET("list", orderApi.ListOrders)
		orderRouterWithoutRecord.GET("refund/list", orderApi.ListRefunds)
		orderRouterWithoutRecord.GET(":orderNo", orderApi.GetOrderDetail)
		orderRouterWithoutRecord.GET(":orderNo/history", orderApi.GetOrderHistory)
	}
}
