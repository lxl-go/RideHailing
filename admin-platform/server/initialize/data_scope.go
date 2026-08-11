package initialize

import (
	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils/datascope"
)

// RegisterDataScopeCallbacks 为主库及所有多库连接注册数据权限 GORM 回调,
// 并接上审计事件异步落表(sys_data_access_logs)。
// 需在 DB(含 GVA_DBList)初始化完成后调用。
func RegisterDataScopeCallbacks() {
	datascope.RegisterCallbacks(global.GVA_DB)
	for _, db := range global.GVA_DBList {
		datascope.RegisterCallbacks(db)
	}
	auditSvc := &service.ServiceGroupApp.SystemServiceGroup.DataAccessLogService
	datascope.SetAuditHook(auditSvc.Enqueue)
	auditSvc.StartWriter()
}
