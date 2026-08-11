package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type TripRouter struct{}

func (r *TripRouter) InitTripRouter(Router *gin.RouterGroup) {
	tripRouter := Router.Group("carpool/trip").Use(middleware.OperationRecord())
	tripRouterWithoutRecord := Router.Group("carpool/trip")
	{
		tripRouter.POST(":id/review", tripApi.ReviewTrip)
		tripRouter.POST(":id/deactivate", tripApi.DeactivateTrip)
	}
	{
		tripRouterWithoutRecord.GET("list", tripApi.ListTrips)
		tripRouterWithoutRecord.GET(":id", tripApi.GetTrip)
	}
}
