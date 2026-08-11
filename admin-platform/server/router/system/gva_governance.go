package system

import (
	"github.com/gin-gonic/gin"

	"ride-hailing/admin-server/middleware"
)

type GvaGovernanceRouter struct{}

func (r *GvaGovernanceRouter) InitGvaGovernanceRouter(Router *gin.RouterGroup) {
	governanceRouter := Router.Group("system/gva-governance").Use(middleware.OperationRecord())
	governanceRouterWithoutRecord := Router.Group("system/gva-governance")
	{
		governanceRouter.POST("export", gvaGovernanceApi.ExportGovernance)
	}
	{
		governanceRouterWithoutRecord.GET("summary", gvaGovernanceApi.GetGovernanceSummary)
		governanceRouterWithoutRecord.GET("routes", gvaGovernanceApi.GetRouteSnapshot)
		governanceRouterWithoutRecord.GET("audit", gvaGovernanceApi.GetAuditSnapshot)
		governanceRouterWithoutRecord.GET("datasource", gvaGovernanceApi.GetDatasourceSnapshot)
		governanceRouterWithoutRecord.GET("timed-task", gvaGovernanceApi.GetTimedTaskSnapshot)
	}
}
