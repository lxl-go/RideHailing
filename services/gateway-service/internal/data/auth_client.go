package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"ride-hailing/pkg/grpcx"
	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	"ride-hailing/services/gateway-service/internal/conf"
)

type AuthClient interface {
	SendLoginCode(ctx context.Context, req *authv1.SendLoginCodeRequest) (*authv1.SendLoginCodeReply, error)
	Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginReply, error)
	VerifyToken(ctx context.Context, authorization string) (*authv1.VerifyTokenReply, error)
	RefreshToken(ctx context.Context, refreshToken string) (*authv1.LoginReply, error)
	Logout(ctx context.Context, refreshToken string) (*authv1.LogoutReply, error)
	CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}

type AuthGRPCClient struct {
	client authv1.AuthServiceClient
}

func NewAuthGRPCClient(client authv1.AuthServiceClient) *AuthGRPCClient {
	return &AuthGRPCClient{client: client}
}

func (c *AuthGRPCClient) SendLoginCode(ctx context.Context, req *authv1.SendLoginCodeRequest) (*authv1.SendLoginCodeReply, error) {
	return c.client.SendLoginCode(ctx, req)
}

func (c *AuthGRPCClient) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginReply, error) {
	return c.client.Login(ctx, req)
}

func (c *AuthGRPCClient) VerifyToken(ctx context.Context, authorization string) (*authv1.VerifyTokenReply, error) {
	return c.client.VerifyToken(ctx, &authv1.VerifyTokenRequest{Authorization: authorization})
}

func (c *AuthGRPCClient) RefreshToken(ctx context.Context, refreshToken string) (*authv1.LoginReply, error) {
	return c.client.RefreshToken(ctx, &authv1.RefreshTokenRequest{RefreshToken: refreshToken})
}

func (c *AuthGRPCClient) Logout(ctx context.Context, refreshToken string) (*authv1.LogoutReply, error) {
	return c.client.Logout(ctx, &authv1.LogoutRequest{RefreshToken: refreshToken})
}

func (c *AuthGRPCClient) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	reply, err := c.client.CheckPermission(ctx, &authv1.CheckPermissionRequest{UserId: userID, PermissionCode: permissionCode})
	if err != nil {
		return false, err
	}
	return reply.Allowed, nil
}

type AuthHTTPClient struct {
	baseURL string
	client  *http.Client
}

type UpstreamHTTPError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *UpstreamHTTPError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return fmt.Sprintf("auth service returned status %d", e.StatusCode)
}

func NewAuthHTTPClient(baseURL string) *AuthHTTPClient {
	return &AuthHTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *AuthHTTPClient) SendLoginCode(ctx context.Context, req *authv1.SendLoginCodeRequest) (*authv1.SendLoginCodeReply, error) {
	var reply authv1.SendLoginCodeReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/sms/send", req, &reply)
	return &reply, err
}

func (c *AuthHTTPClient) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginReply, error) {
	var reply authv1.LoginReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/login", req, &reply)
	return &reply, err
}

func (c *AuthHTTPClient) VerifyToken(ctx context.Context, authorization string) (*authv1.VerifyTokenReply, error) {
	var reply authv1.VerifyTokenReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/verify", &authv1.VerifyTokenRequest{Authorization: authorization}, &reply)
	return &reply, err
}

func (c *AuthHTTPClient) RefreshToken(ctx context.Context, refreshToken string) (*authv1.LoginReply, error) {
	var reply authv1.LoginReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/refresh", &authv1.RefreshTokenRequest{RefreshToken: refreshToken}, &reply)
	return &reply, err
}

func (c *AuthHTTPClient) Logout(ctx context.Context, refreshToken string) (*authv1.LogoutReply, error) {
	var reply authv1.LogoutReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/logout", &authv1.LogoutRequest{RefreshToken: refreshToken}, &reply)
	return &reply, err
}

func (c *AuthHTTPClient) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	var reply authv1.CheckPermissionReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/permission/check", &authv1.CheckPermissionRequest{
		UserId:         userID,
		PermissionCode: permissionCode,
	}, &reply)
	if err != nil {
		return false, err
	}
	return reply.Allowed, nil
}

func (c *AuthHTTPClient) doJSON(ctx context.Context, method string, path string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newUpstreamHTTPError(resp.StatusCode, body)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if msg, ok := out.(proto.Message); ok {
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, msg)
	}
	return json.Unmarshal(body, out)
}

func newUpstreamHTTPError(statusCode int, body []byte) *UpstreamHTTPError {
	err := &UpstreamHTTPError{
		StatusCode: statusCode,
		Body:       string(body),
		Message:    http.StatusText(statusCode),
	}
	var payload struct {
		Message string `json:"message"`
		Msg     string `json:"msg"`
		Reason  string `json:"reason"`
	}
	if json.Unmarshal(body, &payload) == nil {
		switch {
		case strings.TrimSpace(payload.Message) != "":
			err.Message = strings.TrimSpace(payload.Message)
		case strings.TrimSpace(payload.Msg) != "":
			err.Message = strings.TrimSpace(payload.Msg)
		case strings.TrimSpace(payload.Reason) != "":
			err.Message = strings.TrimSpace(payload.Reason)
		}
	}
	return err
}

func NewAuthClient(c *conf.Clients, discovery registry.Discovery) (AuthClient, error) {
	baseURL := "http://127.0.0.1:9010"
	endpoint := ""
	if c != nil && c.Auth != nil && c.Auth.HTTPBaseURL != "" {
		baseURL = c.Auth.HTTPBaseURL
	}
	if c != nil && c.Auth != nil && c.Auth.Endpoint != "" {
		endpoint = c.Auth.Endpoint
	}
	if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
		conn, err := grpcx.DialInsecure(context.Background(), endpoint, discovery, grpcx.ClientOptions{})
		if err != nil {
			return nil, err
		}
		return NewAuthGRPCClient(authv1.NewAuthServiceClient(conn)), nil
	}
	return NewAuthHTTPClient(baseURL), nil
}
