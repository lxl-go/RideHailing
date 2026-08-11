package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"

	"ride-hailing/pkg/alipayx"
	"ride-hailing/services/gateway-service/internal/conf"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type paymentOrderService interface {
	GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error)
	MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error)
}

type paymentAlipayClient interface {
	VerifyNotify(values map[string][]string) error
	TradeQuery(ctx context.Context, outTradeNo string) (*alipayx.TradeQueryReply, error)
}

type paymentAlipayFactory func(*conf.Alipay) (paymentAlipayClient, error)

func registerPaymentRoutes(srv *khttp.Server, alipayCfg *conf.Alipay, orderSvc paymentOrderService) {
	registerPaymentRoutesWithDeps(srv, alipayCfg, orderSvc, defaultPaymentAlipayFactory)
}

func registerPaymentRoutesWithDeps(srv *khttp.Server, alipayCfg *conf.Alipay, orderSvc paymentOrderService, alipayFactory paymentAlipayFactory) {
	router := srv.Route("/")
	router.GET("/pay/success", func(ctx khttp.Context) error {
		orderID, markErr := markPaymentPaidFromReturn(ctx, alipayCfg, orderSvc, alipayFactory)
		if markErr != nil {
			zap.L().Warn("alipay return mark paid skipped", gatewayLogFields(ctx.Request(), zap.Error(markErr))...)
		}
		ctx.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx.Response().WriteHeader(http.StatusOK)
		_, err := ctx.Response().Write([]byte(alipaySuccessHTML(orderID)))
		return err
	})
	router.POST("/api/v1/pay/notify", func(ctx khttp.Context) error {
		req := ctx.Request()
		if err := req.ParseForm(); err != nil {
			zap.L().Warn("parse alipay notify failed", gatewayLogFields(req, zap.Error(err))...)
			return writeAlipayNotifyResult(ctx, "fail")
		}

		client, err := alipayFactory(alipayCfg)
		if err != nil {
			zap.L().Error("alipay notify config unavailable", zap.Error(err))
			return writeAlipayNotifyResult(ctx, "fail")
		}
		if err := client.VerifyNotify(req.PostForm); err != nil {
			zap.L().Warn("verify alipay notify failed",
				gatewayLogFields(req,
					zap.String("out_trade_no", req.PostForm.Get("out_trade_no")),
					zap.String("trade_status", req.PostForm.Get("trade_status")),
					zap.String("app_id", req.PostForm.Get("app_id")),
					zap.String("total_amount", req.PostForm.Get("total_amount")),
					zap.Error(err),
				)...)
			return writeAlipayNotifyResult(ctx, "fail")
		}

		if orderSvc == nil {
			zap.L().Error("alipay notify order service unavailable")
			return writeAlipayNotifyResult(ctx, "fail")
		}
		outTradeNo := strings.TrimSpace(req.PostForm.Get("out_trade_no"))
		alipayTradeNo := strings.TrimSpace(req.PostForm.Get("trade_no"))
		tradeStatus := strings.TrimSpace(req.PostForm.Get("trade_status"))
		totalAmount := normalizeAlipayNotifyAmount(req.PostForm.Get("total_amount"))
		if outTradeNo == "" || alipayTradeNo == "" || totalAmount == "" {
			zap.L().Warn("alipay notify missing required fields",
				gatewayLogFields(req,
					zap.String("out_trade_no", outTradeNo),
					zap.String("alipay_trade_no", alipayTradeNo),
					zap.String("trade_status", tradeStatus),
					zap.String("app_id", req.PostForm.Get("app_id")),
					zap.String("total_amount", req.PostForm.Get("total_amount")),
				)...)
			return writeAlipayNotifyResult(ctx, "fail")
		}
		if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
			zap.L().Warn("alipay notify ignored unsuccessful trade",
				gatewayLogFields(req,
					zap.String("out_trade_no", outTradeNo),
					zap.String("trade_status", tradeStatus),
					zap.String("app_id", req.PostForm.Get("app_id")),
					zap.String("total_amount", totalAmount),
				)...)
			return writeAlipayNotifyResult(ctx, "fail")
		}
		statusReply, err := orderSvc.GetPaymentStatus(ctx, &orderv1.GetPaymentStatusRequest{OutTradeNo: outTradeNo})
		if err != nil {
			zap.L().Warn("alipay notify payment status lookup failed", gatewayLogFields(req, zap.String("out_trade_no", outTradeNo), zap.String("payment_no", outTradeNo), zap.String("app_id", req.PostForm.Get("app_id")), zap.String("total_amount", totalAmount), zap.Error(err))...)
			return writeAlipayNotifyResult(ctx, "fail")
		}
		expectedAmount := normalizeAlipayNotifyAmount(statusReply.GetTotalAmount())
		if expectedAmount == "" || expectedAmount != totalAmount {
			zap.L().Warn("alipay notify amount mismatch",
				zap.String("out_trade_no", outTradeNo),
				zap.String("expected_amount", statusReply.GetTotalAmount()),
				zap.String("notify_amount", req.PostForm.Get("total_amount")),
			)
			return writeAlipayNotifyResult(ctx, "fail")
		}
		reply, err := orderSvc.MarkPaymentPaid(ctx, &orderv1.MarkPaymentPaidRequest{
			OutTradeNo:    outTradeNo,
			AlipayTradeNo: alipayTradeNo,
			AppId:         strings.TrimSpace(req.PostForm.Get("app_id")),
			TotalAmount:   totalAmount,
			TradeStatus:   tradeStatus,
			NotifyPayload: req.PostForm.Encode(),
		})
		if err != nil {
			zap.L().Warn("alipay notify mark paid failed", gatewayLogFields(req, zap.String("out_trade_no", outTradeNo), zap.String("payment_no", outTradeNo), zap.String("app_id", req.PostForm.Get("app_id")), zap.String("total_amount", totalAmount), zap.String("trade_status", tradeStatus), zap.Error(err))...)
			return writeAlipayNotifyResult(ctx, "fail")
		}
		zap.L().Info("alipay notify verified",
			gatewayLogFields(req,
				zap.Int64("order_id", reply.GetOrderId()),
				zap.String("payment_no", outTradeNo),
				zap.String("out_trade_no", outTradeNo),
				zap.String("alipay_trade_no", alipayTradeNo),
				zap.String("app_id", req.PostForm.Get("app_id")),
				zap.String("total_amount", totalAmount),
				zap.String("trade_status", tradeStatus),
				zap.String("status_before", "paying"),
				zap.String("status_after", "paid"),
				zap.Bool("duplicated", reply.GetDuplicated()),
			)...)
		return writeAlipayNotifyResult(ctx, "success")
	})
	router.POST("/carpool/orders/{id}/payment/sync", func(ctx khttp.Context) error {
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		req := new(paymentSyncRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := syncPaymentFromAlipay(ctx, alipayCfg, orderSvc, alipayFactory, orderID, req.paymentNo())
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, paymentSyncResponse(reply), nil)
	})
}

