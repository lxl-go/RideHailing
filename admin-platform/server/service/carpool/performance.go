package carpool

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"gorm.io/gorm"
)

const (
	performanceWorkorderNo = "WO-08"
	verdictPass            = "PASS"
	verdictWarn            = "WARN"
	verdictFail            = "FAIL"
)

type PerformanceService struct{}

type RuntimeSnapshot struct {
	GoVersion     string  `json:"goVersion"`
	NumCPU        int     `json:"numCpu"`
	NumGoroutine  int     `json:"numGoroutine"`
	HeapAllocMB   float64 `json:"heapAllocMb"`
	HeapSysMB     float64 `json:"heapSysMb"`
	GCCycles      uint64  `json:"gcCycles"`
	LastGCPAUSEMS float64 `json:"lastGcPauseMs"`
	Warning       string  `json:"warning,omitempty"`
}

type PerformanceSummary struct {
	TotalScenarios int64                            `json:"totalScenarios"`
	TotalReports   int64                            `json:"totalReports"`
	PassReports    int64                            `json:"passReports"`
	WarnReports    int64                            `json:"warnReports"`
	FailReports    int64                            `json:"failReports"`
	PassRate       float64                          `json:"passRate"`
	Runtime        RuntimeSnapshot                  `json:"runtime"`
	LatestReports  []carpoolModel.PerformanceReport `json:"latestReports"`
}

// WO-08 performance: persist load-test or profiling reports for management-side acceptance.
func (s *PerformanceService) CreateReport(ctx context.Context, req carpoolReq.SavePerformanceReportRequest) (*carpoolModel.PerformanceReport, error) {
	if err := validatePerformanceReport(req); err != nil {
		return nil, err
	}
	report := &carpoolModel.PerformanceReport{
		ReportNo:         fallback(req.ReportNo, nextPerformanceReportNo()),
		WorkorderNo:      fallback(req.WorkorderNo, performanceWorkorderNo),
		Scenario:         strings.TrimSpace(req.Scenario),
		TargetService:    strings.TrimSpace(req.TargetService),
		Tool:             strings.TrimSpace(req.Tool),
		QPS:              roundAnalytics(req.QPS),
		P50MS:            roundAnalytics(req.P50MS),
		P90MS:            roundAnalytics(req.P90MS),
		P99MS:            roundAnalytics(req.P99MS),
		ErrorRate:        req.ErrorRate,
		GoroutinesBefore: req.GoroutinesBefore,
		GoroutinesAfter:  req.GoroutinesAfter,
		HeapBeforeMB:     roundAnalytics(req.HeapBeforeMB),
		HeapAfterMB:      roundAnalytics(req.HeapAfterMB),
		Verdict:          strings.ToUpper(strings.TrimSpace(req.Verdict)),
		ArtifactPath:     strings.TrimSpace(req.ArtifactPath),
		Notes:            strings.TrimSpace(req.Notes),
	}
	if err := global.GVA_DB.WithContext(ctx).Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (s *PerformanceService) ListReports(ctx context.Context, search carpoolReq.PerformanceReportSearch) ([]carpoolModel.PerformanceReport, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PerformanceReport{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("report_no LIKE ? OR target_service LIKE ? OR notes LIKE ?", keyword, keyword, keyword)
	}
	if search.Scenario != "" {
		db = db.Where("scenario = ?", search.Scenario)
	}
	if search.TargetService != "" {
		db = db.Where("target_service = ?", search.TargetService)
	}
	if search.Tool != "" {
		db = db.Where("tool = ?", search.Tool)
	}
	if search.Verdict != "" {
		db = db.Where("verdict = ?", strings.ToUpper(search.Verdict))
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.PerformanceReport
	if err := db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *PerformanceService) GetSummary(ctx context.Context) (*PerformanceSummary, error) {
	summary := &PerformanceSummary{}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PerformanceScenario{}).Where("enabled = ?", true).Count(&summary.TotalScenarios).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PerformanceReport{}).Count(&summary.TotalReports).Error; err != nil {
		return nil, err
	}
	for _, item := range []struct {
		verdict string
		target  *int64
	}{
		{verdictPass, &summary.PassReports},
		{verdictWarn, &summary.WarnReports},
		{verdictFail, &summary.FailReports},
	} {
		if err := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PerformanceReport{}).Where("verdict = ?", item.verdict).Count(item.target).Error; err != nil {
			return nil, err
		}
	}
	if summary.TotalReports > 0 {
		summary.PassRate = roundAnalytics(float64(summary.PassReports) * 100 / float64(summary.TotalReports))
	}
	if err := global.GVA_DB.WithContext(ctx).Order("created_at DESC, id DESC").Limit(5).Find(&summary.LatestReports).Error; err != nil {
		return nil, err
	}
	runtimeSnapshot, err := s.GetRuntimeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	summary.Runtime = *runtimeSnapshot
	return summary, nil
}

