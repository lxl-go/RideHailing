package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ride-hailing/pkg/cozex"
	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

type fakeMobileAIClient struct {
	chatText string
	routeReq cozex.RainRouteWorkflowRequest
}

func (f *fakeMobileAIClient) CallTravelBot(_ context.Context, req cozex.TravelBotRequest) (*cozex.TravelBotResponse, error) {
	f.chatText = req.Text
	return &cozex.TravelBotResponse{RawBody: `{"answer":"ok"}`}, nil
}

func (f *fakeMobileAIClient) CallRainRouteWorkflow(_ context.Context, req cozex.RainRouteWorkflowRequest) (*cozex.RainRouteWorkflowResponse, error) {
	f.routeReq = req
	return &cozex.RainRouteWorkflowResponse{RawBody: `{"route":"safe"}`}, nil
}

type fakeMobileDriverTrackingService struct {
	reportReq *driverv1.ReportDriverLocationRequest
	replayReq *driverv1.ReplayDriverTrackRequest
}

type fakeMobileOrderService struct {
	detailReqPassengerID        int64
	detailDriverID              int64
	rejectReqDriverID           int64
	rejectReason                string
	pendingReqDriverID          int64
	listDriverOrdersReqDriverID int64
	listDriverOrdersReqStatus   int32
	completeReqDriverID         int64
	startPickupReqDriverID      int64
	startDeliveryReqDriverID    int64
	lastIdempotencyKey          string
	acceptErr                   error
	incomeReply                 *orderv1.DriverIncomeReply
}

func (f *fakeMobileOrderService) CreateOrder(context.Context, int64, int64, int32) (*orderv1.CreateOrderReply, error) {
	return &orderv1.CreateOrderReply{}, nil
}

func (f *fakeMobileOrderService) CancelOrder(context.Context, int64, int64, string) error {
	return nil
}

func (f *fakeMobileOrderService) ListOrders(context.Context, int64, int32, int32, int32) (*orderv1.ListOrdersReply, error) {
	return &orderv1.ListOrdersReply{}, nil
}

func (f *fakeMobileOrderService) ListDriverOrders(_ context.Context, driverID int64, status, page, pageSize int32) (*orderv1.ListOrdersReply, error) {
	f.listDriverOrdersReqDriverID = driverID
	f.listDriverOrdersReqStatus = status
	return &orderv1.ListOrdersReply{
		Total: 1,
		Items: []*orderv1.OrderItem{{
			Id:          5001,
			TripId:      1001,
			PassengerId: 3001,
			DriverId:    driverID,
			Origin:      "Shanghai Station",
			Destination: "Hongqiao Airport",
			DepartTime:  "2026-08-11T09:30:00+08:00",
			Status:      status,
		}},
	}, nil
}

func (f *fakeMobileOrderService) GetOrderDetail(_ context.Context, id, passengerID int64) (*orderv1.GetOrderDetailReply, error) {
	f.detailReqPassengerID = passengerID
	driverID := f.detailDriverID
	if driverID == -1 {
		driverID = 0
	} else if driverID <= 0 {
		driverID = passengerID
	}
	return &orderv1.GetOrderDetailReply{Order: &orderv1.OrderItem{
		Id:          id,
		TripId:      1001,
		PassengerId: 3001,
		DriverId:    driverID,
		Origin:      "Shanghai Station",
		Destination: "Hongqiao Airport",
		DepartTime:  "2026-08-11T09:30:00+08:00",
		Status:      1,
	}}, nil
}

func (f *fakeMobileOrderService) PendingOrders(_ context.Context, driverID, tripID int64, page, pageSize int32) (*orderv1.PendingOrdersReply, error) {
	f.pendingReqDriverID = driverID
	return &orderv1.PendingOrdersReply{
		Total: 1,
		Items: []*orderv1.PendingOrderItem{{Id: 5001, TripId: tripID, PassengerId: 3001, Status: 0}},
	}, nil
}

