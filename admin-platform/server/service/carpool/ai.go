package carpool

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/utils/logger"
	"ride-hailing/pkg/cozex"

	"gorm.io/gorm"
)

const (
	aiProviderCoze                 = "coze"
	aiAuditPendingManualReview     = "pending_manual_review"
	aiAuditAutoConfirmed           = "auto_confirmed"
	aiDefaultFloodReviewConfidence = 80
	cozeTravelBotTokenEnv          = "COZE_TRAVEL_BOT_TOKEN"
	cozeRainRouteWorkflowTokenEnv  = "COZE_RAIN_ROUTE_WORKFLOW_TOKEN"
)

type AIProvider interface {
	Chat(ctx context.Context, req carpoolReq.AIChatRequest) (string, error)
	PlanRainRoute(ctx context.Context, req carpoolReq.RainRouteRequest) (string, error)
}

type AIService struct {
	provider AIProvider
}

type AIChatResult struct {
	SessionID string `json:"sessionId"`
	Answer    string `json:"answer"`
	Success   bool   `json:"success"`
	Fallback  bool   `json:"fallback"`
	LatencyMS int64  `json:"latencyMs"`
	TraceID   string `json:"traceId"`
}

type AIRoutePlanResult struct {
	RoutePlanNo string `json:"routePlanNo"`
	RawResult   string `json:"rawResult"`
	Success     bool   `json:"success"`
	Fallback    bool   `json:"fallback"`
	LatencyMS   int64  `json:"latencyMs"`
	TraceID     string `json:"traceId"`
}

type AISummary struct {
	TotalCalls    int64   `json:"totalCalls"`
	SuccessCalls  int64   `json:"successCalls"`
	FallbackCalls int64   `json:"fallbackCalls"`
	RoutePlans    int64   `json:"routePlans"`
	FloodReports  int64   `json:"floodReports"`
	AvgLatencyMS  float64 `json:"avgLatencyMs"`
	PendingFloods int64   `json:"pendingFloods"`
	LatestTraceID string  `json:"latestTraceId"`
}

func NewAIServiceWithProvider(provider AIProvider) *AIService {
	return &AIService{provider: provider}
}

func (s *AIService) providerOrDefault() AIProvider {
	if s != nil && s.provider != nil {
		return s.provider
	}
	return cozeAIProvider{client: cozex.NewClient(cozex.Config{
		TravelBotToken:         os.Getenv(cozeTravelBotTokenEnv),
		RainRouteWorkflowToken: os.Getenv(cozeRainRouteWorkflowTokenEnv),
	}, nil)}
}

