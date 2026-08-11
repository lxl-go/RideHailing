package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type AIRouter struct{}

func (r *AIRouter) InitAIRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	aiRouter := Router.Group("carpool/ai").Use(middleware.OperationRecord())
	aiRouterWithoutRecord := Router.Group("carpool/ai")
	travelPublicRouter := PublicRouter.Group("travel")
	{
		aiRouter.POST("chat", aiApi.Chat)
		aiRouter.POST("rain-route", aiApi.PlanRainRoute)
		aiRouter.POST("chat-with-route", aiApi.ChatWithRainRoute)
		aiRouter.POST("flood-report", aiApi.ReportFlooding)
		aiRouter.POST("flood-report/audit", aiApi.AuditFloodReport)
		aiRouter.POST("export", aiApi.ExportAI)
	}
	{
		aiRouterWithoutRecord.GET("summary", aiApi.GetAISummary)
		aiRouterWithoutRecord.GET("conversation/list", aiApi.ListConversationLogs)
		aiRouterWithoutRecord.GET("route-plan/list", aiApi.ListRoutePlanLogs)
		aiRouterWithoutRecord.GET("flood-report/list", aiApi.ListFloodReports)
	}
	{
		travelPublicRouter.POST("route-info", aiApi.GetTravelRouteInfo)
	}
}
