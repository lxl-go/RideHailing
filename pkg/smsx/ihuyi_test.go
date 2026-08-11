package smsx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIhuyiClientSendVerificationCodePostsForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		require.NoError(t, r.ParseForm())
		require.Equal(t, "C76777563", r.Form.Get("account"))
		require.Equal(t, "api-key", r.Form.Get("password"))
		require.Equal(t, "13800138000", r.Form.Get("mobile"))
		require.Equal(t, "您的验证码是：123456。请不要把验证码泄露给其他人。", r.Form.Get("content"))
		_ = json.NewEncoder(w).Encode(Response{Code: 2, Msg: "提交成功", SMSID: "sms-1"})
	}))
	defer server.Close()

	client := NewIhuyiClient(IhuyiConfig{
		Account:  "C76777563",
		Password: "api-key",
		Endpoint: server.URL,
	})

	err := client.SendVerificationCode(t.Context(), "13800138000", "123456")

	require.NoError(t, err)
}

func TestIhuyiClientReturnsErrorWhenProviderRejects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response{Code: 405, Msg: "手机号码格式不正确"})
	}))
	defer server.Close()

	client := NewIhuyiClient(IhuyiConfig{
		Account:  "C76777563",
		Password: "api-key",
		Endpoint: server.URL,
	})

	err := client.SendVerificationCode(t.Context(), "bad-mobile", "123456")

	require.Error(t, err)
	require.Contains(t, err.Error(), "手机号码格式不正确")
}
