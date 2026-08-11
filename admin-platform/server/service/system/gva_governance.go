package system

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	systemModel "ride-hailing/admin-server/model/system"
)

const gvaGovernanceExportPrefix = "WO09-GVA-EXPORT-"

type GvaGovernanceService struct{}

type GvaGovernanceSummary struct {
	Route      GvaRouteSnapshot      `json:"route"`
	Audit      GvaAuditSnapshot      `json:"audit"`
	Datasource GvaDatasourceSnapshot `json:"datasource"`
	TimedTask  GvaTimedTaskSnapshot  `json:"timedTask"`
	Warnings   []string              `json:"warnings"`
}

type GvaRouteSnapshot struct {
	TotalMenus      int64    `json:"totalMenus"`
	HiddenMenus     int64    `json:"hiddenMenus"`
	DefaultMenus    int64    `json:"defaultMenus"`
	DuplicateNames  int      `json:"duplicateNames"`
	RouteVersion    string   `json:"routeVersion"`
	WhitelistStatus string   `json:"whitelistStatus"`
	AllowedPrefixes []string `json:"allowedPrefixes"`
	Warnings        []string `json:"warnings"`
}

type GvaAuditSnapshot struct {
	DataAccessLogs         int64                            `json:"dataAccessLogs"`
	BlockedWrites          int64                            `json:"blockedWrites"`
	NoIdentityEvents       int64                            `json:"noIdentityEvents"`
	OperationRecords       int64                            `json:"operationRecords"`
	MissingTraceRecords    int64                            `json:"missingTraceRecords"`
	RecentDataAccessLogs   []systemModel.SysDataAccessLog   `json:"recentDataAccessLogs"`
	RecentOperationRecords []systemModel.SysOperationRecord `json:"recentOperationRecords"`
	Warnings               []string                         `json:"warnings"`
}

type GvaDatasourceSnapshot struct {
	DBType       string         `json:"dbType"`
	ActiveDBName string         `json:"activeDbName"`
	Healthy      bool           `json:"healthy"`
	Warning      string         `json:"warning,omitempty"`
	Pool         DBPoolSnapshot `json:"pool"`
}

type DBPoolSnapshot struct {
	OpenConnections int `json:"openConnections"`
	InUse           int `json:"inUse"`
	Idle            int `json:"idle"`
}

type GvaTimedTaskSnapshot struct {
	TotalTasks    int64                         `json:"totalTasks"`
	EnabledTasks  int64                         `json:"enabledTasks"`
	DisabledTasks int64                         `json:"disabledTasks"`
	InvalidTasks  []TimedTaskIssue              `json:"invalidTasks"`
	LatestLogs    []systemModel.SysTimedTaskLog `json:"latestLogs"`
	Warnings      []string                      `json:"warnings"`
}

type TimedTaskIssue struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func (s *GvaGovernanceService) GetGovernanceSummary(ctx context.Context) (*GvaGovernanceSummary, error) {
	route, err := s.GetRouteSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	audit, err := s.GetAuditSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	datasource, err := s.GetDatasourceSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	timedTask, err := s.GetTimedTaskSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	warnings := append([]string{}, route.Warnings...)
	warnings = append(warnings, audit.Warnings...)
	if datasource.Warning != "" {
		warnings = append(warnings, datasource.Warning)
	}
	warnings = append(warnings, timedTask.Warnings...)
	return &GvaGovernanceSummary{Route: *route, Audit: *audit, Datasource: *datasource, TimedTask: *timedTask, Warnings: warnings}, nil
}

func (s *GvaGovernanceService) GetRouteSnapshot(ctx context.Context) (*GvaRouteSnapshot, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&systemModel.SysBaseMenu{})
	snapshot := &GvaRouteSnapshot{AllowedPrefixes: []string{"view/", "plugin/", "layout/", "routerHolder"}}
	if err := db.Count(&snapshot.TotalMenus).Error; err != nil {
		return nil, err
	}
	if err := db.Where("hidden = ?", true).Count(&snapshot.HiddenMenus).Error; err != nil {
		return nil, err
	}
	if err := db.Where("default_menu = ?", true).Count(&snapshot.DefaultMenus).Error; err != nil {
		return nil, err
	}
	var menus []systemModel.SysBaseMenu
	if err := global.GVA_DB.WithContext(ctx).Find(&menus).Error; err != nil {
		return nil, err
	}
	nameCount := map[string]int{}
	var latestUnix int64
	for _, menu := range menus {
		nameCount[menu.Name]++
		if menu.UpdatedAt.Unix() > latestUnix {
			latestUnix = menu.UpdatedAt.Unix()
		}
		if strings.TrimSpace(menu.Component) == "" {
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("菜单 %s 缺少 component", menu.Name))
			continue
		}
		if !allowedGvaComponent(menu.Component, snapshot.AllowedPrefixes) {
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("菜单 %s component 前缀未在白名单: %s", menu.Name, menu.Component))
		}
	}
	for name, count := range nameCount {
		if name != "" && count > 1 {
			snapshot.DuplicateNames++
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("路由名称重复: %s", name))
		}
	}
	snapshot.RouteVersion = fmt.Sprintf("menus-%d-%d", snapshot.TotalMenus, latestUnix)
	if len(snapshot.Warnings) == 0 {
		snapshot.WhitelistStatus = "PASS"
	} else {
		snapshot.WhitelistStatus = "WARN"
	}
	return snapshot, nil
}

