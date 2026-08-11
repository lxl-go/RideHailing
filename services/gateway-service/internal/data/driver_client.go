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
	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/gateway-service/internal/conf"
)

type DriverClient interface {
	EnsureDriver(ctx context.Context, id int64, phone string) (*driverv1.DriverProfileReply, error)
	GetDriver(ctx context.Context, id int64) (*driverv1.DriverProfileReply, error)
	UpdateDriver(ctx context.Context, req *driverv1.UpdateDriverRequest) (*driverv1.DriverProfileReply, error)
	SubmitCertification(ctx context.Context, req *driverv1.SubmitCertificationRequest) (*driverv1.CertificationReply, error)
	GetCertification(ctx context.Context, id int64) (*driverv1.CertificationReply, error)
	SaveVehicle(ctx context.Context, req *driverv1.SaveVehicleRequest) (*driverv1.VehicleReply, error)
	UpdateVehicle(ctx context.Context, req *driverv1.UpdateVehicleRequest) (*driverv1.VehicleReply, error)
	DeleteVehicle(ctx context.Context, req *driverv1.DeleteVehicleRequest) (*driverv1.DeleteVehicleReply, error)
	ListVehicles(ctx context.Context, id int64) (*driverv1.ListVehiclesReply, error)
	ListMessages(ctx context.Context, driverID int64) (*driverv1.ListMessagesReply, error)
	AckMessage(ctx context.Context, req *driverv1.AckMessageRequest) (*driverv1.AckMessageReply, error)
	ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error)
	ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error)
}

type DriverGRPCClient struct {
	client driverv1.DriverServiceClient
}

func NewDriverGRPCClient(client driverv1.DriverServiceClient) *DriverGRPCClient {
	return &DriverGRPCClient{client: client}
}

func (c *DriverGRPCClient) EnsureDriver(ctx context.Context, id int64, phone string) (*driverv1.DriverProfileReply, error) {
	return c.client.EnsureDriver(ctx, &driverv1.EnsureDriverRequest{Id: id, Phone: phone})
}

func (c *DriverGRPCClient) GetDriver(ctx context.Context, id int64) (*driverv1.DriverProfileReply, error) {
	return c.client.GetDriver(ctx, &driverv1.GetDriverRequest{Id: id})
}

func (c *DriverGRPCClient) UpdateDriver(ctx context.Context, req *driverv1.UpdateDriverRequest) (*driverv1.DriverProfileReply, error) {
	return c.client.UpdateDriver(ctx, req)
}

func (c *DriverGRPCClient) SubmitCertification(ctx context.Context, req *driverv1.SubmitCertificationRequest) (*driverv1.CertificationReply, error) {
	return c.client.SubmitCertification(ctx, req)
}

func (c *DriverGRPCClient) GetCertification(ctx context.Context, id int64) (*driverv1.CertificationReply, error) {
	return c.client.GetCertification(ctx, &driverv1.GetCertificationRequest{Id: id})
}

func (c *DriverGRPCClient) SaveVehicle(ctx context.Context, req *driverv1.SaveVehicleRequest) (*driverv1.VehicleReply, error) {
	return c.client.SaveVehicle(ctx, req)
}

func (c *DriverGRPCClient) UpdateVehicle(ctx context.Context, req *driverv1.UpdateVehicleRequest) (*driverv1.VehicleReply, error) {
	return c.client.UpdateVehicle(ctx, req)
}

func (c *DriverGRPCClient) DeleteVehicle(ctx context.Context, req *driverv1.DeleteVehicleRequest) (*driverv1.DeleteVehicleReply, error) {
	return c.client.DeleteVehicle(ctx, req)
}

func (c *DriverGRPCClient) ListVehicles(ctx context.Context, id int64) (*driverv1.ListVehiclesReply, error) {
	return c.client.ListVehicles(ctx, &driverv1.ListVehiclesRequest{Id: id})
}

func (c *DriverGRPCClient) ListMessages(ctx context.Context, driverID int64) (*driverv1.ListMessagesReply, error) {
	return c.client.ListMessages(ctx, &driverv1.ListMessagesRequest{DriverId: driverID})
}

func (c *DriverGRPCClient) AckMessage(ctx context.Context, req *driverv1.AckMessageRequest) (*driverv1.AckMessageReply, error) {
	return c.client.AckMessage(ctx, req)
}

func (c *DriverGRPCClient) ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error) {
	return c.client.ReportDriverLocation(ctx, req)
}

func (c *DriverGRPCClient) ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error) {
	return c.client.ReplayDriverTrack(ctx, req)
}

type DriverHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewDriverHTTPClient(baseURL string) *DriverHTTPClient {
	return &DriverHTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *DriverHTTPClient) EnsureDriver(ctx context.Context, id int64, phone string) (*driverv1.DriverProfileReply, error) {
	var reply driverv1.DriverProfileReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/drivers/ensure", &driverv1.EnsureDriverRequest{Id: id, Phone: phone}, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) GetDriver(ctx context.Context, id int64) (*driverv1.DriverProfileReply, error) {
	var reply driverv1.DriverProfileReply
	err := c.get(ctx, fmt.Sprintf("/v1/drivers/%d", id), &reply)
	return &reply, err
}

func (c *DriverHTTPClient) UpdateDriver(ctx context.Context, req *driverv1.UpdateDriverRequest) (*driverv1.DriverProfileReply, error) {
	var reply driverv1.DriverProfileReply
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/drivers/%d", req.Id), req, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) SubmitCertification(ctx context.Context, req *driverv1.SubmitCertificationRequest) (*driverv1.CertificationReply, error) {
	var reply driverv1.CertificationReply
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/drivers/%d/certification", req.Id), req, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) GetCertification(ctx context.Context, id int64) (*driverv1.CertificationReply, error) {
	var reply driverv1.CertificationReply
	err := c.get(ctx, fmt.Sprintf("/v1/drivers/%d/certification", id), &reply)
	return &reply, err
}

func (c *DriverHTTPClient) SaveVehicle(ctx context.Context, req *driverv1.SaveVehicleRequest) (*driverv1.VehicleReply, error) {
	var reply driverv1.VehicleReply
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/drivers/%d/vehicles", req.Id), req, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) UpdateVehicle(ctx context.Context, req *driverv1.UpdateVehicleRequest) (*driverv1.VehicleReply, error) {
	var reply driverv1.VehicleReply
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/drivers/%d/vehicles/%d", req.DriverId, req.VehicleId), req, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) DeleteVehicle(ctx context.Context, req *driverv1.DeleteVehicleRequest) (*driverv1.DeleteVehicleReply, error) {
	var reply driverv1.DeleteVehicleReply
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+fmt.Sprintf("/v1/drivers/%d/vehicles/%d", req.DriverId, req.VehicleId), nil)
	if err != nil {
		return nil, err
	}
	err = c.execute(httpReq, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) ListVehicles(ctx context.Context, id int64) (*driverv1.ListVehiclesReply, error) {
	var reply driverv1.ListVehiclesReply
	err := c.get(ctx, fmt.Sprintf("/v1/drivers/%d/vehicles", id), &reply)
	return &reply, err
}

func (c *DriverHTTPClient) ListMessages(ctx context.Context, driverID int64) (*driverv1.ListMessagesReply, error) {
	var reply driverv1.ListMessagesReply
	err := c.get(ctx, fmt.Sprintf("/v1/drivers/%d/messages", driverID), &reply)
	return &reply, err
}

func (c *DriverHTTPClient) AckMessage(ctx context.Context, req *driverv1.AckMessageRequest) (*driverv1.AckMessageReply, error) {
	var reply driverv1.AckMessageReply
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/drivers/%d/messages/%d/ack", req.DriverId, req.MessageId), req, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error) {
	var reply driverv1.DriverLocationReply
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/drivers/%d/location/report", req.DriverId), req, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error) {
	var reply driverv1.ReplayDriverTrackReply
	path := fmt.Sprintf(
		"/v1/drivers/%d/track/replay?orderId=%d&order_id=%d&page=%d&pageSize=%d&page_size=%d",
		req.DriverId,
		req.OrderId,
		req.OrderId,
		req.Page,
		req.PageSize,
		req.PageSize,
	)
	err := c.get(ctx, path, &reply)
	return &reply, err
}

func (c *DriverHTTPClient) get(ctx context.Context, path string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.execute(httpReq, out)
}

func (c *DriverHTTPClient) doJSON(ctx context.Context, method string, path string, in any, out any) error {
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

func (c *DriverHTTPClient) execute(req *http.Request, out any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newUpstreamHTTPError(resp.StatusCode, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if msg, ok := out.(proto.Message); ok {
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, msg)
	}
	return json.Unmarshal(body, out)
}

func NewDriverClient(c *conf.Clients, discovery registry.Discovery) (DriverClient, error) {
	baseURL := "http://127.0.0.1:9030"
	endpoint := ""
	if c != nil && c.Driver != nil && c.Driver.HTTPBaseURL != "" {
		baseURL = c.Driver.HTTPBaseURL
	}
	if c != nil && c.Driver != nil && c.Driver.Endpoint != "" {
		endpoint = c.Driver.Endpoint
	}
	if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
		conn, err := grpcx.DialInsecure(context.Background(), endpoint, discovery, grpcx.ClientOptions{})
		if err != nil {
			return nil, err
		}
		return NewDriverGRPCClient(driverv1.NewDriverServiceClient(conn)), nil
	}
	return NewDriverHTTPClient(baseURL), nil
}
