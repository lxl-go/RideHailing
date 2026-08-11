package alipayx

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	sandboxGateway    = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
	productionGateway = "https://openapi.alipay.com/gateway.do"
	defaultTimeout    = 10 * time.Second
)

type Config struct {
	AppID           string
	PrivateKey      string
	AlipayPublicKey string
	Production      bool
	NotifyURL       string
	ReturnURL       string
	Timeout         time.Duration
	GatewayURL      string
}

type Client struct {
	cfg        Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	httpClient *http.Client
	now        func() time.Time
}

type WapPayRequest struct {
	OutTradeNo  string
	Subject     string
	TotalAmount string
	ProductCode string
}

type TradeQueryReply struct {
	OutTradeNo  string
	TradeNo     string
	TradeStatus string
	TotalAmount string
}

func NewClient(cfg Config) (*Client, error) {
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.PrivateKey = strings.TrimSpace(cfg.PrivateKey)
	cfg.AlipayPublicKey = strings.TrimSpace(cfg.AlipayPublicKey)
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.AppID == "" {
		return nil, errors.New("alipay app id is required")
	}
	if cfg.PrivateKey == "" {
		return nil, errors.New("alipay private key is required")
	}
	privateKey, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	var publicKey *rsa.PublicKey
	if cfg.AlipayPublicKey != "" {
		publicKey, err = parsePublicKey(cfg.AlipayPublicKey)
		if err != nil {
			return nil, err
		}
	}
	return &Client{cfg: cfg, privateKey: privateKey, publicKey: publicKey, httpClient: &http.Client{Timeout: cfg.Timeout}, now: time.Now}, nil
}

func (c *Client) CreateWapPay(ctx context.Context, req WapPayRequest) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	req.OutTradeNo = strings.TrimSpace(req.OutTradeNo)
	req.Subject = strings.TrimSpace(req.Subject)
	req.TotalAmount = strings.TrimSpace(req.TotalAmount)
	if req.OutTradeNo == "" || req.Subject == "" || req.TotalAmount == "" {
		return "", errors.New("out trade no, subject and total amount are required")
	}
	productCode := strings.TrimSpace(req.ProductCode)
	if productCode == "" {
		productCode = "QUICK_WAP_WAY"
	}
	bizContent, err := json.Marshal(map[string]string{
		"out_trade_no": req.OutTradeNo,
		"product_code": productCode,
		"total_amount": req.TotalAmount,
		"subject":      req.Subject,
	})
	if err != nil {
		return "", err
	}
	params := map[string]string{
		"app_id":      c.cfg.AppID,
		"method":      "alipay.trade.wap.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   c.now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContent),
	}
	if strings.TrimSpace(c.cfg.NotifyURL) != "" {
		params["notify_url"] = strings.TrimSpace(c.cfg.NotifyURL)
	}
	if strings.TrimSpace(c.cfg.ReturnURL) != "" {
		params["return_url"] = strings.TrimSpace(c.cfg.ReturnURL)
	}
	sign, err := c.sign(params)
	if err != nil {
		return "", err
	}
	params["sign"] = sign
	return buildAutoSubmitForm(c.gateway(), params), nil
}

