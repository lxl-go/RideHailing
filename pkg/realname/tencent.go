package realname

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultTencentEndpoint = "https://ap-beijing.cloudmarket-apigw.com/service-hl92rg03/id/check"

var (
	ErrInvalidRequest = errors.New("invalid real-name request")
	ErrConfigMissing  = errors.New("real-name config missing")
)

type Request struct {
	RealName string
	IDCardNo string
}

type Result struct {
	ErrorCode int
	Reason    string
	Matched   bool
	RealName  string
	IDCardNo  string
	Province  string
	City      string
	District  string
	Area      string
	Sex       string
	Birthday  string
}

type Verifier interface {
	Verify(ctx context.Context, req Request) (Result, error)
}

type TencentConfig struct {
	SecretID  string
	SecretKey string
	Endpoint  string
	Timeout   time.Duration
}

type TencentClient struct {
	cfg        TencentConfig
	httpClient *http.Client
}

func NewTencentClient(cfg TencentConfig) *TencentClient {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultTencentEndpoint
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &TencentClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *TencentClient) Verify(ctx context.Context, req Request) (Result, error) {
	realName := strings.TrimSpace(req.RealName)
	idCardNo := strings.TrimSpace(req.IDCardNo)
	if realName == "" || idCardNo == "" {
		return Result{}, ErrInvalidRequest
	}
	if c.cfg.SecretID == "" || c.cfg.SecretKey == "" {
		return Result{}, ErrConfigMissing
	}

	auth, datetime := calcAuthorization(c.cfg.SecretID, c.cfg.SecretKey, time.Now)
	values := url.Values{}
	values.Set("cardNo", idCardNo)
	values.Set("realName", realName)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return Result{}, err
	}
	httpReq.Header.Set("Authorization", auth)
	httpReq.Header.Set("X-Date", datetime)
	httpReq.Header.Set("request-id", uuid.NewString())
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("real-name service returned status %d", resp.StatusCode)
	}
	return parseTencentResponse(body)
}

func calcAuthorization(secretID string, secretKey string, now func() time.Time) (auth string, datetime string) {
	timeLocation, _ := time.LoadLocation("Etc/GMT")
	datetime = now().In(timeLocation).Format("Mon, 02 Jan 2006 15:04:05 GMT")
	signStr := fmt.Sprintf("x-date: %s", datetime)
	mac := hmac.New(sha1.New, []byte(secretKey))
	_, _ = mac.Write([]byte(signStr))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	auth = fmt.Sprintf("{\"id\":\"%s\", \"x-date\":\"%s\", \"signature\":\"%s\"}", secretID, datetime, sign)
	return auth, datetime
}

type tencentResponse struct {
	ErrorCode int           `json:"error_code"`
	Reason    string        `json:"reason"`
	Result    tencentResult `json:"result"`
}

type tencentResult struct {
	RealName     string          `json:"realname"`
	IDCardNo     string          `json:"idcard"`
	Matched      bool            `json:"isok"`
	IDCardInfo   *tencentIDCard  `json:"IdCardInfor"`
	IDCardInfoLC *tencentIDCard  `json:"idCardInfor"`
	RawInfo      json.RawMessage `json:"-"`
}

type tencentIDCard struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Area     string `json:"area"`
	Sex      string `json:"sex"`
	Birthday string `json:"birthday"`
}

func parseTencentResponse(body []byte) (Result, error) {
	var payload tencentResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return Result{}, err
	}
	info := payload.Result.IDCardInfo
	if info == nil {
		info = payload.Result.IDCardInfoLC
	}
	out := Result{
		ErrorCode: payload.ErrorCode,
		Reason:    payload.Reason,
		Matched:   payload.ErrorCode == 0 && payload.Result.Matched,
		RealName:  payload.Result.RealName,
		IDCardNo:  payload.Result.IDCardNo,
	}
	if info != nil {
		out.Province = info.Province
		out.City = info.City
		out.District = info.District
		out.Area = info.Area
		out.Sex = info.Sex
		out.Birthday = info.Birthday
	}
	return out, nil
}
