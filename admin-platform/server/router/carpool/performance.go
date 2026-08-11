package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type PerformanceRouter struct{}

func (r *PerformanceRouter) InitPerformanceRouter(Router *gin.RouterGroup) {
	performanceRouter := Router.Group("carpool/performance").Use(middleware.OperationRecord())
	performanceRouterWithoutRecord := Router.Group("carpool/performance")
	{
		performanceRouter.POST("report", performanceApi.CreatePerformanceReport)
		performanceRouter.POST("export", performanceApi.ExportPerformance)
	}
	{
		performanceRouterWithoutRecord.GET("summary", performanceApi.GetPerformanceSummary)
		performanceRouterWithoutRecord.GET("report/list", performanceApi.ListPerformanceReports)
		performanceRouterWithoutRecord.GET("scenario/list", performanceApi.ListPerformanceScenarios)
		performanceRouterWithoutRecord.GET("runtime", performanceApi.GetRuntimeSnapshot)
	}
}
