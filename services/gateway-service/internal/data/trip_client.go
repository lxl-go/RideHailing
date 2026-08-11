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

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

type SearchTripsRequest struct {
	Origin      string
	Destination string
	DepartDate  string
	Page        int32
	PageSize    int32
}

type TripClient interface {
	SearchTrips(ctx context.Context, req SearchTripsRequest) (*tripv1.SearchTripsReply, error)
	GetTripDetail(ctx context.Context, id int64) (*tripv1.GetTripDetailReply, error)
	PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error)
	ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error)
	UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) error
	ListCoupons(ctx context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error)
	ClaimCoupon(ctx context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error)
	PublishDemand(ctx context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error)
	ListMyDemands(ctx context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error)
	CancelDemand(ctx context.Context, req *tripv1.CancelDemandRequest) error
	DeleteTrip(ctx context.Context, req *tripv1.DeleteTripRequest) error
	ValidateLocation(ctx context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error)
	PreviewTripPrice(ctx context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error)
	SuggestLocations(ctx context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error)
}

type TripHTTPClient struct {
	baseURL string
	client  *http.Client
}

type TripGRPCClient struct {
	client tripv1.TripServiceClient
}

func NewTripGRPCClient(client tripv1.TripServiceClient) *TripGRPCClient {
	return &TripGRPCClient{client: client}
}

func (c *TripGRPCClient) SearchTrips(ctx context.Context, req SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	return c.client.SearchTrips(ctx, &tripv1.SearchTripsRequest{
		Origin:      req.Origin,
		Destination: req.Destination,
		DepartDate:  req.DepartDate,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
}

func (c *TripGRPCClient) GetTripDetail(ctx context.Context, id int64) (*tripv1.GetTripDetailReply, error) {
	return c.client.GetTripDetail(ctx, &tripv1.GetTripDetailRequest{Id: id})
}

func (c *TripGRPCClient) PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	return c.client.PublishTrip(ctx, req)
}

func (c *TripGRPCClient) ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	return c.client.ListDriverTrips(ctx, req)
}

func (c *TripGRPCClient) UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) error {
	_, err := c.client.UpdateTripStatus(ctx, req)
	return err
}

func (c *TripGRPCClient) ListCoupons(ctx context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error) {
	return c.client.ListCoupons(ctx, req)
}

func (c *TripGRPCClient) ClaimCoupon(ctx context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error) {
	return c.client.ClaimCoupon(ctx, req)
}

func (c *TripGRPCClient) PublishDemand(ctx context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error) {
	return c.client.PublishDemand(ctx, req)
}

func (c *TripGRPCClient) ListMyDemands(ctx context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error) {
	return c.client.ListMyDemands(ctx, req)
}

func (c *TripGRPCClient) CancelDemand(ctx context.Context, req *tripv1.CancelDemandRequest) error {
	_, err := c.client.CancelDemand(ctx, req)
	return err
}

func (c *TripGRPCClient) DeleteTrip(ctx context.Context, req *tripv1.DeleteTripRequest) error {
	_, err := c.client.DeleteTrip(ctx, req)
	return err
}

func (c *TripGRPCClient) ValidateLocation(ctx context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error) {
	return c.client.ValidateLocation(ctx, req)
}

func (c *TripGRPCClient) PreviewTripPrice(ctx context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error) {
	return c.client.PreviewTripPrice(ctx, req)
}

func (c *TripGRPCClient) SuggestLocations(ctx context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error) {
	return c.client.SuggestLocations(ctx, req)
}

