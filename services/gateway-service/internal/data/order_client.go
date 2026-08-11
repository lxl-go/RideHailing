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
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type OrderClient interface {
	CreateOrder(ctx context.Context, tripID, passengerID int64, seatsBooked int32) (*orderv1.CreateOrderReply, error)
	CancelOrder(ctx context.Context, id, passengerID int64, idempotencyKey string) error
	ListOrders(ctx context.Context, passengerID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error)
	ListDriverOrders(ctx context.Context, driverID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error)
	GetOrderDetail(ctx context.Context, id, passengerID int64) (*orderv1.GetOrderDetailReply, error)
	PendingOrders(ctx context.Context, driverID, tripID int64, page, pageSize int32) (*orderv1.PendingOrdersReply, error)
	AcceptOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error
	RejectOrder(ctx context.Context, id, driverID int64, idempotencyKey, rejectReason string) error
	StartPickup(ctx context.Context, id, driverID int64, idempotencyKey string) error
	StartDelivery(ctx context.Context, id, driverID int64, idempotencyKey string) error
	CompleteOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error
	GetDriverIncome(ctx context.Context, driverID int64, startTime, endTime string, page, pageSize int32) (*orderv1.DriverIncomeReply, error)
	CreatePayment(ctx context.Context, orderID, passengerID int64, channel string) (*orderv1.CreatePaymentReply, error)
	MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error)
	GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error)
}

type OrderGRPCClient struct {
	client orderv1.OrderServiceClient
}

func NewOrderGRPCClient(client orderv1.OrderServiceClient) *OrderGRPCClient {
	return &OrderGRPCClient{client: client}
}

func (c *OrderGRPCClient) CreateOrder(ctx context.Context, tripID, passengerID int64, seatsBooked int32) (*orderv1.CreateOrderReply, error) {
	return c.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{TripId: tripID, PassengerId: passengerID, SeatsBooked: seatsBooked})
}

func (c *OrderGRPCClient) CancelOrder(ctx context.Context, id, passengerID int64, idempotencyKey string) error {
	_, err := c.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{Id: id, PassengerId: passengerID, IdempotencyKey: idempotencyKey})
	return err
}

func (c *OrderGRPCClient) ListOrders(ctx context.Context, passengerID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	return c.client.ListOrders(ctx, &orderv1.ListOrdersRequest{PassengerId: passengerID, Status: status, Page: page, PageSize: pageSize})
}

func (c *OrderGRPCClient) ListDriverOrders(ctx context.Context, driverID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	return c.client.ListOrders(ctx, &orderv1.ListOrdersRequest{DriverId: driverID, Status: status, Page: page, PageSize: pageSize})
}

func (c *OrderGRPCClient) GetOrderDetail(ctx context.Context, id, passengerID int64) (*orderv1.GetOrderDetailReply, error) {
	return c.client.GetOrderDetail(ctx, &orderv1.GetOrderDetailRequest{Id: id, PassengerId: passengerID})
}

func (c *OrderGRPCClient) PendingOrders(ctx context.Context, driverID, tripID int64, page, pageSize int32) (*orderv1.PendingOrdersReply, error) {
	return c.client.PendingOrders(ctx, &orderv1.PendingOrdersRequest{DriverId: driverID, TripId: tripID, Page: page, PageSize: pageSize})
}

func (c *OrderGRPCClient) AcceptOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	_, err := c.client.AcceptOrder(ctx, &orderv1.AcceptOrderRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey})
	return err
}

func (c *OrderGRPCClient) RejectOrder(ctx context.Context, id, driverID int64, idempotencyKey, rejectReason string) error {
	_, err := c.client.RejectOrder(ctx, &orderv1.RejectOrderRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey, RejectReason: rejectReason})
	return err
}

func (c *OrderGRPCClient) StartPickup(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	_, err := c.client.StartPickup(ctx, &orderv1.StartPickupRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey})
	return err
}

func (c *OrderGRPCClient) StartDelivery(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	_, err := c.client.StartDelivery(ctx, &orderv1.StartDeliveryRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey})
	return err
}

func (c *OrderGRPCClient) CompleteOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	_, err := c.client.CompleteOrder(ctx, &orderv1.CompleteOrderRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey})
	return err
}

func (c *OrderGRPCClient) GetDriverIncome(ctx context.Context, driverID int64, startTime, endTime string, page, pageSize int32) (*orderv1.DriverIncomeReply, error) {
	return c.client.GetDriverIncome(ctx, &orderv1.DriverIncomeRequest{DriverId: driverID, StartTime: startTime, EndTime: endTime, Page: page, PageSize: pageSize})
}

