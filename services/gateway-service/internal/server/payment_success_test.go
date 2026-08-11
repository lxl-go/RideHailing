package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/require"

	"ride-hailing/pkg/alipayx"
	"ride-hailing/services/gateway-service/internal/conf"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

func TestAlipaySuccessRouteReturnsSuccessPage(t *testing.T) {
	srv := khttp.NewServer()
	registerPaymentRoutes(srv, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/pay/success", nil)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Header().Get("Content-Type"), "text/html")
	require.Contains(t, res.Body.String(), `data-action="return-orders"`)
	require.Contains(t, res.Body.String(), `location.href='/#/pages/orders/orders'`)
	require.Contains(t, res.Body.String(), `data-action="return-home"`)
	require.Contains(t, res.Body.String(), `location.href='/#/pages/home/home'`)
	require.Contains(t, res.Body.String(), "支付成功")
	require.Contains(t, res.Body.String(), "返回订单")
}

func TestAlipaySuccessRouteMarksPaymentPaidAndLinksOrderDetail(t *testing.T) {
	srv := khttp.NewServer()
	privateKey, publicKey := testAlipayKeyPair(t)
	orderSvc := &fakePaymentOrderService{}
	registerPaymentRoutes(srv, &conf.Alipay{
		AppID:           "9021000161685035",
		PrivateKey:      privateKey,
		AlipayPublicKey: publicKey,
	}, orderSvc)
	query := signedAlipayNotify(t, privateKey, map[string]string{
		"app_id":       "9021000161685035",
		"out_trade_no": "PAY7001",
		"trade_no":     "2026080422001",
		"total_amount": "39.80",
		"seller_id":    "2088721000161685035",
		"charset":      "utf-8",
		"sign_type":    "RSA2",
	})

	req := httptest.NewRequest(http.MethodGet, "/pay/success?"+query, nil)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, 1, orderSvc.markCalls)
	require.Equal(t, "PAY7001", orderSvc.markReq.OutTradeNo)
	require.Equal(t, "2026080422001", orderSvc.markReq.AlipayTradeNo)
	require.Equal(t, "39.80", orderSvc.markReq.TotalAmount)
	require.Equal(t, "TRADE_SUCCESS", orderSvc.markReq.TradeStatus)
	require.Contains(t, strings.ReplaceAll(res.Body.String(), "&amp;", "&"), "/#/pages/orderDetail/orderDetail?id=5001")
	require.Contains(t, res.Body.String(), `/carpool/orders/5001/payment/sync`)
}

type fakePaymentAlipayQueryClient struct {
	outTradeNo string
	reply      *alipayx.TradeQueryReply
}

func (f *fakePaymentAlipayQueryClient) VerifyNotify(map[string][]string) error {
	return nil
}

func (f *fakePaymentAlipayQueryClient) TradeQuery(_ context.Context, outTradeNo string) (*alipayx.TradeQueryReply, error) {
	f.outTradeNo = outTradeNo
	if f.reply != nil {
		return f.reply, nil
	}
	return &alipayx.TradeQueryReply{
		OutTradeNo:  outTradeNo,
		TradeNo:     "2026081122001",
		TradeStatus: "TRADE_SUCCESS",
		TotalAmount: "39.80",
	}, nil
}

func TestPaymentSyncRouteQueriesAlipayAndMarksPaymentPaid(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakePaymentOrderService{statusReply: &orderv1.GetPaymentStatusReply{
		OrderId:     5001,
		OutTradeNo:  "PAY7001",
		TotalAmount: "39.80",
		Status:      0,
	}}
	alipayClient := &fakePaymentAlipayQueryClient{}
	registerPaymentRoutesWithDeps(srv, nil, orderSvc, func(*conf.Alipay) (paymentAlipayClient, error) {
		return alipayClient, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/carpool/orders/5001/payment/sync", strings.NewReader(`{"payment_no":"PAY7001"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	require.Equal(t, float64(0), body["code"])
	data := body["data"].(map[string]any)
	require.Equal(t, true, data["synced"])
	require.Equal(t, "5001", data["orderId"])
	require.Equal(t, "PAY7001", alipayClient.outTradeNo)
	require.Equal(t, 1, orderSvc.markCalls)
	require.Equal(t, "PAY7001", orderSvc.markReq.OutTradeNo)
	require.Equal(t, "2026081122001", orderSvc.markReq.AlipayTradeNo)
	require.Equal(t, "TRADE_SUCCESS", orderSvc.markReq.TradeStatus)
	require.Equal(t, "39.80", orderSvc.markReq.TotalAmount)
}
