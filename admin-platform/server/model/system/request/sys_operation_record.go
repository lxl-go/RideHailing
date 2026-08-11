package request

import (
	"ride-hailing/admin-server/model/common/request"
	"ride-hailing/admin-server/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.PageInfo
}