func (c *OrderGRPCClient) CreatePayment(ctx context.Context, orderID, passengerID int64, channel string) (*orderv1.CreatePaymentReply, error) {
	return c.client.CreatePayment(ctx, &orderv1.CreatePaymentRequest{OrderId: orderID, PassengerId: passengerID, Channel: channel})
}

func (c *OrderGRPCClient) MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error) {
	return c.client.MarkPaymentPaid(ctx, req)
}

func (c *OrderGRPCClient) GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error) {
	return c.client.GetPaymentStatus(ctx, req)
}

type OrderHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewOrderHTTPClient(baseURL string) *OrderHTTPClient {
	return &OrderHTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *OrderHTTPClient) CreateOrder(ctx context.Context, tripID, passengerID int64, seatsBooked int32) (*orderv1.CreateOrderReply, error) {
	var reply orderv1.CreateOrderReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/orders", &orderv1.CreateOrderRequest{TripId: tripID, PassengerId: passengerID, SeatsBooked: seatsBooked}, &reply)
	return &reply, err
}

func (c *OrderHTTPClient) CancelOrder(ctx context.Context, id, passengerID int64, idempotencyKey string) error {
	var reply orderv1.CancelOrderReply
	return c.doJSONWithHeaders(ctx, http.MethodPost, fmt.Sprintf("/v1/orders/%d/cancel", id), &orderv1.CancelOrderRequest{Id: id, PassengerId: passengerID, IdempotencyKey: idempotencyKey}, idempotencyHeaders(idempotencyKey), &reply)
}

func (c *OrderHTTPClient) ListOrders(ctx context.Context, passengerID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	values := url.Values{}
	values.Set("passenger_id", strconv.FormatInt(passengerID, 10))
	values.Set("status", strconv.Itoa(int(status)))
	values.Set("page", strconv.Itoa(int(page)))
	values.Set("page_size", strconv.Itoa(int(pageSize)))
	var reply orderv1.ListOrdersReply
	err := c.get(ctx, "/v1/orders?"+values.Encode(), &reply)
	return &reply, err
}

func (c *OrderHTTPClient) ListDriverOrders(ctx context.Context, driverID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	values := url.Values{}
	values.Set("driver_id", strconv.FormatInt(driverID, 10))
	values.Set("status", strconv.Itoa(int(status)))
	values.Set("page", strconv.Itoa(int(page)))
	values.Set("page_size", strconv.Itoa(int(pageSize)))
	var reply orderv1.ListOrdersReply
	err := c.get(ctx, "/v1/orders?"+values.Encode(), &reply)
	return &reply, err
}

func (c *OrderHTTPClient) GetOrderDetail(ctx context.Context, id, passengerID int64) (*orderv1.GetOrderDetailReply, error) {
	values := url.Values{}
	values.Set("passenger_id", strconv.FormatInt(passengerID, 10))
	var reply orderv1.GetOrderDetailReply
	err := c.get(ctx, fmt.Sprintf("/v1/orders/%d?%s", id, values.Encode()), &reply)
	return &reply, err
}

func (c *OrderHTTPClient) PendingOrders(ctx context.Context, driverID, tripID int64, page, pageSize int32) (*orderv1.PendingOrdersReply, error) {
	values := url.Values{}
	values.Set("driver_id", strconv.FormatInt(driverID, 10))
	values.Set("trip_id", strconv.FormatInt(tripID, 10))
	values.Set("page", strconv.Itoa(int(page)))
	values.Set("page_size", strconv.Itoa(int(pageSize)))
	var reply orderv1.PendingOrdersReply
	err := c.get(ctx, "/v1/driver/orders/pending?"+values.Encode(), &reply)
	return &reply, err
}

func (c *OrderHTTPClient) AcceptOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	var reply orderv1.AcceptOrderReply
	return c.doJSONWithHeaders(ctx, http.MethodPost, fmt.Sprintf("/v1/driver/orders/%d/accept", id), &orderv1.AcceptOrderRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey}, idempotencyHeaders(idempotencyKey), &reply)
}

func (c *OrderHTTPClient) RejectOrder(ctx context.Context, id, driverID int64, idempotencyKey, rejectReason string) error {
	var reply orderv1.RejectOrderReply
	return c.doJSONWithHeaders(ctx, http.MethodPost, fmt.Sprintf("/v1/driver/orders/%d/reject", id), &orderv1.RejectOrderRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey, RejectReason: rejectReason}, idempotencyHeaders(idempotencyKey), &reply)
}

