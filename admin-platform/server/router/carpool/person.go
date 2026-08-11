package carpool

import (
	"ride-hailing/admin-server/middleware"

	"github.com/gin-gonic/gin"
)

type PersonRouter struct{}

func (r *PersonRouter) InitPersonRouter(Router *gin.RouterGroup) {
	personRouter := Router.Group("carpool/person").Use(middleware.OperationRecord())
	personRouterWithoutRecord := Router.Group("carpool/person")
	driverRouterWithoutRecord := Router.Group("carpool/driver")
	passengerRouterWithoutRecord := Router.Group("carpool/passenger")
	{
		personRouter.POST("", personApi.CreatePerson)
		personRouter.PUT(":id", personApi.UpdatePerson)
		personRouter.POST("roles", personApi.AssignRoles)
		personRouter.POST("batch/status", personApi.BatchStatus)
		personRouter.POST("driver/batch/delete", personApi.BatchDeleteDrivers)
		personRouter.POST("import/preview", personApi.PreviewImport)
		personRouter.POST("import/commit", personApi.CommitImport)
		personRouter.POST("export", personApi.ExportPersons)
	}
	{
		personRouterWithoutRecord.GET("list", personApi.ListPersons)
		personRouterWithoutRecord.GET("import/errors", personApi.ListImportErrors)
		personRouterWithoutRecord.GET(":id", personApi.GetPerson)
	}
	{
		driverRouterWithoutRecord.GET("stats", personApi.GetDriverStats)
		passengerRouterWithoutRecord.GET("stats", personApi.GetPassengerStats)
	}
}
