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
	"ride-hailing/services/gateway-service/internal/conf"
	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
)

type PassengerClient interface {
	EnsurePassenger(ctx context.Context, id int64, phone string) (*passengerv1.PassengerProfileReply, error)
	GetPassenger(ctx context.Context, id int64) (*passengerv1.PassengerProfileReply, error)
	UpdatePassenger(ctx context.Context, req *passengerv1.UpdatePassengerRequest) (*passengerv1.PassengerProfileReply, error)
}

type PassengerGRPCClient struct {
	client passengerv1.PassengerServiceClient
}

func NewPassengerGRPCClient(client passengerv1.PassengerServiceClient) *PassengerGRPCClient {
	return &PassengerGRPCClient{client: client}
}

func (c *PassengerGRPCClient) EnsurePassenger(ctx context.Context, id int64, phone string) (*passengerv1.PassengerProfileReply, error) {
	return c.client.EnsurePassenger(ctx, &passengerv1.EnsurePassengerRequest{Id: id, Phone: phone})
}

func (c *PassengerGRPCClient) GetPassenger(ctx context.Context, id int64) (*passengerv1.PassengerProfileReply, error) {
	return c.client.GetPassenger(ctx, &passengerv1.GetPassengerRequest{Id: id})
}

func (c *PassengerGRPCClient) UpdatePassenger(ctx context.Context, req *passengerv1.UpdatePassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	return c.client.UpdatePassenger(ctx, req)
}

type PassengerHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewPassengerHTTPClient(baseURL string) *PassengerHTTPClient {
	return &PassengerHTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *PassengerHTTPClient) EnsurePassenger(ctx context.Context, id int64, phone string) (*passengerv1.PassengerProfileReply, error) {
	var reply passengerv1.PassengerProfileReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/passengers/ensure", &passengerv1.EnsurePassengerRequest{Id: id, Phone: phone}, &reply)
	return &reply, err
}

func (c *PassengerHTTPClient) GetPassenger(ctx context.Context, id int64) (*passengerv1.PassengerProfileReply, error) {
	var reply passengerv1.PassengerProfileReply
	err := c.get(ctx, fmt.Sprintf("/v1/passengers/%d", id), &reply)
	return &reply, err
}

func (c *PassengerHTTPClient) UpdatePassenger(ctx context.Context, req *passengerv1.UpdatePassengerRequest) (*passengerv1.PassengerProfileReply, error) {
	var reply passengerv1.PassengerProfileReply
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/passengers/%d", req.Id), req, &reply)
	return &reply, err
}

func (c *PassengerHTTPClient) get(ctx context.Context, path string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.execute(httpReq, out)
}

func (c *PassengerHTTPClient) doJSON(ctx context.Context, method string, path string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return c.execute(httpReq, out)
}

func (c *PassengerHTTPClient) execute(req *http.Request, out any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("passenger service returned status %d", resp.StatusCode)
	}
	if msg, ok := out.(proto.Message); ok {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, msg)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func NewPassengerClient(c *conf.Clients, discovery registry.Discovery) (PassengerClient, error) {
	baseURL := "http://127.0.0.1:9020"
	endpoint := ""
	if c != nil && c.Passenger != nil && c.Passenger.HTTPBaseURL != "" {
		baseURL = c.Passenger.HTTPBaseURL
	}
	if c != nil && c.Passenger != nil && c.Passenger.Endpoint != "" {
		endpoint = c.Passenger.Endpoint
	}
	if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
		conn, err := grpcx.DialInsecure(context.Background(), endpoint, discovery, grpcx.ClientOptions{})
		if err != nil {
			return nil, err
		}
		return NewPassengerGRPCClient(passengerv1.NewPassengerServiceClient(conn)), nil
	}
	return NewPassengerHTTPClient(baseURL), nil
}