func (s *PerformanceService) ListScenarios(ctx context.Context, search carpoolReq.PerformanceScenarioSearch) ([]carpoolModel.PerformanceScenario, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PerformanceScenario{})
	if search.Scope != "" {
		db = db.Where("scope = ?", search.Scope)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	var list []carpoolModel.PerformanceScenario
	if err := db.Order("scope ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *PerformanceService) GetRuntimeSnapshot(ctx context.Context) (*RuntimeSnapshot, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	snapshot := &RuntimeSnapshot{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		HeapAllocMB:  bytesToMB(mem.HeapAlloc),
		HeapSysMB:    bytesToMB(mem.HeapSys),
		GCCycles:     uint64(mem.NumGC),
	}

	if mem.NumGC > 0 {
		lastPauseIndex := (mem.NumGC + 255) % 256
		snapshot.LastGCPAUSEMS = roundAnalytics(float64(mem.PauseNs[lastPauseIndex]) / 1e6)
	}
	return snapshot, nil
}

func (s *PerformanceService) ExportTaskID(ctx context.Context) string {
	return fmt.Sprintf("WO08-EXPORT-%d", time.Now().UnixNano())
}

func SeedPerformanceDefaults(db *gorm.DB) error {
	defaults := []carpoolModel.PerformanceScenario{
		{ID: 81001, Scenario: "admin_http", Name: "Admin API HTTP Smoke", Scope: "admin", TargetQPS: 1000, TargetP99MS: 200, MaxErrorRate: 0.001, MaxGoroutineDeltaPercent: 5, Enabled: true},
		{ID: 81002, Scenario: "passenger_ws_map", Name: "Passenger WebSocket Map Tracking", Scope: "passenger", TargetQPS: 5000, TargetP99MS: 200, MaxErrorRate: 0.001, MaxGoroutineDeltaPercent: 5, Enabled: true},
		{ID: 81003, Scenario: "driver_location", Name: "Driver Location Reporting", Scope: "driver", TargetQPS: 5000, TargetP99MS: 100, MaxErrorRate: 0.0001, MaxGoroutineDeltaPercent: 5, Enabled: true},
		{ID: 81004, Scenario: "driver_ws_dispatch", Name: "Driver Dispatch WebSocket", Scope: "driver", TargetQPS: 5000, TargetP99MS: 100, MaxErrorRate: 0.001, MaxGoroutineDeltaPercent: 5, Enabled: true},
		{ID: 81005, Scenario: "mixed", Name: "Mixed Ride-Hailing Flow", Scope: "backend", TargetQPS: 2000, TargetP99MS: 500, MaxErrorRate: 0.001, MaxGoroutineDeltaPercent: 5, Enabled: true},
	}
	for _, scenario := range defaults {
		var count int64
		if err := db.Model(&carpoolModel.PerformanceScenario{}).Where("scenario = ?", scenario.Scenario).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&scenario).Error; err != nil {
			return err
		}
	}
	return nil
}

func validatePerformanceReport(req carpoolReq.SavePerformanceReportRequest) error {
	if strings.TrimSpace(req.Scenario) == "" || strings.TrimSpace(req.TargetService) == "" || strings.TrimSpace(req.Tool) == "" {
		return errors.New("scenario, targetService and tool are required")
	}
	if req.P50MS < 0 {
		return errors.New("p50Ms must be non-negative")
	}
	if req.P90MS < 0 {
		return errors.New("p90Ms must be non-negative")
	}
	if req.P99MS < 0 {
		return errors.New("p99Ms must be non-negative")
	}
	if req.QPS < 0 {
		return errors.New("qps must be non-negative")
	}
	if req.ErrorRate < 0 || req.ErrorRate > 1 {
		return errors.New("errorRate must be between 0 and 1")
	}
	verdict := strings.ToUpper(strings.TrimSpace(req.Verdict))
	if verdict != verdictPass && verdict != verdictWarn && verdict != verdictFail {
		return errors.New("verdict must be PASS, WARN or FAIL")
	}
	return nil
}

func nextPerformanceReportNo() string {
	return fmt.Sprintf("PERF-WO08-%d", time.Now().UnixNano())
}

func bytesToMB(v uint64) float64 {
	return roundAnalytics(float64(v) / 1024 / 1024)
}
