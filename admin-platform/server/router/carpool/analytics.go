package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type AnalyticsRouter struct{}

func (r *AnalyticsRouter) InitAnalyticsRouter(Router *gin.RouterGroup) {
	analyticsRouter := Router.Group("carpool/analytics").Use(middleware.OperationRecord())
	analyticsRouterWithoutRecord := Router.Group("carpool/analytics")
	{
		analyticsRouter.POST("export", analyticsApi.ExportAnalytics)
	}
	{
		analyticsRouterWithoutRecord.GET("dashboard", analyticsApi.GetDashboard)
		analyticsRouterWithoutRecord.GET("order-volume", analyticsApi.GetOrderVolume)
		analyticsRouterWithoutRecord.GET("classification", analyticsApi.GetOrderClassification)
		analyticsRouterWithoutRecord.GET("conversion", analyticsApi.GetConversion)
		analyticsRouterWithoutRecord.GET("repurchase", analyticsApi.GetRepurchase)
	}
}
