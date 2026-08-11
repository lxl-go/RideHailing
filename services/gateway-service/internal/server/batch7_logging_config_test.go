package server

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestBatch7RequestTraceID(t *testing.T) {
	req := httptest.NewRequest("POST", "/carpool/orders/123/pay", nil)
	req.Header.Set("X-Trace-Id", "trace-passenger-pay-001")

	if got := requestTraceID(req); got != "trace-passenger-pay-001" {
		t.Fatalf("requestTraceID() = %q, want %q", got, "trace-passenger-pay-001")
	}
}

func TestBatch7GatewayFinalSandboxConfig(t *testing.T) {
	raw, err := os.ReadFile("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("read gateway config: %v", err)
	}
	content := string(raw)

	for _, placeholder := range []string{"${ALIPAY_APP_ID}", "${ALIPAY_PRIVATE_KEY}", "${ALIPAY_PUBLIC_KEY}", "${ALIPAY_NOTIFY_URL}", "${ALIPAY_RETURN_URL}", "${AMAP_WEB_KEY}"} {
		if strings.Contains(content, placeholder) {
			t.Fatalf("gateway config still contains placeholder %s", placeholder)
		}
	}
	for _, want := range []string{
		`app_id: "9021000161685035"`,
		`notify_url: "http://417391e.r21.vip.cpolar.cn/api/v1/pay/notify"`,
		`return_url: "http://417391e.r21.vip.cpolar.cn/pay/success"`,
		`web_key: "22ba26c4d757d904aef8138acda60ab7"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("gateway config missing %s", want)
		}
	}
}
