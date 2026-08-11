package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCORSFilterHandlesPreflightFromPassengerAndDriverH5(t *testing.T) {
	filter := NewCORSFilter()
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight request should not reach downstream handler")
	}))

	for _, origin := range []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://127.0.0.1:5176",
	} {
		req := httptest.NewRequest(http.MethodOptions, "/carpool/auth/sms/send", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "content-type,idempotency-key,x-trace-id")
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		require.Equal(t, http.StatusNoContent, res.Code)
		require.Equal(t, origin, res.Header().Get("Access-Control-Allow-Origin"))
		requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
		requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Methods"), http.MethodOptions)
		requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key")
		requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Headers"), "X-Trace-Id")
		require.Equal(t, "Origin", res.Header().Get("Vary"))
	}
}

func TestCORSFilterRejectsNonLocalH5Origins(t *testing.T) {
	filter := NewCORSFilter()
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight request should not reach downstream handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/carpool/auth/sms/send", nil)
	req.Header.Set("Origin", "http://evil.test:5175")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	require.Empty(t, res.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSFilterAddsHeadersToNormalResponses(t *testing.T) {
	filter := NewCORSFilter()
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/carpool/auth/login", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5174")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "http://127.0.0.1:5174", res.Header().Get("Access-Control-Allow-Origin"))
	requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
	requireHeaderContains(t, res.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	require.Equal(t, "Origin", res.Header().Get("Vary"))
}

func requireHeaderContains(t *testing.T, headerValue, expected string) {
	t.Helper()
	values := strings.Split(headerValue, ",")
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), expected) {
			return
		}
	}
	require.Failf(t, "missing header value", "expected %q in %q", expected, headerValue)
}
