package response

import (
	autoModel "ride-hailing/admin-server/plugin/ai/model"
	sysModel "ride-hailing/admin-server/model/system"
)

type SysMcpListItem struct {
	autoModel.SysMcp
	ApiCount int64 `json:"apiCount"`
}

type SysMcpListResponse struct {
	List     []SysMcpListItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type SysMcpBoundApi struct {
	autoModel.SysMcpApi
	Api sysModel.SysApi `json:"api"`
}

type SysMcpDetailResponse struct {
	Mcp      autoModel.SysMcp  `json:"mcp"`
	Bindings []SysMcpBoundApi `json:"bindings"`
}
