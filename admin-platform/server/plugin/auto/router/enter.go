package router

import (
	api "ride-hailing/admin-server/api/v1"
)

type RouterGroup struct {
	AutoCodeRouter
}

var (
	autoCodeApi         = api.ApiGroupApp.SystemApiGroup.AutoCodeApi
	autoCodePluginApi   = api.ApiGroupApp.SystemApiGroup.AutoCodePluginApi
	autocodeHistoryApi  = api.ApiGroupApp.SystemApiGroup.AutoCodeHistoryApi
	autoCodePackageApi  = api.ApiGroupApp.SystemApiGroup.AutoCodePackageApi
	autoCodeTemplateApi = api.ApiGroupApp.SystemApiGroup.AutoCodeTemplateApi
)

var RouterGroupApp = new(RouterGroup)