func defaultPaymentAlipayFactory(cfg *conf.Alipay) (paymentAlipayClient, error) {
	return newAlipayClient(cfg)
}

type paymentSyncRequest struct {
	PaymentNo       string `json:"payment_no"`
	PaymentNoCamel  string `json:"paymentNo"`
	OutTradeNo      string `json:"out_trade_no"`
	OutTradeNoCamel string `json:"outTradeNo"`
}

func (r *paymentSyncRequest) paymentNo() string {
	if r == nil {
		return ""
	}
	for _, value := range []string{r.PaymentNo, r.PaymentNoCamel, r.OutTradeNo, r.OutTradeNoCamel} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

type paymentSyncReply struct {
	OrderID       int64
	PaymentNo     string
	PaymentStatus int32
	OrderStatus   int32
	Synced        bool
	Duplicated    bool
}

func syncPaymentFromAlipay(ctx context.Context, alipayCfg *conf.Alipay, orderSvc paymentOrderService, alipayFactory paymentAlipayFactory, orderID int64, paymentNo string) (*paymentSyncReply, error) {
	if orderSvc == nil {
		return nil, fmt.Errorf("order service unavailable")
	}
	if alipayFactory == nil {
		alipayFactory = defaultPaymentAlipayFactory
	}
	statusReq := &orderv1.GetPaymentStatusRequest{OrderId: orderID}
	if paymentNo = strings.TrimSpace(paymentNo); paymentNo != "" {
		statusReq.OutTradeNo = paymentNo
	}
	statusReply, err := orderSvc.GetPaymentStatus(ctx, statusReq)
	if err != nil {
		return nil, err
	}
	if statusReply.GetOrderId() != orderID {
		return nil, fmt.Errorf("payment order mismatch")
	}
	outTradeNo := strings.TrimSpace(statusReply.GetOutTradeNo())
	if outTradeNo == "" {
		outTradeNo = paymentNo
	}
	if outTradeNo == "" {
		return nil, fmt.Errorf("payment no is required")
	}
	if statusReply.GetStatus() == 1 {
		return &paymentSyncReply{OrderID: orderID, PaymentNo: outTradeNo, PaymentStatus: statusReply.GetStatus(), OrderStatus: 4, Synced: true, Duplicated: true}, nil
	}
	client, err := alipayFactory(alipayCfg)
	if err != nil {
		return nil, err
	}
	queryReply, err := client.TradeQuery(ctx, outTradeNo)
	if err != nil {
		return nil, err
	}
	tradeStatus := strings.TrimSpace(queryReply.TradeStatus)
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		return nil, fmt.Errorf("alipay trade is not paid")
	}
	totalAmount := normalizeAlipayNotifyAmount(queryReply.TotalAmount)
	expectedAmount := normalizeAlipayNotifyAmount(statusReply.GetTotalAmount())
	if expectedAmount == "" || expectedAmount != totalAmount {
		return nil, fmt.Errorf("alipay trade amount mismatch")
	}
	markReply, err := orderSvc.MarkPaymentPaid(ctx, &orderv1.MarkPaymentPaidRequest{
		OutTradeNo:    outTradeNo,
		AlipayTradeNo: strings.TrimSpace(queryReply.TradeNo),
		AppId:         alipayAppID(alipayCfg),
		TotalAmount:   totalAmount,
		TradeStatus:   tradeStatus,
		NotifyPayload: "active_sync=1&out_trade_no=" + outTradeNo,
	})
	if err != nil {
		return nil, err
	}
	return &paymentSyncReply{
		OrderID:       markReply.GetOrderId(),
		PaymentNo:     markReply.GetOutTradeNo(),
		PaymentStatus: markReply.GetStatus(),
		OrderStatus:   markReply.GetOrderStatus(),
		Synced:        true,
		Duplicated:    markReply.GetDuplicated(),
	}, nil
}