func (s *AIService) Chat(ctx context.Context, req carpoolReq.AIChatRequest) (*AIChatResult, error) {
	if err := validateAIChat(req); err != nil {
		return nil, err
	}
	traceID := traceIDFromContext(ctx)
	start := time.Now()
	answer, providerErr := s.providerOrDefault().Chat(ctx, req)
	latencyMS := time.Since(start).Milliseconds()
	success := providerErr == nil
	fallback := false
	if providerErr != nil {
		fallback = true
		answer = s.fallbackText(ctx, "chat", req.UserRole)
	}

	result := &AIChatResult{
		SessionID: strings.TrimSpace(req.SessionID),
		Answer:    answer,
		Success:   success,
		Fallback:  fallback,
		LatencyMS: latencyMS,
		TraceID:   traceID,
	}
	log := &carpoolModel.AiConversationLog{
		SessionID:    result.SessionID,
		UserID:       req.UserID,
		UserRole:     normalizeAIRole(req.UserRole),
		Question:     strings.TrimSpace(req.Text),
		Answer:       answer,
		Provider:     aiProviderCoze,
		Success:      success,
		Fallback:     fallback,
		ErrorMessage: safeError(providerErr),
		LatencyMS:    latencyMS,
		TraceID:      traceID,
	}
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(log).Error
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AIService) PlanRainRoute(ctx context.Context, req carpoolReq.RainRouteRequest) (*AIRoutePlanResult, error) {
	if err := validateRainRoute(req); err != nil {
		return nil, err
	}
	traceID := traceIDFromContext(ctx)
	routePlanNo := nextAIRoutePlanNo()
	start := time.Now()
	raw, providerErr := s.providerOrDefault().PlanRainRoute(ctx, req)
	latencyMS := time.Since(start).Milliseconds()
	success := providerErr == nil
	fallback := false
	if providerErr != nil {
		fallback = true
		raw = s.fallbackText(ctx, "route", req.UserRole)
	}

	result := &AIRoutePlanResult{
		RoutePlanNo: routePlanNo,
		RawResult:   raw,
		Success:     success,
		Fallback:    fallback,
		LatencyMS:   latencyMS,
		TraceID:     traceID,
	}
	log := &carpoolModel.AiRoutePlanLog{
		RoutePlanNo:  routePlanNo,
		SessionID:    strings.TrimSpace(req.SessionID),
		UserRole:     normalizeAIRole(req.UserRole),
		Origin:       strings.TrimSpace(req.Origin),
		Destination:  strings.TrimSpace(req.Destination),
		City:         strings.TrimSpace(req.City),
		Weather:      strings.TrimSpace(req.Weather),
		Avoid:        strings.TrimSpace(req.Avoid),
		Preference:   strings.TrimSpace(req.Preference),
		RawResult:    raw,
		Provider:     aiProviderCoze,
		Success:      success,
		Fallback:     fallback,
		ErrorMessage: safeError(providerErr),
		LatencyMS:    latencyMS,
		TraceID:      traceID,
	}
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(log).Error
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AIService) ChatWithRainRoute(ctx context.Context, req carpoolReq.ChatWithRouteRequest) (*AIChatResult, error) {
	routeResult, err := s.PlanRainRoute(ctx, req.Route)
	if err != nil {
		return nil, err
	}
	chat := req.Chat
	if strings.TrimSpace(chat.SessionID) == "" {
		chat.SessionID = req.Route.SessionID
	}
	if strings.TrimSpace(chat.UserRole) == "" {
		chat.UserRole = req.Route.UserRole
	}
	chat.Text = strings.TrimSpace(chat.Text)
	if chat.Text != "" {
		chat.Text = chat.Text + "\nRoute context: " + routeResult.RawResult
	}
	return s.Chat(ctx, chat)
}

