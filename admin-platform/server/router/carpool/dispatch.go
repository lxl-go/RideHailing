package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type DispatchRouter struct{}

func (r *DispatchRouter) InitDispatchRouter(Router *gin.RouterGroup) {
	dispatchRouter := Router.Group("carpool/dispatch").Use(middleware.OperationRecord())
	dispatchRouterWithoutRecord := Router.Group("carpool/dispatch")
	{
		dispatchRouter.POST("order/:id/cancel", dispatchApi.CancelDispatchOrder)
		dispatchRouter.POST("order/:id/reassign", dispatchApi.ReassignDispatchOrder)
		dispatchRouter.POST("score", dispatchApi.ScoreDrivers)
		dispatchRouter.POST("config", dispatchApi.SaveDispatchConfig)
		dispatchRouter.POST("config/:id/publish", dispatchApi.PublishDispatchConfig)
		dispatchRouter.POST("config/:id/rollback", dispatchApi.RollbackDispatchConfig)
		dispatchRouter.POST("export", dispatchApi.ExportDispatch)
	}
	{
		dispatchRouterWithoutRecord.GET("order/list", dispatchApi.ListDispatchOrders)
		dispatchRouterWithoutRecord.GET("order/:id", dispatchApi.GetDispatchOrderDetail)
		dispatchRouterWithoutRecord.GET("config/list", dispatchApi.ListDispatchConfigs)
		dispatchRouterWithoutRecord.GET("audit/list", dispatchApi.ListDispatchAudits)
		dispatchRouterWithoutRecord.GET("track/replay", dispatchApi.ReplayTrack)
	}
}
