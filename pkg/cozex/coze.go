package cozex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultTravelBotURL       = "https://fff2xdtnzj.coze.site/stream_run"
	DefaultRainRouteURL       = "https://xchnkhx636.coze.site/run"
	DefaultTravelBotProjectID = int64(7668272524714786851)
	DefaultTravelBotSessionID = "mJWDi7xEUp4wEQNj3RK1F"
	DefaultHTTPTimeout        = 5 * time.Second
)

type Config struct {
	TravelBotURL           string
	RainRouteWorkflowURL   string
	TravelBotToken         string
	RainRouteWorkflowToken string
	TravelBotProjectID     int64
	TravelBotSessionID     string
	Timeout                time.Duration
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type TravelBotRequest struct {
	Text      string
	SessionID string
}

type TravelBotResponse struct {
	RawBody string `json:"rawBody"`
}

type RainRouteWorkflowRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	City        string `json:"city"`
	Weather     string `json:"weather"`
	Avoid       string `json:"avoid"`
	Preference  string `json:"preference"`
	UserRole    string `json:"user_role"`
}

type RainRouteWorkflowResponse struct {
	RawBody string `json:"rawBody"`
}

type travelBotPayload struct {
	Content   travelBotContent `json:"content"`
	Type      string           `json:"type"`
	SessionID string           `json:"session_id"`
	ProjectID int64            `json:"project_id"`
}

type travelBotContent struct {
	Query travelBotQuery `json:"query"`
}

type travelBotQuery struct {
	Prompt []travelBotPrompt `json:"prompt"`
}

type travelBotPrompt struct {
	Type    string              `json:"type"`
	Content travelBotPromptText `json:"content"`
}

type travelBotPromptText struct {
	Text string `json:"text"`
}

func NewClient(cfg Config, httpClient *http.Client) *Client {
	if cfg.TravelBotURL == "" {
		cfg.TravelBotURL = DefaultTravelBotURL
	}
	if cfg.RainRouteWorkflowURL == "" {
		cfg.RainRouteWorkflowURL = DefaultRainRouteURL
	}
	if cfg.TravelBotProjectID == 0 {
		cfg.TravelBotProjectID = DefaultTravelBotProjectID
	}
	if cfg.TravelBotSessionID == "" {
		cfg.TravelBotSessionID = DefaultTravelBotSessionID
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultHTTPTimeout
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{cfg: cfg, httpClient: httpClient}
}

func (c *Client) CallTravelBot(ctx context.Context, req TravelBotRequest) (*TravelBotResponse, error) {
	if strings.TrimSpace(c.cfg.TravelBotToken) == "" {
		return nil, errors.New("coze travel bot token is required")
	}
	payload := travelBotPayload{
		Content: travelBotContent{Query: travelBotQuery{Prompt: []travelBotPrompt{
			{Type: "text", Content: travelBotPromptText{Text: req.Text}},
		}}},
		Type:      "query",
		SessionID: c.travelBotSessionID(req.SessionID),
		ProjectID: c.cfg.TravelBotProjectID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	raw, err := c.doPost(ctx, c.cfg.TravelBotURL, c.cfg.TravelBotToken, body)
	if err != nil {
		return nil, err
	}
	return &TravelBotResponse{RawBody: raw}, nil
}

func (c *Client) CallRainRouteWorkflow(ctx context.Context, req RainRouteWorkflowRequest) (*RainRouteWorkflowResponse, error) {
	if strings.TrimSpace(c.cfg.RainRouteWorkflowToken) == "" {
		return nil, errors.New("coze rain route workflow token is required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	raw, err := c.doPost(ctx, c.cfg.RainRouteWorkflowURL, c.cfg.RainRouteWorkflowToken, body)
	if err != nil {
		return nil, err
	}
	return &RainRouteWorkflowResponse{RawBody: raw}, nil
}

func (c *Client) doPost(ctx context.Context, url string, token string, body []byte) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if readErr != nil {
			return "", fmt.Errorf("coze private endpoint returned status=%d read-body-err=%w", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("coze private endpoint returned status=%d body=%s", resp.StatusCode, previewBody(string(raw)))
	}
	if readErr != nil {
		return "", readErr
	}
	return string(raw), nil
}

func (c *Client) travelBotSessionID(requestSessionID string) string {
	if sessionID := strings.TrimSpace(c.cfg.TravelBotSessionID); sessionID != "" {
		return sessionID
	}
	return strings.TrimSpace(requestSessionID)
}

func previewBody(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 256 {
		return body[:256]
	}
	return body
}