func (s *AIService) ReportFlooding(ctx context.Context, req carpoolReq.FloodReportRequest) (*carpoolModel.AiFloodReport, error) {
	if err := validateFloodReport(req); err != nil {
		return nil, err
	}
	auditStatus := aiAuditAutoConfirmed
	if req.Confidence < aiDefaultFloodReviewConfidence {
		auditStatus = aiAuditPendingManualReview
	}
	report := &carpoolModel.AiFloodReport{
		ReportNo:     nextAIFloodReportNo(),
		ReporterID:   req.ReporterID,
		ReporterRole: normalizeAIRole(req.ReporterRole),
		City:         strings.TrimSpace(req.City),
		LocationText: strings.TrimSpace(req.LocationText),
		PhotoURL:     strings.TrimSpace(req.PhotoURL),
		DepthCM:      req.DepthCM,
		Confidence:   req.Confidence,
		AuditStatus:  auditStatus,
		TraceID:      traceIDFromContext(ctx),
	}
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(report).Error
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s *AIService) ListConversationLogs(ctx context.Context, search carpoolReq.AIConversationLogSearch) ([]carpoolModel.AiConversationLog, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.AiConversationLog{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("session_id LIKE ? OR question LIKE ? OR answer LIKE ?", keyword, keyword, keyword)
	}
	if search.SessionID != "" {
		db = db.Where("session_id = ?", strings.TrimSpace(search.SessionID))
	}
	if search.UserRole != "" {
		db = db.Where("user_role = ?", normalizeAIRole(search.UserRole))
	}
	if search.Success != nil {
		db = db.Where("success = ?", *search.Success)
	}
	if search.Fallback != nil {
		db = db.Where("fallback = ?", *search.Fallback)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.AiConversationLog
	if err := db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *AIService) ListRoutePlanLogs(ctx context.Context, search carpoolReq.AIRoutePlanLogSearch) ([]carpoolModel.AiRoutePlanLog, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.AiRoutePlanLog{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("route_plan_no LIKE ? OR origin LIKE ? OR destination LIKE ?", keyword, keyword, keyword)
	}
	if search.RoutePlanNo != "" {
		db = db.Where("route_plan_no = ?", strings.TrimSpace(search.RoutePlanNo))
	}
	if search.SessionID != "" {
		db = db.Where("session_id = ?", strings.TrimSpace(search.SessionID))
	}
	if search.UserRole != "" {
		db = db.Where("user_role = ?", normalizeAIRole(search.UserRole))
	}
	if search.City != "" {
		db = db.Where("city = ?", strings.TrimSpace(search.City))
	}
	if search.Success != nil {
		db = db.Where("success = ?", *search.Success)
	}
	if search.Fallback != nil {
		db = db.Where("fallback = ?", *search.Fallback)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.AiRoutePlanLog
	if err := db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *AIService) ListFloodReports(ctx context.Context, search carpoolReq.AIFloodReportSearch) ([]carpoolModel.AiFloodReport, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.AiFloodReport{})
	if search.Keyword != "" {
		keyword := "%" + strings.TrimSpace(search.Keyword) + "%"
		db = db.Where("report_no LIKE ? OR location_text LIKE ?", keyword, keyword)
	}
	if search.ReportNo != "" {
		db = db.Where("report_no = ?", strings.TrimSpace(search.ReportNo))
	}
	if search.City != "" {
		db = db.Where("city = ?", strings.TrimSpace(search.City))
	}
	if search.AuditStatus != "" {
		db = db.Where("audit_status = ?", strings.TrimSpace(search.AuditStatus))
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	if limit <= 0 {
		limit = 20
	}
	var list []carpoolModel.AiFloodReport
	if err := db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *AIService) AuditFloodReport(ctx context.Context, req carpoolReq.AIFloodReportAuditRequest) error {
	reportNo := strings.TrimSpace(req.ReportNo)
	if reportNo == "" {
		return errors.New("reportNo is required")
	}
	status := strings.TrimSpace(req.AuditStatus)
	if status == "" {
		return errors.New("auditStatus is required")
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&carpoolModel.AiFloodReport{}).
			Where("report_no = ?", reportNo).
			Updates(map[string]any{
				"audit_status": status,
				"audit_remark": strings.TrimSpace(req.AuditRemark),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (s *AIService) GetSummary(ctx context.Context) (*AISummary, error) {
	summary := &AISummary{}
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&carpoolModel.AiConversationLog{}).Count(&summary.TotalCalls).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&carpoolModel.AiConversationLog{}).Where("success = ?", true).Count(&summary.SuccessCalls).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&carpoolModel.AiConversationLog{}).Where("fallback = ?", true).Count(&summary.FallbackCalls).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&carpoolModel.AiRoutePlanLog{}).Count(&summary.RoutePlans).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&carpoolModel.AiFloodReport{}).Count(&summary.FloodReports).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&carpoolModel.AiFloodReport{}).Where("audit_status = ?", aiAuditPendingManualReview).Count(&summary.PendingFloods).Error; err != nil {
		return nil, err
	}
	var avg float64
	if err := db.Model(&carpoolModel.AiConversationLog{}).Select("COALESCE(AVG(latency_ms),0)").Scan(&avg).Error; err != nil {
		return nil, err
	}
	summary.AvgLatencyMS = roundAnalytics(avg)
	var latest carpoolModel.AiConversationLog
	if err := db.Order("created_at DESC, id DESC").First(&latest).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	summary.LatestTraceID = latest.TraceID
	return summary, nil
}

