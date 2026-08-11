package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"ride-hailing/pkg/grpcx"
	"ride-hailing/services/gateway-service/internal/conf"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
)

type ReviewClient interface {
	SubmitReview(ctx context.Context, req *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error)
	GetMyReview(ctx context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error)
}

type ReviewGRPCClient struct {
	client reviewv1.ReviewServiceClient
}

func NewReviewGRPCClient(client reviewv1.ReviewServiceClient) *ReviewGRPCClient {
	return &ReviewGRPCClient{client: client}
}

func (c *ReviewGRPCClient) SubmitReview(ctx context.Context, req *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error) {
	return c.client.SubmitReview(ctx, req)
}

func (c *ReviewGRPCClient) GetMyReview(ctx context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error) {
	return c.client.GetMyReview(ctx, req)
}

type ReviewHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewReviewHTTPClient(baseURL string) *ReviewHTTPClient {
	return &ReviewHTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *ReviewHTTPClient) SubmitReview(ctx context.Context, req *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/reviews", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("review service returned status %d", resp.StatusCode)
	}
	var reply reviewv1.SubmitReviewReply
	if err := decodeProtoBody(resp, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *ReviewHTTPClient) GetMyReview(ctx context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error) {
	values := url.Values{}
	values.Set("from_user_id", strconv.FormatInt(req.FromUserId, 10))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/reviews/mine/%d?%s", c.baseURL, req.OrderId, values.Encode()), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("review service returned status %d", resp.StatusCode)
	}
	var reply reviewv1.GetMyReviewReply
	if err := decodeProtoBody(resp, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func decodeProtoBody(resp *http.Response, msg proto.Message) error {
	if msg == nil {
		return fmt.Errorf("nil decode target")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, msg)
}

func NewReviewClient(c *conf.Clients, discovery registry.Discovery) (ReviewClient, error) {
	baseURL := "http://127.0.0.1:9060"
	endpoint := ""
	if c != nil && c.Review != nil && c.Review.HTTPBaseURL != "" {
		baseURL = c.Review.HTTPBaseURL
	}
	if c != nil && c.Review != nil && c.Review.Endpoint != "" {
		endpoint = c.Review.Endpoint
	}
	if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
		conn, err := grpcx.DialInsecure(context.Background(), endpoint, discovery, grpcx.ClientOptions{})
		if err != nil {
			return nil, err
		}
		return NewReviewGRPCClient(reviewv1.NewReviewServiceClient(conn)), nil
	}
	return NewReviewHTTPClient(baseURL), nil
}
