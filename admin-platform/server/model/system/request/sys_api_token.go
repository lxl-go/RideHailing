package request

import (
	"ride-hailing/admin-server/model/common/request"
	"ride-hailing/admin-server/model/system"
)

type SysApiTokenSearch struct {
	system.SysApiToken
	request.PageInfo
    Status *bool `json:"status" form:"status"`
}
