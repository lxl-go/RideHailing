package server

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/require"

	"ride-hailing/services/gateway-service/internal/conf"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type fakePaymentOrderService struct {
	statusReply *orderv1.GetPaymentStatusReply
	markReq     *orderv1.MarkPaymentPaidRequest
	markCalls   int
}

func (f *fakePaymentOrderService) GetPaymentStatus(_ context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error) {
	if f.statusReply != nil {
		return f.statusReply, nil
	}
	return &orderv1.GetPaymentStatusReply{
		OrderId:     5001,
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: "39.80",
		Status:      int32(0),
	}, nil
}

func (f *fakePaymentOrderService) MarkPaymentPaid(_ context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error) {
	copy := *req
	f.markReq = &copy
	f.markCalls++
	return &orderv1.MarkPaymentPaidReply{OrderId: 5001, OutTradeNo: req.OutTradeNo, Status: int32(1)}, nil
}

func TestAlipayNotifyRouteReturnsFailForInvalidSignature(t *testing.T) {
	srv := khttp.NewServer()
	privateKey, publicKey := testAlipayKeyPair(t)
	registerPaymentRoutes(srv, &conf.Alipay{
		AppID:           "9021000161685035",
		PrivateKey:      privateKey,
		AlipayPublicKey: publicKey,
	}, &fakePaymentOrderService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pay/notify", strings.NewReader("app_id=9021000161685035&trade_status=TRADE_SUCCESS&sign=bad"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "text/plain; charset=utf-8", res.Header().Get("Content-Type"))
	require.Equal(t, "fail", strings.TrimSpace(res.Body.String()))
}

func TestAlipayNotifyMarksPaymentPaidAfterValidNotify(t *testing.T) {
	srv := khttp.NewServer()
	privateKey, publicKey := testAlipayKeyPair(t)
	orderSvc := &fakePaymentOrderService{}
	registerPaymentRoutes(srv, &conf.Alipay{
		AppID:           "9021000161685035",
		PrivateKey:      privateKey,
		AlipayPublicKey: publicKey,
	}, orderSvc)
	form := signedAlipayNotify(t, privateKey, map[string]string{
		"app_id":       "9021000161685035",
		"out_trade_no": "PAY7001",
		"trade_no":     "2026080422001",
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "39.80",
		"seller_id":    "2088721000161685035",
		"charset":      "utf-8",
		"sign_type":    "RSA2",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pay/notify", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "success", strings.TrimSpace(res.Body.String()))
	require.Equal(t, 1, orderSvc.markCalls)
	require.Equal(t, "PAY7001", orderSvc.markReq.OutTradeNo)
	require.Equal(t, "2026080422001", orderSvc.markReq.AlipayTradeNo)
	require.Equal(t, "39.80", orderSvc.markReq.TotalAmount)
	require.Equal(t, "TRADE_SUCCESS", orderSvc.markReq.TradeStatus)
}

func TestAlipayNotifyReturnsFailWhenAmountMismatch(t *testing.T) {
	srv := khttp.NewServer()
	privateKey, publicKey := testAlipayKeyPair(t)
	orderSvc := &fakePaymentOrderService{statusReply: &orderv1.GetPaymentStatusReply{
		OrderId:     5001,
		OutTradeNo:  "PAY7001",
		TotalAmount: "39.80",
		Status:      int32(0),
	}}
	registerPaymentRoutes(srv, &conf.Alipay{
		AppID:           "9021000161685035",
		PrivateKey:      privateKey,
		AlipayPublicKey: publicKey,
	}, orderSvc)
	form := signedAlipayNotify(t, privateKey, map[string]string{
		"app_id":       "9021000161685035",
		"out_trade_no": "PAY7001",
		"trade_no":     "2026080422001",
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "40.00",
		"charset":      "utf-8",
		"sign_type":    "RSA2",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pay/notify", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "fail", strings.TrimSpace(res.Body.String()))
	require.Equal(t, 0, orderSvc.markCalls)
}

func testAlipayKeyPair(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privateKey := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	publicKey := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
	return privateKey, publicKey
}

func signedAlipayNotify(t *testing.T, privateKeyBase64 string, params map[string]string) string {
	t.Helper()
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	require.NoError(t, err)
	key, err := x509.ParsePKCS1PrivateKey(keyBytes)
	require.NoError(t, err)
	canonical := canonicalAlipayParams(params)
	digest := sha256.Sum256([]byte(canonical))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	require.NoError(t, err)
	params["sign"] = base64.StdEncoding.EncodeToString(signature)
	values := make([]string, 0, len(params))
	for key, value := range params {
		values = append(values, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}
	sort.Strings(values)
	return strings.Join(values, "&")
}

func canonicalAlipayParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || key == "sign_type" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	return strings.Join(parts, "&")
}