func alipayAppID(cfg *conf.Alipay) string {
	if cfg == nil {
		return configOrEnv("", "ALIPAY_APP_ID")
	}
	return configOrEnv(cfg.AppID, "ALIPAY_APP_ID")
}

func paymentSyncResponse(reply *paymentSyncReply) map[string]any {
	if reply == nil {
		return map[string]any{}
	}
	return map[string]any{
		"orderId":        int64String(reply.OrderID),
		"order_id":       int64String(reply.OrderID),
		"paymentNo":      reply.PaymentNo,
		"payment_no":     reply.PaymentNo,
		"paymentStatus":  reply.PaymentStatus,
		"payment_status": reply.PaymentStatus,
		"orderStatus":    reply.OrderStatus,
		"order_status":   reply.OrderStatus,
		"synced":         reply.Synced,
		"duplicated":     reply.Duplicated,
	}
}

func markPaymentPaidFromReturn(ctx khttp.Context, alipayCfg *conf.Alipay, orderSvc paymentOrderService, alipayFactory paymentAlipayFactory) (int64, error) {
	if ctx == nil || ctx.Request() == nil {
		return 0, nil
	}
	req := ctx.Request()
	query := req.URL.Query()
	outTradeNo := strings.TrimSpace(query.Get("out_trade_no"))
	if outTradeNo == "" {
		return 0, nil
	}
	if orderSvc == nil {
		return 0, fmt.Errorf("order service unavailable")
	}
	if alipayFactory == nil {
		alipayFactory = defaultPaymentAlipayFactory
	}
	client, err := alipayFactory(alipayCfg)
	if err != nil {
		return 0, err
	}
	if err := client.VerifyNotify(query); err != nil {
		return 0, err
	}
	alipayTradeNo := strings.TrimSpace(query.Get("trade_no"))
	totalAmount := normalizeAlipayNotifyAmount(query.Get("total_amount"))
	if alipayTradeNo == "" || totalAmount == "" {
		return 0, fmt.Errorf("alipay return missing required fields")
	}
	statusReply, err := orderSvc.GetPaymentStatus(ctx, &orderv1.GetPaymentStatusRequest{OutTradeNo: outTradeNo})
	if err != nil {
		return 0, err
	}
	expectedAmount := normalizeAlipayNotifyAmount(statusReply.GetTotalAmount())
	if expectedAmount == "" || expectedAmount != totalAmount {
		return 0, fmt.Errorf("alipay return amount mismatch")
	}
	reply, err := orderSvc.MarkPaymentPaid(ctx, &orderv1.MarkPaymentPaidRequest{
		OutTradeNo:    outTradeNo,
		AlipayTradeNo: alipayTradeNo,
		AppId:         strings.TrimSpace(query.Get("app_id")),
		TotalAmount:   totalAmount,
		TradeStatus:   "TRADE_SUCCESS",
		NotifyPayload: query.Encode(),
	})
	if err != nil {
		return 0, err
	}
	return reply.GetOrderId(), nil
}