func (f *fakeMobileOrderService) AcceptOrder(_ context.Context, _ int64, _ int64, idempotencyKey string) error {
	f.lastIdempotencyKey = idempotencyKey
	return f.acceptErr
}

func (f *fakeMobileOrderService) RejectOrder(_ context.Context, _ int64, driverID int64, idempotencyKey, rejectReason string) error {
	f.rejectReqDriverID = driverID
	f.lastIdempotencyKey = idempotencyKey
	f.rejectReason = rejectReason
	return nil
}

func (f *fakeMobileOrderService) GetDriverIncome(context.Context, int64, string, string, int32, int32) (*orderv1.DriverIncomeReply, error) {
	if f.incomeReply != nil {
		return f.incomeReply, nil
	}
	return &orderv1.DriverIncomeReply{}, nil
}

func (f *fakeMobileOrderService) StartPickup(_ context.Context, _ int64, driverID int64, idempotencyKey string) error {
	f.startPickupReqDriverID = driverID
	f.lastIdempotencyKey = idempotencyKey
	return nil
}

func (f *fakeMobileOrderService) StartDelivery(_ context.Context, _ int64, driverID int64, idempotencyKey string) error {
	f.startDeliveryReqDriverID = driverID
	f.lastIdempotencyKey = idempotencyKey
	return nil
}

func (f *fakeMobileOrderService) CompleteOrder(_ context.Context, _ int64, driverID int64, idempotencyKey string) error {
	f.completeReqDriverID = driverID
	f.lastIdempotencyKey = idempotencyKey
	return nil
}

func (f *fakeMobileDriverTrackingService) ReportDriverLocation(_ context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error) {
	copy := *req
	f.reportReq = &copy
	return &driverv1.DriverLocationReply{Location: &driverv1.DriverLocationPoint{
		DriverId:   req.DriverId,
		OrderId:    req.OrderId,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		ReportedAt: req.ReportedAt,
	}}, nil
}

func (f *fakeMobileDriverTrackingService) ReplayDriverTrack(_ context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error) {
	copy := *req
	f.replayReq = &copy
	return &driverv1.ReplayDriverTrackReply{
		Total: 3,
		Items: []*driverv1.DriverLocationPoint{
			{DriverId: req.DriverId, OrderId: req.OrderId, Latitude: 31.2304, Longitude: 121.4737, ReportedAt: "2026-07-31T10:00:00+08:00"},
		},
	}, nil
}

func TestMobileAIRoutesCallCozeClient(t *testing.T) {
	srv := khttp.NewServer()
	aiClient := &fakeMobileAIClient{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, nil, aiClient, newMobileTrackStore())

	body := `{"sessionId":"s1","text":"下雨天怎么叫车","userId":1001,"userRole":"passenger"}`
	res := doGatewayJSON(srv, http.MethodPost, "/api/v1/ai/chat", body)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "下雨天怎么叫车", aiClient.chatText)
	require.Equal(t, float64(0), decodeGatewayBody(t, res)["code"])

	body = `{"sessionId":"s1","origin":"A","destination":"B","city":"上海","weather":"暴雨","avoid":"积水","preference":"安全","userRole":"passenger"}`
	res = doGatewayJSON(srv, http.MethodPost, "/api/v1/ai/rain-route", body)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "A", aiClient.routeReq.Origin)
	require.Equal(t, "B", aiClient.routeReq.Destination)
	require.Equal(t, "passenger", aiClient.routeReq.UserRole)
	require.Equal(t, float64(0), decodeGatewayBody(t, res)["code"])
}

