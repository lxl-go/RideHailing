package data

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	authv1 "ride-hailing/services/auth-service/api/auth/v1"
)

func TestAuthHTTPClientLoginUsesAuthPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/auth/login", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userId":"1001","role":"passenger","accessToken":"token-1","tokenType":"Bearer","expiresIn":"7200","refreshToken":"refresh-1"}`))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	reply, err := client.Login(t.Context(), &authv1.LoginRequest{Principal: "13800138000", Role: "passenger"})

	require.NoError(t, err)
	require.Equal(t, int64(1001), reply.UserId)
	require.Equal(t, "token-1", reply.AccessToken)
	require.Equal(t, "refresh-1", reply.RefreshToken)
}

func TestAuthHTTPClientVerifyTokenUsesVerifyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/auth/verify", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userId":"2001","role":"driver","jti":"jti-1"}`))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	reply, err := client.VerifyToken(t.Context(), "Bearer token-1")

	require.NoError(t, err)
	require.Equal(t, int64(2001), reply.UserId)
	require.Equal(t, "driver", reply.Role)
	require.Equal(t, "jti-1", reply.Jti)
}

func TestAuthHTTPClientSendLoginCodeUsesSMSPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/auth/sms/send", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sent":true}`))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	reply, err := client.SendLoginCode(t.Context(), &authv1.SendLoginCodeRequest{Mobile: "13800138000", Role: "passenger"})

	require.NoError(t, err)
	require.True(t, reply.Sent)
}

func TestAuthHTTPClientPreservesUpstreamHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"message":"invalid principal"}`))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	_, err := client.SendLoginCode(t.Context(), &authv1.SendLoginCodeRequest{})

	require.Error(t, err)
	upstreamErr, ok := err.(*UpstreamHTTPError)
	require.True(t, ok)
	require.Equal(t, http.StatusBadRequest, upstreamErr.StatusCode)
	require.Equal(t, "invalid principal", upstreamErr.Message)
}

func TestAuthHTTPClientRefreshTokenUsesRefreshPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/auth/refresh", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userId":"1001","role":"passenger","accessToken":"token-2","refreshToken":"refresh-2","tokenType":"Bearer","expiresIn":"7200"}`))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	reply, err := client.RefreshToken(t.Context(), "refresh-1")

	require.NoError(t, err)
	require.Equal(t, "token-2", reply.AccessToken)
	require.Equal(t, "refresh-2", reply.RefreshToken)
}

func TestAuthHTTPClientLogoutUsesLogoutPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/auth/logout", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"logged_out":true}`))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	reply, err := client.Logout(t.Context(), "refresh-1")

	require.NoError(t, err)
	require.True(t, reply.LoggedOut)
}

func TestAuthHTTPClientCheckPermissionUsesPermissionCheckPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, http.MethodPost, r.Method)
		var req authv1.CheckPermissionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, int64(1001), req.UserId)
		require.Equal(t, "order:create", req.PermissionCode)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(&authv1.CheckPermissionReply{Allowed: true}))
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	allowed, err := client.CheckPermission(t.Context(), 1001, "order:create")

	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, "/v1/auth/permission/check", gotPath)
}
