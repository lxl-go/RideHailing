package request

import commonReq "ride-hailing/admin-server/model/common/request"

type PerformanceReportSearch struct {
	commonReq.PageInfo
	Scenario      string `json:"scenario" form:"scenario"`
	TargetService string `json:"targetService" form:"targetService"`
	Tool          string `json:"tool" form:"tool"`
	Verdict       string `json:"verdict" form:"verdict"`
}

type SavePerformanceReportRequest struct {
	ReportNo         string  `json:"reportNo"`
	WorkorderNo      string  `json:"workorderNo"`
	Scenario         string  `json:"scenario" binding:"required"`
	TargetService    string  `json:"targetService" binding:"required"`
	Tool             string  `json:"tool" binding:"required"`
	QPS              float64 `json:"qps"`
	P50MS            float64 `json:"p50Ms"`
	P90MS            float64 `json:"p90Ms"`
	P99MS            float64 `json:"p99Ms"`
	ErrorRate        float64 `json:"errorRate"`
	GoroutinesBefore int     `json:"goroutinesBefore"`
	GoroutinesAfter  int     `json:"goroutinesAfter"`
	HeapBeforeMB     float64 `json:"heapBeforeMb"`
	HeapAfterMB      float64 `json:"heapAfterMb"`
	Verdict          string  `json:"verdict" binding:"required"`
	ArtifactPath     string  `json:"artifactPath"`
	Notes            string  `json:"notes"`
}

type PerformanceScenarioSearch struct {
	Scope   string `json:"scope" form:"scope"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}