func NewTripHTTPClient(baseURL string) *TripHTTPClient {
	return &TripHTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TripHTTPClient) SearchTrips(ctx context.Context, req SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	values := url.Values{}
	values.Set("origin", req.Origin)
	values.Set("destination", req.Destination)
	values.Set("depart_date", req.DepartDate)
	values.Set("page", strconv.Itoa(int(req.Page)))
	values.Set("page_size", strconv.Itoa(int(req.PageSize)))
	var reply tripv1.SearchTripsReply
	if err := c.get(ctx, "/v1/trips?"+values.Encode(), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) GetTripDetail(ctx context.Context, id int64) (*tripv1.GetTripDetailReply, error) {
	var reply tripv1.GetTripDetailReply
	if err := c.get(ctx, fmt.Sprintf("/v1/trips/%d", id), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) PublishTrip(ctx context.Context, req *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	var reply tripv1.PublishTripReply
	if err := c.doJSON(ctx, http.MethodPost, "/v1/driver/trips", req, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) ListDriverTrips(ctx context.Context, req *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	values := url.Values{}
	values.Set("status", strconv.Itoa(int(req.Status)))
	values.Set("page", strconv.Itoa(int(req.Page)))
	values.Set("page_size", strconv.Itoa(int(req.PageSize)))
	values.Set("driver_id", strconv.FormatInt(req.DriverId, 10))
	var reply tripv1.ListDriverTripsReply
	if err := c.get(ctx, "/v1/driver/trips?"+values.Encode(), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) UpdateTripStatus(ctx context.Context, req *tripv1.UpdateTripStatusRequest) error {
	var reply tripv1.UpdateTripStatusReply
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/driver/trips/%d/status", req.Id), req, &reply)
}

func (c *TripHTTPClient) ListCoupons(ctx context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error) {
	values := url.Values{}
	values.Set("user_id", strconv.FormatInt(req.UserId, 10))
	values.Set("page", strconv.Itoa(int(req.Page)))
	values.Set("page_size", strconv.Itoa(int(req.PageSize)))
	var reply tripv1.ListCouponsReply
	if err := c.get(ctx, "/v1/coupons?"+values.Encode(), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) ClaimCoupon(ctx context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error) {
	var reply tripv1.ClaimCouponReply
	if err := c.doJSON(ctx, http.MethodPost, "/v1/coupons/claim", req, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) PublishDemand(ctx context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error) {
	var reply tripv1.PublishDemandReply
	if err := c.doJSON(ctx, http.MethodPost, "/v1/trips/demands", req, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) ListMyDemands(ctx context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error) {
	values := url.Values{}
	values.Set("passenger_id", strconv.FormatInt(req.PassengerId, 10))
	values.Set("status", strconv.Itoa(int(req.Status)))
	values.Set("page", strconv.Itoa(int(req.Page)))
	values.Set("page_size", strconv.Itoa(int(req.PageSize)))
	var reply tripv1.ListMyDemandsReply
	if err := c.get(ctx, "/v1/trips/demands/mine?"+values.Encode(), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) CancelDemand(ctx context.Context, req *tripv1.CancelDemandRequest) error {
	var reply tripv1.CancelDemandReply
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/trips/demands/%d/cancel", req.Id), req, &reply)
}

func (c *TripHTTPClient) DeleteTrip(ctx context.Context, req *tripv1.DeleteTripRequest) error {
	values := url.Values{}
	values.Set("driver_id", strconv.FormatInt(req.DriverId, 10))
	var reply tripv1.DeleteTripReply
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/v1/driver/trips/%d?%s", req.Id, values.Encode()), nil, &reply)
}

func (c *TripHTTPClient) ValidateLocation(ctx context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error) {
	var reply tripv1.ValidateLocationReply
	if err := c.doJSON(ctx, http.MethodPost, "/v1/driver/trips/locations/validate", req, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) PreviewTripPrice(ctx context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error) {
	var reply tripv1.PreviewTripPriceReply
	if err := c.doJSON(ctx, http.MethodPost, "/v1/driver/trips/price-preview", req, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) SuggestLocations(ctx context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error) {
	var reply tripv1.SuggestLocationsReply
	if err := c.doJSON(ctx, http.MethodPost, "/v1/driver/trips/locations/suggest", req, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *TripHTTPClient) get(ctx context.Context, path string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.execute(httpReq, out)
}

func (c *TripHTTPClient) doJSON(ctx context.Context, method string, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if in != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	return c.execute(httpReq, out)
}

func (c *TripHTTPClient) execute(req *http.Request, out any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
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
