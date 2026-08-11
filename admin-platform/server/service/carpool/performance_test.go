package carpool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

func newPerformanceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.PerformanceReport{},
		&carpoolModel.PerformanceScenario{},
	))
	global.GVA_DB = db
	return db
}

func TestPerformanceServiceReportLifecycleValidationAndSummary(t *testing.T) {
	db := newPerformanceTestDB(t)
	service := PerformanceService{}
	ctx := context.Background()

	require.NoError(t, SeedPerformanceDefaults(db))
	require.NoError(t, SeedPerformanceDefaults(db))

	scenarios, err := service.ListScenarios(ctx, carpoolReq.PerformanceScenarioSearch{Scope: "driver"})
	require.NoError(t, err)
	require.Len(t, scenarios, 2)

	report, err := service.CreateReport(ctx, carpoolReq.SavePerformanceReportRequest{
		ReportNo:         "PERF-WO08-0001",
		Scenario:         "driver_location",
		TargetService:    "driver-api",
		Tool:             "k6",
		QPS:              5200,
		P50MS:            38,
		P90MS:            72,
		P99MS:            96,
		ErrorRate:        0.0001,
		GoroutinesBefore: 120,
		GoroutinesAfter:  123,
		HeapBeforeMB:     96,
		HeapAfterMB:      101,
		Verdict:          "PASS",
		ArtifactPath:     "docs/performance/reports/wo08-driver-location.sample.json",
		Notes:            "location reporting target met",
	})
	require.NoError(t, err)
	require.Equal(t, "WO-08", report.WorkorderNo)
	require.Equal(t, "driver_location", report.Scenario)

	_, err = service.CreateReport(ctx, carpoolReq.SavePerformanceReportRequest{
		ReportNo:      "PERF-WO08-0002",
		Scenario:      "passenger_ws_map",
		TargetService: "message-svc",
		Tool:          "k6",
		QPS:           1800,
		P50MS:         80,
		P90MS:         160,
		P99MS:         260,
		ErrorRate:     0.001,
		Verdict:       "WARN",
		ArtifactPath:  "docs/performance/reports/wo08-passenger-ws-map.sample.json",
		Notes:         "ws map p99 above target",
	})
	require.NoError(t, err)

	list, total, err := service.ListReports(ctx, carpoolReq.PerformanceReportSearch{Scenario: "driver_location", Verdict: "PASS"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "PERF-WO08-0001", list[0].ReportNo)

	summary, err := service.GetSummary(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 5, summary.TotalScenarios)
	require.EqualValues(t, 2, summary.TotalReports)
	require.EqualValues(t, 1, summary.PassReports)
	require.EqualValues(t, 1, summary.WarnReports)
	require.Equal(t, 50.0, summary.PassRate)
	require.Equal(t, "passenger_ws_map", summary.LatestReports[0].Scenario)
	require.True(t, summary.Runtime.NumGoroutine > 0)
	require.NotEmpty(t, summary.Runtime.GoVersion)

	_, err = service.CreateReport(ctx, carpoolReq.SavePerformanceReportRequest{
		ReportNo:      "PERF-WO08-BAD1",
		Scenario:      "driver_location",
		TargetService: "driver-api",
		Tool:          "k6",
		P99MS:         -1,
		ErrorRate:     0,
		Verdict:       "PASS",
		ArtifactPath:  "docs/performance/reports/bad.json",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "p99Ms must be non-negative")

	_, err = service.CreateReport(ctx, carpoolReq.SavePerformanceReportRequest{
		ReportNo:      "PERF-WO08-BAD2",
		Scenario:      "driver_location",
		TargetService: "driver-api",
		Tool:          "k6",
		P99MS:         90,
		ErrorRate:     1.5,
		Verdict:       "PASS",
		ArtifactPath:  "docs/performance/reports/bad.json",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "errorRate must be between 0 and 1")

	_, err = service.CreateReport(ctx, carpoolReq.SavePerformanceReportRequest{
		ReportNo:      "PERF-WO08-BAD3",
		Scenario:      "driver_location",
		TargetService: "driver-api",
		Tool:          "k6",
		P99MS:         90,
		ErrorRate:     0,
		Verdict:       "UNKNOWN",
		ArtifactPath:  "docs/performance/reports/bad.json",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "verdict must be PASS, WARN or FAIL")
}

func TestPerformanceServiceRuntimeSnapshotAndExport(t *testing.T) {
	newPerformanceTestDB(t)
	service := PerformanceService{}
	ctx := context.Background()

	snapshot, err := service.GetRuntimeSnapshot(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, snapshot.GoVersion)
	require.Greater(t, snapshot.NumCPU, 0)
	require.Greater(t, snapshot.NumGoroutine, 0)
	require.GreaterOrEqual(t, snapshot.HeapAllocMB, 0.0)
	require.GreaterOrEqual(t, snapshot.GCCycles, uint64(0))
	require.NoError(t, json.NewEncoder(&strings.Builder{}).Encode(snapshot))

	taskID := service.ExportTaskID(ctx)
	require.True(t, strings.HasPrefix(taskID, "WO08-EXPORT-"))
}