func alipaySuccessHTML(orderID int64) string {
	detailButton := ""
	syncScript := ""
	if orderID > 0 {
		detailButton = fmt.Sprintf(`<button data-action="return-detail" onclick="location.href='/#/pages/orderDetail/orderDetail?id=%d'" style="height: 44px; padding: 0 18px; border: 0; border-radius: 8px; background: #1677ff; color: #fff; font-size: 15px;">查看订单详情</button>`, orderID)
		syncScript = fmt.Sprintf(`<script>
(function(){
  var paymentNo = new URLSearchParams(location.search).get('out_trade_no') || '';
  if (!paymentNo || !window.fetch) return;
  fetch('/carpool/orders/%d/payment/sync', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({payment_no: paymentNo})
  }).catch(function(){});
})();
</script>`, orderID)
	}
	return `<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>支付成功</title></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; padding: 24px; background: #f5f7fb; color: #172033;">
<main style="max-width: 480px; margin: 12vh auto 0; padding: 28px; border-radius: 16px; background: #fff; box-shadow: 0 12px 36px rgba(16, 24, 40, .08);">
<h1>支付成功</h1>
<p>订单支付结果已同步，您可以返回应用继续查看订单状态。</p>
<div style="display: flex; gap: 12px; flex-wrap: wrap;">
` + detailButton + `
<button data-action="return-orders" onclick="location.href='/#/pages/orders/orders'" style="height: 44px; padding: 0 18px; border: 1px solid #d0d5dd; border-radius: 8px; background: #fff; color: #344054; font-size: 15px;">返回订单列表</button>
<button data-action="return-home" onclick="location.href='/#/pages/home/home'" style="height: 44px; padding: 0 18px; border: 1px solid #d0d5dd; border-radius: 8px; background: #fff; color: #344054; font-size: 15px;">返回首页</button>
</div>
</main>
` + syncScript + `
</body>
</html>`
}

func writeAlipayNotifyResult(ctx khttp.Context, result string) error {
	ctx.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Response().WriteHeader(http.StatusOK)
	_, err := ctx.Response().Write([]byte(result))
	return err
}

func normalizeAlipayNotifyAmount(value string) string {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(float64(int64(amount*100+0.5))/100, 'f', 2, 64)
}