func (s *GvaGovernanceService) GetAuditSnapshot(ctx context.Context) (*GvaAuditSnapshot, error) {
	snapshot := &GvaAuditSnapshot{}
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&systemModel.SysDataAccessLog{}).Count(&snapshot.DataAccessLogs).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysDataAccessLog{}).Where("event_type = ?", "blocked_write").Count(&snapshot.BlockedWrites).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysDataAccessLog{}).Where("event_type = ?", "no_identity").Count(&snapshot.NoIdentityEvents).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysOperationRecord{}).Count(&snapshot.OperationRecords).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysOperationRecord{}).Where("request_id = '' OR trace_id = ''").Count(&snapshot.MissingTraceRecords).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id desc").Limit(5).Find(&snapshot.RecentDataAccessLogs).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id desc").Limit(5).Find(&snapshot.RecentOperationRecords).Error; err != nil {
		return nil, err
	}
	if snapshot.MissingTraceRecords > 0 {
		snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("%d 条操作记录缺少 request_id 或 trace_id", snapshot.MissingTraceRecords))
	}
	return snapshot, nil
}

func (s *GvaGovernanceService) GetDatasourceSnapshot(ctx context.Context) (*GvaDatasourceSnapshot, error) {
	snapshot := &GvaDatasourceSnapshot{DBType: global.GVA_CONFIG.System.DbType}
	if global.GVA_ACTIVE_DBNAME != nil {
		snapshot.ActiveDBName = *global.GVA_ACTIVE_DBNAME
	}
	sqlDB, err := global.GVA_DB.WithContext(ctx).DB()
	if err != nil {
		snapshot.Warning = err.Error()
		return snapshot, nil
	}
	stats := sqlDB.Stats()
	snapshot.Pool = DBPoolSnapshot{OpenConnections: stats.OpenConnections, InUse: stats.InUse, Idle: stats.Idle}
	if err := sqlDB.PingContext(ctx); err != nil {
		snapshot.Warning = err.Error()
		return snapshot, nil
	}
	snapshot.Healthy = true
	return snapshot, nil
}

func (s *GvaGovernanceService) GetTimedTaskSnapshot(ctx context.Context) (*GvaTimedTaskSnapshot, error) {
	snapshot := &GvaTimedTaskSnapshot{}
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&systemModel.SysTimedTask{}).Count(&snapshot.TotalTasks).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysTimedTask{}).Where("enabled = ?", true).Count(&snapshot.EnabledTasks).Error; err != nil {
		return nil, err
	}
	snapshot.DisabledTasks = snapshot.TotalTasks - snapshot.EnabledTasks
	var tasks []systemModel.SysTimedTask
	if err := db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	for _, task := range tasks {
		if len(task.Params) > 0 && !json.Valid(task.Params) {
			snapshot.InvalidTasks = append(snapshot.InvalidTasks, TimedTaskIssue{ID: task.ID, Name: task.Name, Reason: "params 不是合法 JSON"})
		}
		if len(task.HttpHeader) > 0 && !json.Valid(task.HttpHeader) {
			snapshot.InvalidTasks = append(snapshot.InvalidTasks, TimedTaskIssue{ID: task.ID, Name: task.Name, Reason: "httpHeader 不是合法 JSON"})
		}
		if strings.TrimSpace(task.ExecutorType) == "" {
			snapshot.InvalidTasks = append(snapshot.InvalidTasks, TimedTaskIssue{ID: task.ID, Name: task.Name, Reason: "executorType 为空"})
		}
	}
	if err := db.Order("id desc").Limit(5).Find(&snapshot.LatestLogs).Error; err != nil {
		return nil, err
	}
	if len(snapshot.InvalidTasks) > 0 {
		snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("%d 个定时任务参数需要治理", len(snapshot.InvalidTasks)))
	}
	return snapshot, nil
}

func (s *GvaGovernanceService) ExportGovernance(ctx context.Context) string {
	return fmt.Sprintf("%s%d", gvaGovernanceExportPrefix, time.Now().UnixNano())
}

func allowedGvaComponent(component string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(component, prefix) {
			return true
		}
	}
	return false
}
