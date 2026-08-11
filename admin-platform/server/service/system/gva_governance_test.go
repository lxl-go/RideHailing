package system

import (
	"context"
	"strings"
	"testing"

	"ride-hailing/admin-server/config"
	"ride-hailing/admin-server/global"
	systemModel "ride-hailing/admin-server/model/system"

	"github.com/glebarez/sqlite"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupGvaGovernanceTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&systemModel.SysBaseMenu{},
		&systemModel.SysAuthorityMenu{},
		&systemModel.SysDataAccessLog{},
		&systemModel.SysOperationRecord{},
		&systemModel.SysTimedTask{},
		&systemModel.SysTimedTaskLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	global.GVA_DB = db
	global.GVA_CONFIG.System.DbType = "sqlite"
	global.GVA_CONFIG.Sqlite = config.Sqlite{GeneralDB: config.GeneralDB{Dbname: ":memory:"}}
	global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Sqlite.Dbname
}

func TestGvaGovernanceServiceSummarySnapshotsAndWarnings(t *testing.T) {
	setupGvaGovernanceTestDB(t)
	ctx := context.Background()
	if err := global.GVA_DB.Create(&[]systemModel.SysBaseMenu{
		{Name: "RideHailingWorkorders", Path: "workorders", Component: "view/rideHailing/workorders/index.vue", Meta: systemModel.Meta{Title: "工单管理"}},
		{Name: "RideHailingWorkorders", Path: "duplicate", Component: "view/rideHailing/duplicate.vue", Meta: systemModel.Meta{Title: "重复路由"}},
		{Name: "UnsafeMenu", Path: "unsafe", Component: "unknown/unsafe.vue", Meta: systemModel.Meta{Title: "未知组件"}},
	}).Error; err != nil {
		t.Fatalf("seed menus: %v", err)
	}
	if err := global.GVA_DB.Create(&[]systemModel.SysDataAccessLog{
		{EventType: "blocked_write", TargetTable: "order_main", Operation: "update", RequestID: "req-1", Path: "/carpool/order/refund"},
		{EventType: "no_identity", TargetTable: "carpool_trip", Operation: "query", Path: "/carpool/trip/list"},
	}).Error; err != nil {
		t.Fatalf("seed data access logs: %v", err)
	}
	if err := global.GVA_DB.Create(&systemModel.SysOperationRecord{Method: "POST", Path: "/system/menu/addBaseMenu", Status: 200, RequestID: "req-op", TraceID: "trace-op", LatencyMs: 42}).Error; err != nil {
		t.Fatalf("seed operation record: %v", err)
	}
	if err := global.GVA_DB.Create(&[]systemModel.SysTimedTask{
		{Name: "valid", Spec: "@daily", ExecutorType: systemModel.TimedTaskExecutorHTTP, HttpUrl: "https://example.com", HttpMethod: "POST", HttpHeader: datatypes.JSON(`{"X-Test":"ok"}`), Enabled: true},
		{Name: "invalid-json", Spec: "@daily", ExecutorType: systemModel.TimedTaskExecutorMethod, MethodName: "unknown", Params: datatypes.JSON(`{"bad"`), Enabled: false},
	}).Error; err != nil {
		t.Fatalf("seed timed tasks: %v", err)
	}

	svc := GvaGovernanceService{}
	summary, err := svc.GetGovernanceSummary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Route.TotalMenus != 3 || summary.Audit.DataAccessLogs != 2 || summary.TimedTask.TotalTasks != 2 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.Route.DuplicateNames != 1 {
		t.Fatalf("expected one duplicate route name, got %d", summary.Route.DuplicateNames)
	}
	if len(summary.Route.Warnings) == 0 || len(summary.TimedTask.InvalidTasks) == 0 {
		t.Fatalf("expected route and timed task warnings: %+v", summary)
	}
	if !summary.Datasource.Healthy {
		t.Fatalf("expected sqlite datasource healthy: %+v", summary.Datasource)
	}
}

func TestGvaGovernanceServiceExportTaskID(t *testing.T) {
	setupGvaGovernanceTestDB(t)
	taskID := (&GvaGovernanceService{}).ExportGovernance(context.Background())
	if !strings.HasPrefix(taskID, "WO09-GVA-EXPORT-") {
		t.Fatalf("unexpected task id: %s", taskID)
	}
}
