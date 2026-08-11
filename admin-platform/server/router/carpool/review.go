package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type ReviewRouter struct{}

func (r *ReviewRouter) InitReviewRouter(Router *gin.RouterGroup) {
	reviewRouter := Router.Group("carpool/review").Use(middleware.OperationRecord())
	reviewRouterWithoutRecord := Router.Group("carpool/review")
	{
		reviewRouter.POST(":id/approve", reviewApi.ApproveAudit)
		reviewRouter.POST(":id/reject", reviewApi.RejectAudit)
		reviewRouter.POST("vehicle/:id/action", reviewApi.HandleVehicleReview)
	}
	{
		reviewRouterWithoutRecord.GET("list", reviewApi.ListAudits)
		reviewRouterWithoutRecord.GET("vehicle/list", reviewApi.ListVehicleReviews)
		reviewRouterWithoutRecord.GET(":id", reviewApi.GetAudit)
	}
}
