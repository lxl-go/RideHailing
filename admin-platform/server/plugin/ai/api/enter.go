package api

import aiService "ride-hailing/admin-server/plugin/ai/service"

type ApiGroup struct {
	CliApi
	McpApi
}

var ApiGroupApp = new(ApiGroup)

var cliService = aiService.ServiceGroupApp.CliService
var mcpService = aiService.ServiceGroupApp.McpService
