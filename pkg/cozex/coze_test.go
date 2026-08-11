package cozex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func TestCallTravelBotBuildsExactStreamRunShape(t *testing.T) {
	var gotMethod, gotAuth, gotContentType string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"answer":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		TravelBotURL:           server.URL + "/stream_run",
		RainRouteWorkflowURL:   "https://unused.invalid/run",
		TravelBotToken:         "travel-token",
		RainRouteWorkflowToken: "workflow-token",
	}, server.Client())

	res, err := client.CallTravelBot(context.Background(), TravelBotRequest{
		Text:      "上海暴雨，帮我规划路线",
		SessionID: "session-001",
	})
	if err != nil {
		t.Fatalf("CallTravelBot error: %v", err)
	}
	if res.RawBody != `{"answer":"ok"}` {
		t.Fatalf("unexpected raw body: %s", res.RawBody)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s, want POST", gotMethod)
	}
	if gotAuth != "Bearer travel-token" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content type = %q", gotContentType)
	}
	if gotBody["type"] != "query" {
		t.Fatalf("outer type = %#v", gotBody["type"])
	}
	if gotBody["session_id"] != DefaultTravelBotSessionID {
		t.Fatalf("session_id = %#v", gotBody["session_id"])
	}
	if gotBody["project_id"] != float64(DefaultTravelBotProjectID) {
		t.Fatalf("project_id = %#v", gotBody["project_id"])
	}

	content := gotBody["content"].(map[string]any)
	query := content["query"].(map[string]any)
	prompt := query["prompt"].([]any)
	item := prompt[0].(map[string]any)
	if item["type"] != "text" {
		t.Fatalf("prompt type = %#v", item["type"])
	}
	text := item["content"].(map[string]any)["text"]
	if text != "上海暴雨，帮我规划路线" {
		t.Fatalf("text = %#v", text)
	}
}

func TestCallTravelBotUsesConfiguredDeploymentSessionID(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"answer":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		TravelBotURL:       server.URL + "/stream_run",
		TravelBotToken:     "travel-token",
		TravelBotSessionID: "mJWDi7xEUp4wEQNj3RK1F",
	}, server.Client())

	_, err := client.CallTravelBot(context.Background(), TravelBotRequest{
		Text:      "测试固定 session",
		SessionID: "frontend-session-should-not-leak",
	})
	if err != nil {
		t.Fatalf("CallTravelBot error: %v", err)
	}
	if gotBody["session_id"] != "mJWDi7xEUp4wEQNj3RK1F" {
		t.Fatalf("session_id = %#v", gotBody["session_id"])
	}
}

func TestCallRainRouteWorkflowBuildsExactSevenFieldShape(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content type = %q", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"risk_level":"MEDIUM"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		TravelBotURL:           "https://unused.invalid/stream_run",
		RainRouteWorkflowURL:   server.URL + "/run",
		TravelBotToken:         "travel-token",
		RainRouteWorkflowToken: "workflow-token",
	}, server.Client())

	res, err := client.CallRainRouteWorkflow(context.Background(), RainRouteWorkflowRequest{
		Origin:      "上海静安寺",
		Destination: "虹桥火车站",
		City:        "上海",
		Weather:     "暴雨黄色预警",
		Avoid:       "积水路段、隧道",
		Preference:  "安全",
		UserRole:    "passenger",
	})
	if err != nil {
		t.Fatalf("CallRainRouteWorkflow error: %v", err)
	}
	if res.RawBody != `{"risk_level":"MEDIUM"}` {
		t.Fatalf("unexpected raw body: %s", res.RawBody)
	}
	if gotAuth != "Bearer workflow-token" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if _, ok := gotBody["project_id"]; ok {
		t.Fatalf("workflow body must not contain project_id: %#v", gotBody)
	}

	keys := make([]string, 0, len(gotBody))
	for k := range gotBody {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	want := []string{"avoid", "city", "destination", "origin", "preference", "user_role", "weather"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %#v, want %#v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %#v, want %#v", keys, want)
		}
	}
	if gotBody["user_role"] != "passenger" {
		t.Fatalf("user_role = %#v", gotBody["user_role"])
	}
}

func TestCozeClientRejectsMissingDedicatedTokens(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{
		TravelBotURL:           server.URL + "/stream_run",
		RainRouteWorkflowURL:   server.URL + "/run",
		RainRouteWorkflowToken: "workflow-token",
	}, server.Client())
	if _, err := client.CallTravelBot(context.Background(), TravelBotRequest{Text: "hi", SessionID: "s"}); err == nil {
		t.Fatal("expected missing travel token error")
	}

	client = NewClient(Config{
		TravelBotURL:         server.URL + "/stream_run",
		RainRouteWorkflowURL: server.URL + "/run",
		TravelBotToken:       "travel-token",
	}, server.Client())
	if _, err := client.CallRainRouteWorkflow(context.Background(), RainRouteWorkflowRequest{Origin: "a", Destination: "b", City: "c"}); err == nil {
		t.Fatal("expected missing workflow token error")
	}

	if calls != 0 {
		t.Fatalf("expected no HTTP calls, got %d", calls)
	}
}
