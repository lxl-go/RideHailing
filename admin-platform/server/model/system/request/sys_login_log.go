package request

import (
	"ride-hailing/admin-server/model/common/request"
	"ride-hailing/admin-server/model/system"
)

type SysLoginLogSearch struct {
	system.SysLoginLog
	request.PageInfo
}
