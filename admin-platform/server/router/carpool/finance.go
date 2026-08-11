package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type FinanceRouter struct{}

func (r *FinanceRouter) InitFinanceRouter(Router *gin.RouterGroup) {
	financeRouter := Router.Group("carpool/finance").Use(middleware.OperationRecord())
	financeRouterWithoutRecord := Router.Group("carpool/finance")
	{
		financeRouter.POST("export", financeApi.ExportFinance)
	}
	{
		financeRouterWithoutRecord.GET("transaction/list", financeApi.ListTransactions)
		financeRouterWithoutRecord.GET("refund/list", financeApi.ListRefunds)
		financeRouterWithoutRecord.GET("summary", financeApi.GetSummary)
		financeRouterWithoutRecord.GET("abnormal/list", financeApi.ListAbnormalTransactions)
	}
}