func TestDriverTravelAdviceForwardsToPythonAgent(t *testing.T) {
	var forwarded map[string]any
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/travel-advice", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&forwarded))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"riskLevel":     "medium",
			"summary":       "Rain may affect this trip.",
			"weatherAdvice": []string{"Check rain warning"},
			"routeAdvice":   []string{"Leave 10 minutes earlier"},
			"safetyAdvice":  []string{"Keep distance"},
			"displayText":   "AI advice ready",
		}))
	}))
	defer agent.Close()
	t.Setenv("AI_TRAVEL_AGENT_URL", agent.URL)

	srv := khttp.NewServer()
	registerMobileAIDispatchRoutesWithDeps(srv, &fakeMobileOrderService{}, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/ai/travel-advice", `{
		"orderId":"5001",
		"startAddress":"Shanghai Station",
		"endAddress":"Hongqiao Airport",
		"driverLat":31.2304,
		"driverLng":121.4737,
		"scene":"before_departure"
	}`, 2001)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "5001", forwarded["orderId"])
	require.Equal(t, "2001", forwarded["driverId"])
	require.Equal(t, "driver", forwarded["userRole"])
	require.Equal(t, "Shanghai Station", forwarded["startAddress"])
	require.Equal(t, "Hongqiao Airport", forwarded["endAddress"])

	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	require.Equal(t, "medium", data["riskLevel"])
	require.Equal(t, "AI advice ready", data["displayText"])
}

func TestDriverNearbyTravelAdviceDoesNotRequireOrderID(t *testing.T) {
	var forwarded map[string]any
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/travel-advice", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&forwarded))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"mode":           "nearby",
			"riskLevel":      "low",
			"summary":        "周边 5 公里路况可接单。",
			"trafficAdvice":  []string{"周边道路整体畅通"},
			"dispatchAdvice": []string{"可继续接单"},
			"displayText":    "AI预警已生成",
		}))
	}))
	defer agent.Close()
	t.Setenv("AI_TRAVEL_AGENT_URL", agent.URL)

	srv := khttp.NewServer()
	registerMobileAIDispatchRoutesWithDeps(srv, nil, nil, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/ai/travel-advice", `{
		"mode":"nearby",
		"driverLat":31.2304,
		"driverLng":121.4737,
		"scene":"idle_warning"
	}`, 2001)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "nearby", forwarded["mode"])
	require.Equal(t, "2001", forwarded["driverId"])
	require.Equal(t, float64(31.2304), forwarded["driverLat"])
	require.Equal(t, float64(121.4737), forwarded["driverLng"])

	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	require.Equal(t, "nearby", data["mode"])
	require.Equal(t, "AI预警已生成", data["displayText"])
}

func TestGatewayTravelRouteInfoIsPublicForCozeCallback(t *testing.T) {
	srv := khttp.NewServer()
	registerMobileAIDispatchRoutesWithDeps(srv, nil, nil, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSON(srv, http.MethodPost, "/api/travel/route-info", `{
		"origin":"上海静安寺",
		"destination":"虹桥火车站",
		"city":"上海",
		"weather":"暴雨黄色预警",
		"userRole":"passenger"
	}`)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	require.Equal(t, "上海静安寺", data["origin"])
	require.Equal(t, "虹桥火车站", data["destination"])
	require.Equal(t, "上海", data["city"])
	require.Equal(t, true, data["available"])
	require.Equal(t, "gateway-service", data["source"])
}

func TestMobileTrackRoutesStoreAndReplayDriverLocation(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/location/report", `{
		"driverId":2001,
		"orderId":"3001",
		"city":"上海",
		"lat":31.2304,
		"lng":121.4737,
		"reportedAt":"2026-07-31T10:00:00+08:00"
	}`, 2001)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, float64(0), decodeGatewayBody(t, res)["code"])
	require.Equal(t, int64(2001), driverSvc.reportReq.DriverId)
	require.Equal(t, int64(3001), driverSvc.reportReq.OrderId)
	require.Equal(t, float64(31.2304), driverSvc.reportReq.Latitude)

	res = doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/track/replay?driverId=2001&orderId=3001&page=1&pageSize=20", "", 2001)
	require.Equal(t, http.StatusOK, res.Code)
	replay := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), replay["code"])
	replayData := replay["data"].(map[string]any)
	require.Equal(t, float64(3), replayData["total"])
	require.Equal(t, int64(2001), driverSvc.replayReq.DriverId)
	require.Equal(t, int64(3001), driverSvc.replayReq.OrderId)
	require.Equal(t, int32(20), driverSvc.replayReq.PageSize)

	res = doGatewayJSON(srv, http.MethodGet, "/api/v1/passenger/orders/3001/track?driverId=2001", "")
	require.Equal(t, http.StatusOK, res.Code)
	track := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), track["code"])
	trackData := track["data"].(map[string]any)
	require.Equal(t, "3001", trackData["orderId"])
	require.Equal(t, float64(3), trackData["total"])
}

func TestMobileDriverLocationReportRequiresOrderID(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/location/report", `{
		"lat":31.2304,
		"lng":121.4737
	}`, 2001)

	require.Equal(t, http.StatusBadRequest, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(400), payload["code"])
	require.Equal(t, "orderId is required", payload["msg"])
	require.Nil(t, driverSvc.reportReq)
}

func TestMobileDriverLocationReportAcceptsLatitudeLongitudeAliases(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/location/report", `{
		"orderId":"3001",
		"latitude":34.021714619798054,
		"longitude":118.28257515301729,
		"reportedAt":"2026-08-11T10:00:00+08:00"
	}`, 2001)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, float64(0), decodeGatewayBody(t, res)["code"])
	require.Equal(t, int64(2001), driverSvc.reportReq.DriverId)
	require.Equal(t, int64(3001), driverSvc.reportReq.OrderId)
	require.Equal(t, float64(34.021714619798054), driverSvc.reportReq.Latitude)
	require.Equal(t, float64(118.28257515301729), driverSvc.reportReq.Longitude)
}

func TestPassengerOrderTrackUsesOrderDetailDriverAndOrderID(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{detailDriverID: 2001}
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/passenger/orders/3001/track?page=1&pageSize=20", "", 3001)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	require.Equal(t, int64(3001), orderSvc.detailReqPassengerID)
	require.Equal(t, int64(2001), driverSvc.replayReq.DriverId)
	require.Equal(t, int64(3001), driverSvc.replayReq.OrderId)
	require.Equal(t, int32(20), driverSvc.replayReq.PageSize)
}

func TestPassengerOrderTrackFallsBackToQueryDriverIDWhenDetailMissingDriver(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{detailDriverID: -1}
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/passenger/orders/3001/track?driverId=2001&page=1&pageSize=20", "", 3001)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	require.Equal(t, int64(3001), orderSvc.detailReqPassengerID)
	require.Equal(t, int64(2001), driverSvc.replayReq.DriverId)
	require.Equal(t, int64(3001), driverSvc.replayReq.OrderId)
	require.Equal(t, "2001", payload["data"].(map[string]any)["driverId"])
}

func TestMobileDriverTrackReplayRequiresOrderID(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/track/replay?page=1&pageSize=20", "", 2001)

	require.Equal(t, http.StatusBadRequest, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(400), payload["code"])
	require.Equal(t, "orderId is required", payload["msg"])
	require.Nil(t, driverSvc.replayReq)
}

func TestMobileDriverTrackRoutesUseCurrentUserInsteadOfRequestDriverID(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/location/report", `{
		"driverId":9999,
		"orderId":"3001",
		"city":"上海",
		"lat":31.2304,
		"lng":121.4737
	}`, 2001)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, int64(2001), driverSvc.reportReq.DriverId)

	res = doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/track/replay?driverId=9999&orderId=3001&page=1&pageSize=20", "", 2001)
	require.Equal(t, http.StatusOK, res.Code)
	replay := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), replay["code"])
	data := replay["data"].(map[string]any)
	require.Equal(t, "2001", data["driverId"])
	require.Equal(t, float64(3), data["total"])
	points := data["points"].([]any)
	require.Equal(t, "2001", points[0].(map[string]any)["driverId"])
	require.Equal(t, int64(2001), driverSvc.replayReq.DriverId)
}

func TestMobileTrackReplayReturnsFilteredTotalNotPageSize(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, nil, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	for i := 0; i < 3; i++ {
		res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/location/report", `{
			"driverId":2001,
			"orderId":"3001",
			"city":"上海",
			"lat":31.2304,
			"lng":121.4737
		}`, 2001)
		require.Equal(t, http.StatusOK, res.Code)
	}

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/track/replay?orderId=3001&page=2&pageSize=2", "", 2001)
	require.Equal(t, http.StatusOK, res.Code)
	data := decodeGatewayBody(t, res)["data"].(map[string]any)
	require.Equal(t, float64(3), data["total"])
	require.Len(t, data["points"].([]any), 1)
	require.Equal(t, int32(2), driverSvc.replayReq.Page)
	require.Equal(t, int32(2), driverSvc.replayReq.PageSize)
}

func TestMobileDriverOrderRoutesUseCurrentDriver(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{}
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/orders/5001", "", 2001)
	require.Equal(t, http.StatusOK, res.Code)
	detail := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), detail["code"])
	require.Equal(t, int64(2001), orderSvc.detailReqPassengerID)

	res = doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/orders/5001/reject", `{"reject_reason":"vehicle fault"}`, 2001)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, int64(2001), orderSvc.rejectReqDriverID)
	require.Equal(t, "vehicle fault", orderSvc.rejectReason)

	res = doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/orders?page=1&pageSize=20", "", 2001)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, int64(2001), orderSvc.listDriverOrdersReqDriverID)

	res = doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/stats", "", 2001)
	require.Equal(t, http.StatusOK, res.Code)
	stats := decodeGatewayBody(t, res)["data"].(map[string]any)
	require.Equal(t, float64(1), stats["pending"])

	res = doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/orders/5001/complete", `{}`, 2001)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, float64(0), decodeGatewayBody(t, res)["code"])
	require.Equal(t, int64(2001), orderSvc.completeReqDriverID)
}

func TestMobileDriverOrdersReturnsAcceptedOrdersForLocationSync(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/orders?status=accepted&page=1&pageSize=20", "", 2001)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	items := data["items"].([]any)
	require.Len(t, items, 1)
	order := items[0].(map[string]any)
	require.Equal(t, "5001", order["id"])
	require.Equal(t, "accepted", order["status"])
	require.Equal(t, "Shanghai Station", order["origin"])
	require.Equal(t, "Hongqiao Airport", order["destination"])
	require.Equal(t, int64(2001), orderSvc.listDriverOrdersReqDriverID)
	require.Equal(t, int32(1), orderSvc.listDriverOrdersReqStatus)
	require.Equal(t, int64(0), orderSvc.pendingReqDriverID)
}

func TestMobileDriverStatsUsesIncomeService(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{incomeReply: &orderv1.DriverIncomeReply{
		TodayOrders:     2,
		TodayIncome:     88.8,
		PendingWithdraw: 88.8,
	}}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/stats", "", 2001)

	require.Equal(t, http.StatusOK, res.Code)
	stats := decodeGatewayBody(t, res)["data"].(map[string]any)
	require.Equal(t, float64(2), stats["todayOrders"])
	require.Equal(t, float64(88.8), stats["todayIncome"])
	require.Equal(t, float64(88.8), stats["pendingWithdraw"])
	require.Equal(t, float64(1), stats["pending"])
}

func TestMobileDriverAvailableOrdersReturnRouteFieldsFromDetail(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{detailDriverID: 2001}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/orders/available?page=1&pageSize=20", "", 2001)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	items := data["items"].([]any)
	require.Len(t, items, 1)
	item := items[0].(map[string]any)
	require.Equal(t, "5001", item["id"])
	require.Equal(t, "2001", item["driverId"])
	require.Equal(t, "Shanghai Station", item["origin"])
	require.Equal(t, "Hongqiao Airport", item["destination"])
	require.Equal(t, "2026-08-11T09:30:00+08:00", item["departTime"])
	require.Equal(t, int64(2001), orderSvc.detailReqPassengerID)
}

func TestMobileDriverLifecycleRoutesForwardIdempotencyKey(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUserAndHeader(srv, http.MethodPost, "/api/v1/driver/orders/5001/accept", `{"idempotency_key":"body-accept"}`, 2001, "header-accept")
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "header-accept", orderSvc.lastIdempotencyKey)

	res = doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/orders/5001/start-pickup", `{"idempotency_key":"body-pickup"}`, 2001)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, int64(2001), orderSvc.startPickupReqDriverID)
	require.Equal(t, "body-pickup", orderSvc.lastIdempotencyKey)

	res = doGatewayJSONWithUserAndHeader(srv, http.MethodPost, "/api/v1/driver/orders/5001/start-delivery", `{}`, 2001, "header-delivery")
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, int64(2001), orderSvc.startDeliveryReqDriverID)
	require.Equal(t, "header-delivery", orderSvc.lastIdempotencyKey)

	res = doGatewayJSONWithUserAndHeader(srv, http.MethodPost, "/api/v1/driver/orders/5001/complete", `{"idempotency_key":"body-complete"}`, 2001, "header-complete")
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, int64(2001), orderSvc.completeReqDriverID)
	require.Equal(t, "header-complete", orderSvc.lastIdempotencyKey)
}

func TestMobileDriverOrderRoutesRejectInvalidOrderID(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/orders/not-a-number", "", 2001)

	require.Equal(t, http.StatusBadRequest, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(400), payload["code"])
	require.Equal(t, "订单ID格式不正确，请刷新后重试", payload["msg"])
}

func TestMobileDriverOrderRoutesReturnChineseJSONForBusinessError(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{acceptErr: status.Error(codes.PermissionDenied, "订单已被其他司机处理")}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodPost, "/api/v1/driver/orders/5001/accept", `{}`, 2001)

	require.Equal(t, http.StatusForbidden, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(403), payload["code"])
	require.Equal(t, "订单已被其他司机处理", payload["msg"])
}

func TestMobileOrderRoutesReturnFrontendSafeOrderDTO(t *testing.T) {
	srv := khttp.NewServer()
	orderSvc := &fakeMobileOrderService{}
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, &fakeMobileDriverTrackingService{}, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/orders/5001", "", 2001)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	order := payload["data"].(map[string]any)["order"].(map[string]any)
	require.Equal(t, "5001", order["id"])
	require.Equal(t, "accepted", order["status"])
	require.Equal(t, float64(1), order["rawStatus"])
}

func TestMobileDriverLocationHistoryAliasesTrackReplay(t *testing.T) {
	srv := khttp.NewServer()
	driverSvc := &fakeMobileDriverTrackingService{}
	registerMobileAIDispatchRoutesWithDeps(srv, &fakeMobileOrderService{}, driverSvc, &fakeMobileAIClient{}, newMobileTrackStore())

	res := doGatewayJSONWithUser(srv, http.MethodGet, "/api/v1/driver/location/history?orderId=5001&page=1&pageSize=10", "", 2001)

	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	require.Equal(t, int64(2001), driverSvc.replayReq.DriverId)
	require.Equal(t, int64(5001), driverSvc.replayReq.OrderId)
	require.Equal(t, int32(10), driverSvc.replayReq.PageSize)
}

func doGatewayJSON(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	return doGatewayJSONWithUser(handler, method, path, body, 0)
}

func doGatewayJSONWithUser(handler http.Handler, method, path, body string, userID int64) *httptest.ResponseRecorder {
	return doGatewayJSONWithUserAndHeader(handler, method, path, body, userID, "")
}

func doGatewayJSONWithUserAndHeader(handler http.Handler, method, path, body string, userID int64, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if userID > 0 {
		req.Header.Set("X-User-Id", strconv.FormatInt(userID, 10))
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func decodeGatewayBody(t *testing.T, res *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	return payload
}