func (c *Client) TradeQuery(ctx context.Context, outTradeNo string) (*TradeQueryReply, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	outTradeNo = strings.TrimSpace(outTradeNo)
	if outTradeNo == "" {
		return nil, errors.New("out trade no is required")
	}
	bizContent, err := json.Marshal(map[string]string{"out_trade_no": outTradeNo})
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"app_id":      c.cfg.AppID,
		"method":      "alipay.trade.query",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   c.now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContent),
	}
	sign, err := c.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gateway(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("alipay trade query returned status %d", resp.StatusCode)
	}
	var payload struct {
		Response struct {
			Code        string `json:"code"`
			Msg         string `json:"msg"`
			SubMsg      string `json:"sub_msg"`
			OutTradeNo  string `json:"out_trade_no"`
			TradeNo     string `json:"trade_no"`
			TradeStatus string `json:"trade_status"`
			TotalAmount string `json:"total_amount"`
		} `json:"alipay_trade_query_response"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Response.Code != "10000" {
		msg := strings.TrimSpace(payload.Response.SubMsg)
		if msg == "" {
			msg = strings.TrimSpace(payload.Response.Msg)
		}
		if msg == "" {
			msg = payload.Response.Code
		}
		return nil, fmt.Errorf("alipay trade query failed: %s", msg)
	}
	return &TradeQueryReply{
		OutTradeNo:  strings.TrimSpace(payload.Response.OutTradeNo),
		TradeNo:     strings.TrimSpace(payload.Response.TradeNo),
		TradeStatus: strings.TrimSpace(payload.Response.TradeStatus),
		TotalAmount: strings.TrimSpace(payload.Response.TotalAmount),
	}, nil
}

func (c *Client) VerifyNotify(values map[string][]string) error {
	if c.publicKey == nil {
		return errors.New("alipay public key is required")
	}
	params := flattenValues(values)
	sign := strings.TrimSpace(params["sign"])
	if sign == "" {
		return errors.New("alipay notify sign is required")
	}
	if params["app_id"] != c.cfg.AppID {
		return fmt.Errorf("alipay app id mismatch: %s", params["app_id"])
	}
	delete(params, "sign")
	delete(params, "sign_type")
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("decode alipay notify sign: %w", err)
	}
	digest := sha256.Sum256([]byte(canonicalString(params)))
	if err := rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return fmt.Errorf("verify alipay notify: %w", err)
	}
	return nil
}

func (c *Client) sign(params map[string]string) (string, error) {
	digest := sha256.Sum256([]byte(canonicalString(params)))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (c *Client) gateway() string {
	if endpoint := strings.TrimSpace(c.cfg.GatewayURL); endpoint != "" {
		return endpoint
	}
	if c.cfg.Production {
		return productionGateway
	}
	return sandboxGateway
}

func canonicalString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || strings.TrimSpace(value) == "" {
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

func buildAutoSubmitForm(endpoint string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(`<form id="alipay_submit" name="alipay_submit" action="`)
	b.WriteString(html.EscapeString(endpoint))
	b.WriteString(`" method="POST">`)
	for _, key := range keys {
		b.WriteString(`<input type="hidden" name="`)
		b.WriteString(html.EscapeString(key))
		b.WriteString(`" value="`)
		b.WriteString(html.EscapeString(params[key]))
		b.WriteString(`"/>`)
	}
	b.WriteString(`</form><script>document.forms['alipay_submit'].submit();</script>`)
	return b.String()
}

func parsePrivateKey(value string) (*rsa.PrivateKey, error) {
	raw, err := decodeKey(value)
	if err != nil {
		return nil, fmt.Errorf("decode alipay private key: %w", err)
	}
	if key, err := x509.ParsePKCS1PrivateKey(raw); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(raw)
	if err != nil {
		return nil, fmt.Errorf("parse alipay private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("alipay private key must be RSA")
	}
	return key, nil
}

func parsePublicKey(value string) (*rsa.PublicKey, error) {
	raw, err := decodeKey(value)
	if err != nil {
		return nil, fmt.Errorf("decode alipay public key: %w", err)
	}
	if key, err := x509.ParsePKCS1PublicKey(raw); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return nil, fmt.Errorf("parse alipay public key: %w", err)
	}
	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("alipay public key must be RSA")
	}
	return key, nil
}

func decodeKey(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if block, _ := pem.Decode([]byte(value)); block != nil {
		return block.Bytes, nil
	}
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")
	return base64.StdEncoding.DecodeString(value)
}

func flattenValues(values map[string][]string) map[string]string {
	out := make(map[string]string, len(values))
	for key, list := range values {
		if len(list) > 0 {
			out[key] = strings.TrimSpace(list[0])
		}
	}
	return out
}

func EncodeParams(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values.Encode()
}
