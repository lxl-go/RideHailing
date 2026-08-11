package carpool

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

type fakeAIProvider struct {
	chatResp  string
	chatErr   error
	routeResp string
	routeErr  error
}

func (f fakeAIProvider) Chat(ctx context.Context, req carpoolReq.AIChatRequest) (string, error) {
	if f.chatErr != nil {
		return "", f.chatErr
	}
	return f.chatResp, nil
}

func (f fakeAIProvider) PlanRainRoute(ctx context.Context, req carpoolReq.RainRouteRequest) (string, error) {
	if f.routeErr != nil {
		return "", f.routeErr
	}
	return f.routeResp, nil
}

func newAIServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.AiConversationLog{},
		&carpoolModel.AiRoutePlanLog{},
		&carpoolModel.AiFloodReport{},
		&carpoolModel.AiFallbackTemplate{},
	))
	global.GVA_DB = db
	return db
}

func TestAIServiceChatPersistsSuccessLog(t *testing.T) {
	newAIServiceTestDB(t)
	service := NewAIServiceWithProvider(fakeAIProvider{chatResp: `{"answer":"route is safe"}`})

	result, err := service.Chat(context.Background(), carpoolReq.AIChatRequest{
		SessionID: "session-001",
		Text:      "Plan my rainy route",
		UserID:    1001,
		UserRole:  "passenger",
	})
	require.NoError(t, err)
	require.False(t, result.Fallback)
	require.Equal(t, `{"answer":"route is safe"}`, result.Answer)

	var log carpoolModel.AiConversationLog
	require.NoError(t, global.GVA_DB.First(&log).Error)
	require.Equal(t, "session-001", log.SessionID)
	require.Equal(t, "passenger", log.UserRole)
	require.True(t, log.Success)
	require.False(t, log.Fallback)
	require.Equal(t, "Plan my rainy route", log.Question)
	require.Equal(t, `{"answer":"route is safe"}`, log.Answer)
}

func TestAIServiceChatFallsBackAndPersistsLog(t *testing.T) {
	db := newAIServiceTestDB(t)
	require.NoError(t, db.Create(&carpoolModel.AiFallbackTemplate{
		Scope:    "chat",
		UserRole: "driver",
		Content:  "AI service is unavailable. Please use manual dispatch guidance.",
		Enabled:  true,
	}).Error)
	service := NewAIServiceWithProvider(fakeAIProvider{chatErr: errors.New("coze 503")})

	result, err := service.Chat(context.Background(), carpoolReq.AIChatRequest{
		SessionID: "session-002",
		Text:      "Any flooded roads near me?",
		UserID:    2001,
		UserRole:  "driver",
	})
	require.NoError(t, err)
	require.True(t, result.Fallback)
	require.Contains(t, result.Answer, "manual dispatch")

	var log carpoolModel.AiConversationLog
	require.NoError(t, global.GVA_DB.First(&log).Error)
	require.False(t, log.Success)
	require.True(t, log.Fallback)
	require.Contains(t, log.ErrorMessage, "coze 503")
}

func TestAIServicePlanRainRoutePersistsLog(t *testing.T) {
	newAIServiceTestDB(t)
	service := NewAIServiceWithProvider(fakeAIProvider{routeResp: `{"risk_level":"MEDIUM","summary":"avoid tunnel"}`})

	result, err := service.PlanRainRoute(context.Background(), carpoolReq.RainRouteRequest{
		SessionID:   "session-003",
		Origin:      "Jingan Temple",
		Destination: "Hongqiao Station",
		City:        "Shanghai",
		Weather:     "heavy rain",
		Avoid:       "tunnel",
		Preference:  "safe",
		UserRole:    "passenger",
	})
	require.NoError(t, err)
	require.False(t, result.Fallback)
	require.NotEmpty(t, result.RoutePlanNo)
	require.Equal(t, `{"risk_level":"MEDIUM","summary":"avoid tunnel"}`, result.RawResult)

	var log carpoolModel.AiRoutePlanLog
	require.NoError(t, global.GVA_DB.First(&log).Error)
	require.Equal(t, result.RoutePlanNo, log.RoutePlanNo)
	require.Equal(t, "Shanghai", log.City)
	require.Equal(t, "Jingan Temple", log.Origin)
	require.Equal(t, "Hongqiao Station", log.Destination)
	require.True(t, log.Success)
}

func TestAIServiceReportFloodingLowConfidencePendingManualReview(t *testing.T) {
	newAIServiceTestDB(t)
	service := NewAIServiceWithProvider(fakeAIProvider{})

	report, err := service.ReportFlooding(context.Background(), carpoolReq.FloodReportRequest{
		ReporterID:   1001,
		ReporterRole: "passenger",
		City:         "Shanghai",
		LocationText: "Yan'an elevated entrance",
		DepthCM:      35,
		Confidence:   72,
	})
	require.NoError(t, err)
	require.NotEmpty(t, report.ReportNo)
	require.Equal(t, "pending_manual_review", report.AuditStatus)

	var saved carpoolModel.AiFloodReport
	require.NoError(t, global.GVA_DB.First(&saved).Error)
	require.Equal(t, report.ReportNo, saved.ReportNo)
	require.Equal(t, "pending_manual_review", saved.AuditStatus)
	require.Equal(t, 72.0, saved.Confidence)
}