func (c *OrderHTTPClient) StartPickup(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	var reply orderv1.StartPickupReply
	return c.doJSONWithHeaders(ctx, http.MethodPost, fmt.Sprintf("/v1/driver/orders/%d/start-pickup", id), &orderv1.StartPickupRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey}, idempotencyHeaders(idempotencyKey), &reply)
}

func (c *OrderHTTPClient) StartDelivery(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	var reply orderv1.StartDeliveryReply
	return c.doJSONWithHeaders(ctx, http.MethodPost, fmt.Sprintf("/v1/driver/orders/%d/start-delivery", id), &orderv1.StartDeliveryRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey}, idempotencyHeaders(idempotencyKey), &reply)
}

func (c *OrderHTTPClient) CompleteOrder(ctx context.Context, id, driverID int64, idempotencyKey string) error {
	var reply orderv1.CompleteOrderReply
	return c.doJSONWithHeaders(ctx, http.MethodPost, fmt.Sprintf("/v1/driver/orders/%d/complete", id), &orderv1.CompleteOrderRequest{Id: id, DriverId: driverID, IdempotencyKey: idempotencyKey}, idempotencyHeaders(idempotencyKey), &reply)
}

func (c *OrderHTTPClient) GetDriverIncome(ctx context.Context, driverID int64, startTime, endTime string, page, pageSize int32) (*orderv1.DriverIncomeReply, error) {
	values := url.Values{}
	values.Set("driver_id", strconv.FormatInt(driverID, 10))
	values.Set("start_time", startTime)
	values.Set("end_time", endTime)
	values.Set("page", strconv.Itoa(int(page)))
	values.Set("page_size", strconv.Itoa(int(pageSize)))
	var reply orderv1.DriverIncomeReply
	err := c.get(ctx, "/v1/driver/income?"+values.Encode(), &reply)
	return &reply, err
}

func (c *OrderHTTPClient) CreatePayment(ctx context.Context, orderID, passengerID int64, channel string) (*orderv1.CreatePaymentReply, error) {
	var reply orderv1.CreatePaymentReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/payments", &orderv1.CreatePaymentRequest{OrderId: orderID, PassengerId: passengerID, Channel: channel}, &reply)
	return &reply, err
}

func (c *OrderHTTPClient) MarkPaymentPaid(ctx context.Context, req *orderv1.MarkPaymentPaidRequest) (*orderv1.MarkPaymentPaidReply, error) {
	var reply orderv1.MarkPaymentPaidReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/payments/paid", req, &reply)
	return &reply, err
}

func (c *OrderHTTPClient) GetPaymentStatus(ctx context.Context, req *orderv1.GetPaymentStatusRequest) (*orderv1.GetPaymentStatusReply, error) {
	values := url.Values{}
	values.Set("out_trade_no", req.GetOutTradeNo())
	values.Set("order_id", strconv.FormatInt(req.GetOrderId(), 10))
	values.Set("passenger_id", strconv.FormatInt(req.GetPassengerId(), 10))
	var reply orderv1.GetPaymentStatusReply
	err := c.get(ctx, "/v1/payments/status?"+values.Encode(), &reply)
	return &reply, err
}

func (c *OrderHTTPClient) get(ctx context.Context, path string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.execute(httpReq, out)
}

func (c *OrderHTTPClient) doJSON(ctx context.Context, method string, path string, in any, out any) error {
	return c.doJSONWithHeaders(ctx, method, path, in, nil, out)
}

func (c *OrderHTTPClient) doJSONWithHeaders(ctx context.Context, method string, path string, in any, headers map[string]string, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}
	return c.execute(httpReq, out)
}

func (c *OrderHTTPClient) execute(req *http.Request, out any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return newUpstreamHTTPError(resp.StatusCode, body)
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

func idempotencyHeaders(idempotencyKey string) map[string]string {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return nil
	}
	return map[string]string{"Idempotency-Key": idempotencyKey}
}

func NewOrderClient(c *conf.Clients, discovery registry.Discovery) (OrderClient, error) {
	baseURL := "http://127.0.0.1:9050"
	endpoint := ""
	if c != nil && c.Order != nil && c.Order.HTTPBaseURL != "" {
		baseURL = c.Order.HTTPBaseURL
	}
	if c != nil && c.Order != nil && c.Order.Endpoint != "" {
		endpoint = c.Order.Endpoint
	}
	if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
		conn, err := grpcx.DialInsecure(context.Background(), endpoint, discovery, grpcx.ClientOptions{})
		if err != nil {
			return nil, err
		}
		return NewOrderGRPCClient(orderv1.NewOrderServiceClient(conn)), nil
	}
	return NewOrderHTTPClient(baseURL), nil
}
