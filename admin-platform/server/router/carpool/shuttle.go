package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type ShuttleRouter struct{}

func (r *ShuttleRouter) InitShuttleRouter(Router *gin.RouterGroup) {
	shuttleRouter := Router.Group("carpool/shuttle").Use(middleware.OperationRecord())
	shuttleRouterWithoutRecord := Router.Group("carpool/shuttle")
	{
		shuttleRouter.POST("line", shuttleApi.CreateLine)
		shuttleRouter.PUT("line/:id", shuttleApi.UpdateLine)
		shuttleRouter.POST("line/publish", shuttleApi.PublishLines)
		shuttleRouter.POST("line/export", shuttleApi.ExportLines)
	}
	{
		shuttleRouterWithoutRecord.GET("line/list", shuttleApi.ListLines)
		shuttleRouterWithoutRecord.GET("line/:id", shuttleApi.GetLine)
	}
}
