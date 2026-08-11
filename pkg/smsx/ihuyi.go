package smsx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultIhuyiEndpoint = "https://api.ihuyi.com/sms/Submit.json"

type IhuyiConfig struct {
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	Endpoint string `yaml:"endpoint"`
}

type Response struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	SMSID string `json:"smsid"`
}

type IhuyiClient struct {
	cfg    IhuyiConfig
	client *http.Client
}

func NewIhuyiClient(cfg IhuyiConfig) *IhuyiClient {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		cfg.Endpoint = defaultIhuyiEndpoint
	}
	return &IhuyiClient{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *IhuyiClient) SendVerificationCode(ctx context.Context, mobile string, code string) error {
	content := fmt.Sprintf("您的验证码是：%s。请不要把验证码泄露给其他人。", strings.TrimSpace(code))
	return c.Send(ctx, mobile, content)
}

func (c *IhuyiClient) Send(ctx context.Context, mobile string, content string) error {
	form := url.Values{}
	form.Set("account", strings.TrimSpace(c.cfg.Account))
	form.Set("password", strings.TrimSpace(c.cfg.Password))
	form.Set("mobile", strings.TrimSpace(mobile))
	form.Set("content", strings.TrimSpace(content))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("ihuyi sms http status %d", resp.StatusCode)
	}

	var reply Response
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return err
	}
	if reply.Code != 2 {
		return fmt.Errorf("ihuyi sms rejected: code=%d msg=%s", reply.Code, reply.Msg)
	}
	return nil
}
