package alipayx

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateWapPayBuildsSandboxFormWithSignature(t *testing.T) {
	privateKey := testPKCS1PrivateKey(t)
	client, err := NewClient(Config{
		AppID:      "9021000161685035",
		PrivateKey: privateKey,
		Production: false,
		NotifyURL:  "https://example.test/api/v1/pay/notify",
		ReturnURL:  "https://example.test/pay/success",
	})
	require.NoError(t, err)

	form, err := client.CreateWapPay(context.Background(), WapPayRequest{
		OutTradeNo:  "PAY202608040001",
		Subject:     "RideHailing Order 5001",
		TotalAmount: "39.80",
	})

	require.NoError(t, err)
	require.Contains(t, form, "https://openapi-sandbox.dl.alipaydev.com/gateway.do")
	require.Contains(t, form, `name="app_id" value="9021000161685035"`)
	require.Contains(t, form, `&#34;out_trade_no&#34;:&#34;PAY202608040001&#34;`)
	require.Contains(t, form, `name="sign" value="`)
	require.NotContains(t, form, privateKey)
}

func TestVerifyNotifyRejectsBadSignature(t *testing.T) {
	privateKey := testPKCS1PrivateKey(t)
	client, err := NewClient(Config{
		AppID:           "9021000161685035",
		PrivateKey:      privateKey,
		AlipayPublicKey: testPublicKeyFromPrivate(t, privateKey),
		Production:      false,
	})
	require.NoError(t, err)

	err = client.VerifyNotify(map[string][]string{
		"app_id":       {"9021000161685035"},
		"out_trade_no": {"PAY202608040001"},
		"trade_status": {"TRADE_SUCCESS"},
		"sign_type":    {"RSA2"},
		"sign":         {base64.StdEncoding.EncodeToString([]byte("bad-signature"))},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "verify alipay notify")
}

func TestTradeQueryPostsSignedSandboxRequest(t *testing.T) {
	privateKey := testPKCS1PrivateKey(t)
	var requestBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())
		requestBody = r.PostForm.Encode()
		require.Equal(t, "9021000161685035", r.PostForm.Get("app_id"))
		require.Equal(t, "alipay.trade.query", r.PostForm.Get("method"))
		require.Contains(t, r.PostForm.Get("biz_content"), `"out_trade_no":"PAY202608040001"`)
		require.NotEmpty(t, r.PostForm.Get("sign"))
		_, _ = w.Write([]byte(`{"alipay_trade_query_response":{"code":"10000","msg":"Success","out_trade_no":"PAY202608040001","trade_no":"2026081122001","trade_status":"TRADE_SUCCESS","total_amount":"39.80"}}`))
	}))
	defer server.Close()
	client, err := NewClient(Config{
		AppID:      "9021000161685035",
		PrivateKey: privateKey,
		GatewayURL: server.URL,
	})
	require.NoError(t, err)

	reply, err := client.TradeQuery(context.Background(), "PAY202608040001")

	require.NoError(t, err)
	require.Contains(t, requestBody, "alipay.trade.query")
	require.Equal(t, "PAY202608040001", reply.OutTradeNo)
	require.Equal(t, "2026081122001", reply.TradeNo)
	require.Equal(t, "TRADE_SUCCESS", reply.TradeStatus)
	require.Equal(t, "39.80", reply.TotalAmount)
}

func testPKCS1PrivateKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
}

func testPublicKeyFromPrivate(t *testing.T, privateKey string) string {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(privateKey))
	require.NoError(t, err)
	key, err := x509.ParsePKCS1PrivateKey(raw)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
}
