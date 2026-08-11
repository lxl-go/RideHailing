package request

import commonReq "ride-hailing/admin-server/model/common/request"

type AIChatRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
	Text      string `json:"text" binding:"required"`
	UserID    uint64 `json:"userId"`
	UserRole  string `json:"userRole" binding:"required"`
}

type RainRouteRequest struct {
	SessionID   string `json:"sessionId"`
	Origin      string `json:"origin" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	City        string `json:"city" binding:"required"`
	Weather     string `json:"weather"`
	Avoid       string `json:"avoid"`
	Preference  string `json:"preference"`
	UserRole    string `json:"userRole" binding:"required"`
}

type ChatWithRouteRequest struct {
	Chat  AIChatRequest    `json:"chat"`
	Route RainRouteRequest `json:"route"`
}

type FloodReportRequest struct {
	ReporterID   uint64  `json:"reporterId"`
	ReporterRole string  `json:"reporterRole" binding:"required"`
	City         string  `json:"city" binding:"required"`
	LocationText string  `json:"locationText" binding:"required"`
	PhotoURL     string  `json:"photoUrl"`
	DepthCM      float64 `json:"depthCm"`
	Confidence   float64 `json:"confidence"`
}

type AIConversationLogSearch struct {
	commonReq.PageInfo
	SessionID string `json:"sessionId" form:"sessionId"`
	UserRole  string `json:"userRole" form:"userRole"`
	Success   *bool  `json:"success" form:"success"`
	Fallback  *bool  `json:"fallback" form:"fallback"`
}

type AIRoutePlanLogSearch struct {
	commonReq.PageInfo
	RoutePlanNo string `json:"routePlanNo" form:"routePlanNo"`
	SessionID   string `json:"sessionId" form:"sessionId"`
	UserRole    string `json:"userRole" form:"userRole"`
	City        string `json:"city" form:"city"`
	Success     *bool  `json:"success" form:"success"`
	Fallback    *bool  `json:"fallback" form:"fallback"`
}

type AIFloodReportSearch struct {
	commonReq.PageInfo
	ReportNo    string `json:"reportNo" form:"reportNo"`
	City        string `json:"city" form:"city"`
	AuditStatus string `json:"auditStatus" form:"auditStatus"`
}

type AIFloodReportAuditRequest struct {
	ReportNo    string `json:"reportNo" binding:"required"`
	AuditStatus string `json:"auditStatus" binding:"required"`
	AuditRemark string `json:"auditRemark"`
}
