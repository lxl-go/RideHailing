package router

import (
	api "ride-hailing/admin-server/api/v1"
	aiApi "ride-hailing/admin-server/plugin/ai/api"
)

type RouterGroup struct {
	CliRouter
	SkillsRouter
	McpRouter
	McpApiRouter
}

var (
	cliApi              = aiApi.ApiGroupApp.CliApi
	mcpApi              = aiApi.ApiGroupApp.McpApi
	skillsApi           = api.ApiGroupApp.SystemApiGroup.SkillsApi
	autoCodeTemplateApi = api.ApiGroupApp.SystemApiGroup.AutoCodeTemplateApi
)

var RouterGroupApp = new(RouterGroup)