func (s *AIService) ExportTaskID(ctx context.Context) string {
	return fmt.Sprintf("WO10-AI-EXPORT-%d", time.Now().UnixNano())
}

func (s *AIService) fallbackText(ctx context.Context, scope, role string) string {
	var template carpoolModel.AiFallbackTemplate
	err := global.GVA_DB.WithContext(ctx).
		Where("scope = ? AND user_role IN ? AND enabled = ?", scope, []string{normalizeAIRole(role), "all"}, true).
		Order(gorm.Expr("CASE WHEN user_role = ? THEN 0 ELSE 1 END, id ASC", normalizeAIRole(role))).
		First(&template).Error
	if err == nil && strings.TrimSpace(template.Content) != "" {
		return template.Content
	}
	if scope == "route" {
		return "AI route planning is temporarily unavailable. Please use standard navigation and avoid flooded or closed roads."
	}
	return "AI assistant is temporarily unavailable. Please retry later or switch to manual service support."
}

type cozeAIProvider struct {
	client *cozex.Client
}

func (p cozeAIProvider) Chat(ctx context.Context, req carpoolReq.AIChatRequest) (string, error) {
	res, err := p.client.CallTravelBot(ctx, cozex.TravelBotRequest{
		Text:      strings.TrimSpace(req.Text),
		SessionID: strings.TrimSpace(req.SessionID),
	})
	if err != nil {
		return "", err
	}
	return res.RawBody, nil
}

func (p cozeAIProvider) PlanRainRoute(ctx context.Context, req carpoolReq.RainRouteRequest) (string, error) {
	res, err := p.client.CallRainRouteWorkflow(ctx, cozex.RainRouteWorkflowRequest{
		Origin:      strings.TrimSpace(req.Origin),
		Destination: strings.TrimSpace(req.Destination),
		City:        strings.TrimSpace(req.City),
		Weather:     strings.TrimSpace(req.Weather),
		Avoid:       strings.TrimSpace(req.Avoid),
		Preference:  strings.TrimSpace(req.Preference),
		UserRole:    normalizeAIRole(req.UserRole),
	})
	if err != nil {
		return "", err
	}
	return res.RawBody, nil
}

func validateAIChat(req carpoolReq.AIChatRequest) error {
	if strings.TrimSpace(req.SessionID) == "" {
		return errors.New("sessionId is required")
	}
	if strings.TrimSpace(req.Text) == "" {
		return errors.New("text is required")
	}
	return validateAIRole(req.UserRole)
}

func validateRainRoute(req carpoolReq.RainRouteRequest) error {
	if strings.TrimSpace(req.Origin) == "" || strings.TrimSpace(req.Destination) == "" || strings.TrimSpace(req.City) == "" {
		return errors.New("origin, destination and city are required")
	}
	return validateAIRole(req.UserRole)
}

func validateFloodReport(req carpoolReq.FloodReportRequest) error {
	if strings.TrimSpace(req.City) == "" || strings.TrimSpace(req.LocationText) == "" {
		return errors.New("city and locationText are required")
	}
	if req.DepthCM < 0 {
		return errors.New("depthCm must be non-negative")
	}
	if req.Confidence < 0 || req.Confidence > 100 {
		return errors.New("confidence must be between 0 and 100")
	}
	return validateAIRole(req.ReporterRole)
}

func validateAIRole(role string) error {
	switch normalizeAIRole(role) {
	case "passenger", "driver", "admin":
		return nil
	default:
		return errors.New("userRole must be passenger, driver or admin")
	}
}

func normalizeAIRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func traceIDFromContext(ctx context.Context) string {
	return logger.FromCtx(ctx).GetTraceID()
}

func safeError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 512 {
		return message[:512]
	}
	return message
}

func nextAIRoutePlanNo() string {
	return fmt.Sprintf("AIR-WO10-%d", time.Now().UnixNano())
}

func nextAIFloodReportNo() string {
	return fmt.Sprintf("FLOOD-WO10-%d", time.Now().UnixNano())
}
