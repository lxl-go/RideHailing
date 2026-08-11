package carpool

import (
	api "ride-hailing/admin-server/api/v1"
)

type RouterGroup struct {
	ReviewRouter
	TripRouter
	ShuttleRouter
	FinanceRouter
	OrderRouter
	PersonRouter
	AnalyticsRouter
	MarketingRouter
	PerformanceRouter
	AIRouter
	DispatchRouter
}

var (
	reviewApi      = api.ApiGroupApp.CarpoolApiGroup.ReviewApi
	tripApi        = api.ApiGroupApp.CarpoolApiGroup.TripApi
	shuttleApi     = api.ApiGroupApp.CarpoolApiGroup.ShuttleApi
	financeApi     = api.ApiGroupApp.CarpoolApiGroup.FinanceApi
	orderApi       = api.ApiGroupApp.CarpoolApiGroup.OrderApi
	personApi      = api.ApiGroupApp.CarpoolApiGroup.PersonApi
	analyticsApi   = api.ApiGroupApp.CarpoolApiGroup.AnalyticsApi
	marketingApi   = api.ApiGroupApp.CarpoolApiGroup.MarketingApi
	performanceApi = api.ApiGroupApp.CarpoolApiGroup.PerformanceApi
	aiApi          = api.ApiGroupApp.CarpoolApiGroup.AIApi
	dispatchApi    = api.ApiGroupApp.CarpoolApiGroup.DispatchApi
)
