package carpool

import "time"

type AiConversationLog struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:ai conversation log id" json:"id"`
	SessionID    string    `gorm:"column:session_id;type:varchar(96);not null;index:idx_ai_conversation_session;comment:session id" json:"sessionId"`
	UserID       uint64    `gorm:"column:user_id;not null;default:0;index:idx_ai_conversation_user;comment:user id" json:"userId"`
	UserRole     string    `gorm:"column:user_role;type:varchar(32);not null;index:idx_ai_conversation_role;comment:passenger driver admin" json:"userRole"`
	Question     string    `gorm:"column:question;type:varchar(1024);not null;comment:user question" json:"question"`
	Answer       string    `gorm:"column:answer;type:text;comment:ai answer or fallback answer" json:"answer"`
	Provider     string    `gorm:"column:provider;type:varchar(32);not null;default:'coze';index:idx_ai_conversation_provider;comment:provider" json:"provider"`
	Success      bool      `gorm:"column:success;type:tinyint;not null;default:0;index:idx_ai_conversation_success;comment:provider success" json:"success"`
	Fallback     bool      `gorm:"column:fallback;type:tinyint;not null;default:0;index:idx_ai_conversation_fallback;comment:fallback used" json:"fallback"`
	ErrorMessage string    `gorm:"column:error_message;type:varchar(512);not null;default:'';comment:masked error message" json:"errorMessage"`
	LatencyMS    int64     `gorm:"column:latency_ms;not null;default:0;comment:provider latency ms" json:"latencyMs"`
	TraceID      string    `gorm:"column:trace_id;type:varchar(64);not null;default:'';index:idx_ai_conversation_trace;comment:trace id" json:"traceId"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (AiConversationLog) TableName() string {
	return "ai_conversation_log"
}

type AiRoutePlanLog struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:ai route plan log id" json:"id"`
	RoutePlanNo  string    `gorm:"column:route_plan_no;type:varchar(64);not null;uniqueIndex:uk_ai_route_plan_no;index:idx_ai_route_plan_no;comment:route plan no" json:"routePlanNo"`
	SessionID    string    `gorm:"column:session_id;type:varchar(96);not null;index:idx_ai_route_plan_session;comment:session id" json:"sessionId"`
	UserRole     string    `gorm:"column:user_role;type:varchar(32);not null;index:idx_ai_route_plan_role;comment:passenger driver admin" json:"userRole"`
	Origin       string    `gorm:"column:origin;type:varchar(255);not null;comment:origin" json:"origin"`
	Destination  string    `gorm:"column:destination;type:varchar(255);not null;comment:destination" json:"destination"`
	City         string    `gorm:"column:city;type:varchar(64);not null;index:idx_ai_route_plan_city;comment:city" json:"city"`
	Weather      string    `gorm:"column:weather;type:varchar(128);not null;default:'';comment:weather" json:"weather"`
	Avoid        string    `gorm:"column:avoid;type:varchar(255);not null;default:'';comment:avoid options" json:"avoid"`
	Preference   string    `gorm:"column:preference;type:varchar(128);not null;default:'';comment:route preference" json:"preference"`
	RawResult    string    `gorm:"column:raw_result;type:text;comment:workflow raw result" json:"rawResult"`
	Provider     string    `gorm:"column:provider;type:varchar(32);not null;default:'coze';index:idx_ai_route_plan_provider;comment:provider" json:"provider"`
	Success      bool      `gorm:"column:success;type:tinyint;not null;default:0;index:idx_ai_route_plan_success;comment:provider success" json:"success"`
	Fallback     bool      `gorm:"column:fallback;type:tinyint;not null;default:0;index:idx_ai_route_plan_fallback;comment:fallback used" json:"fallback"`
	ErrorMessage string    `gorm:"column:error_message;type:varchar(512);not null;default:'';comment:masked error message" json:"errorMessage"`
	LatencyMS    int64     `gorm:"column:latency_ms;not null;default:0;comment:provider latency ms" json:"latencyMs"`
	TraceID      string    `gorm:"column:trace_id;type:varchar(64);not null;default:'';index:idx_ai_route_plan_trace;comment:trace id" json:"traceId"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (AiRoutePlanLog) TableName() string {
	return "ai_route_plan_log"
}

type AiFloodReport struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:ai flood report id" json:"id"`
	ReportNo     string    `gorm:"column:report_no;type:varchar(64);not null;uniqueIndex:uk_ai_flood_report_no;index:idx_ai_flood_report_no;comment:report no" json:"reportNo"`
	ReporterID   uint64    `gorm:"column:reporter_id;not null;default:0;index:idx_ai_flood_report_reporter;comment:reporter id" json:"reporterId"`
	ReporterRole string    `gorm:"column:reporter_role;type:varchar(32);not null;index:idx_ai_flood_report_role;comment:passenger driver admin" json:"reporterRole"`
	City         string    `gorm:"column:city;type:varchar(64);not null;index:idx_ai_flood_report_city;comment:city" json:"city"`
	LocationText string    `gorm:"column:location_text;type:varchar(255);not null;comment:location text" json:"locationText"`
	PhotoURL     string    `gorm:"column:photo_url;type:varchar(512);not null;default:'';comment:photo url" json:"photoUrl"`
	DepthCM      float64   `gorm:"column:depth_cm;type:decimal(8,2);not null;default:0;comment:depth cm" json:"depthCm"`
	Confidence   float64   `gorm:"column:confidence;type:decimal(6,2);not null;default:0;comment:confidence 0-100" json:"confidence"`
	AuditStatus  string    `gorm:"column:audit_status;type:varchar(32);not null;index:idx_ai_flood_report_audit;comment:audit status" json:"auditStatus"`
	AuditRemark  string    `gorm:"column:audit_remark;type:varchar(512);not null;default:'';comment:audit remark" json:"auditRemark"`
	TraceID      string    `gorm:"column:trace_id;type:varchar(64);not null;default:'';index:idx_ai_flood_report_trace;comment:trace id" json:"traceId"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (AiFloodReport) TableName() string {
	return "ai_flood_report"
}

type AiFallbackTemplate struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:ai fallback template id" json:"id"`
	Scope     string    `gorm:"column:scope;type:varchar(32);not null;index:idx_ai_fallback_scope;comment:chat route" json:"scope"`
	UserRole  string    `gorm:"column:user_role;type:varchar(32);not null;index:idx_ai_fallback_role;comment:passenger driver admin all" json:"userRole"`
	Content   string    `gorm:"column:content;type:varchar(1024);not null;comment:fallback content" json:"content"`
	Enabled   bool      `gorm:"column:enabled;type:tinyint;not null;default:1;index:idx_ai_fallback_enabled;comment:enabled" json:"enabled"`
	TraceID   string    `gorm:"column:trace_id;type:varchar(64);not null;default:'';index:idx_ai_fallback_trace;comment:trace id" json:"traceId"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (AiFallbackTemplate) TableName() string {
	return "ai_fallback_template"
}
