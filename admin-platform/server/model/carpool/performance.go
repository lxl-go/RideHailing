package carpool

import "time"

type PerformanceReport struct {
	ID               uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:performance report id" json:"id"`
	ReportNo         string    `gorm:"column:report_no;type:varchar(64);not null;uniqueIndex:uk_performance_report_no;comment:report no" json:"reportNo"`
	WorkorderNo      string    `gorm:"column:workorder_no;type:varchar(16);not null;default:'WO-08';index:idx_performance_report_workorder;comment:workorder no" json:"workorderNo"`
	Scenario         string    `gorm:"column:scenario;type:varchar(64);not null;index:idx_performance_report_scenario;comment:scenario code" json:"scenario"`
	TargetService    string    `gorm:"column:target_service;type:varchar(64);not null;index:idx_performance_report_target;comment:target service" json:"targetService"`
	Tool             string    `gorm:"column:tool;type:varchar(32);not null;index:idx_performance_report_tool;comment:jmeter k6 wrk pprof trace manual" json:"tool"`
	QPS              float64   `gorm:"column:qps;type:decimal(10,2);not null;default:0;comment:qps" json:"qps"`
	P50MS            float64   `gorm:"column:p50_ms;type:decimal(10,2);not null;default:0;comment:p50 latency ms" json:"p50Ms"`
	P90MS            float64   `gorm:"column:p90_ms;type:decimal(10,2);not null;default:0;comment:p90 latency ms" json:"p90Ms"`
	P99MS            float64   `gorm:"column:p99_ms;type:decimal(10,2);not null;default:0;comment:p99 latency ms" json:"p99Ms"`
	ErrorRate        float64   `gorm:"column:error_rate;type:decimal(8,6);not null;default:0;comment:error rate 0-1" json:"errorRate"`
	GoroutinesBefore int       `gorm:"column:goroutines_before;type:int;not null;default:0;comment:goroutines before" json:"goroutinesBefore"`
	GoroutinesAfter  int       `gorm:"column:goroutines_after;type:int;not null;default:0;comment:goroutines after" json:"goroutinesAfter"`
	HeapBeforeMB     float64   `gorm:"column:heap_before_mb;type:decimal(10,2);not null;default:0;comment:heap before mb" json:"heapBeforeMb"`
	HeapAfterMB      float64   `gorm:"column:heap_after_mb;type:decimal(10,2);not null;default:0;comment:heap after mb" json:"heapAfterMb"`
	Verdict          string    `gorm:"column:verdict;type:varchar(16);not null;index:idx_performance_report_verdict;comment:PASS WARN FAIL" json:"verdict"`
	ArtifactPath     string    `gorm:"column:artifact_path;type:varchar(255);not null;default:'';comment:artifact path" json:"artifactPath"`
	Notes            string    `gorm:"column:notes;type:varchar(512);not null;default:'';comment:notes" json:"notes"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (PerformanceReport) TableName() string {
	return "performance_report"
}

type PerformanceScenario struct {
	ID                       uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:performance scenario id" json:"id"`
	Scenario                 string    `gorm:"column:scenario;type:varchar(64);not null;uniqueIndex:uk_performance_scenario;comment:scenario code" json:"scenario"`
	Name                     string    `gorm:"column:name;type:varchar(128);not null;comment:scenario name" json:"name"`
	Scope                    string    `gorm:"column:scope;type:varchar(32);not null;index:idx_performance_scenario_scope;comment:admin passenger driver backend" json:"scope"`
	TargetQPS                float64   `gorm:"column:target_qps;type:decimal(10,2);not null;default:0;comment:target qps" json:"targetQps"`
	TargetP99MS              float64   `gorm:"column:target_p99_ms;type:decimal(10,2);not null;default:0;comment:target p99 ms" json:"targetP99Ms"`
	MaxErrorRate             float64   `gorm:"column:max_error_rate;type:decimal(8,6);not null;default:0.001;comment:max error rate" json:"maxErrorRate"`
	MaxGoroutineDeltaPercent float64   `gorm:"column:max_goroutine_delta_percent;type:decimal(6,2);not null;default:5;comment:max goroutine delta percent" json:"maxGoroutineDeltaPercent"`
	Enabled                  bool      `gorm:"column:enabled;type:tinyint;not null;default:1;comment:enabled" json:"enabled"`
	CreatedAt                time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt                time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (PerformanceScenario) TableName() string {
	return "performance_scenario"
}
