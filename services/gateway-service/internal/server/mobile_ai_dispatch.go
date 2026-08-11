package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"

	"ride-hailing/pkg/cozex"
	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/gateway-service/internal/service"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

const (
	cozeTravelBotTokenEnv         = "COZE_TRAVEL_BOT_TOKEN"
	cozeRainRouteWorkflowTokenEnv = "COZE_RAIN_ROUTE_WORKFLOW_TOKEN"
	cozeTravelBotURLEnv           = "COZE_TRAVEL_BOT_URL"
	cozeRainRouteWorkflowURLEnv   = "COZE_RAIN_ROUTE_WORKFLOW_URL"
	cozeTravelBotProjectIDEnv     = "COZE_TRAVEL_BOT_PROJECT_ID"
	cozeTravelBotSessionIDEnv     = "COZE_TRAVEL_BOT_SESSION_ID"
	aiTravelAgentURLEnv           = "AI_TRAVEL_AGENT_URL"
)

type mobileAIClient interface {
	CallTravelBot(ctx context.Context, req cozex.TravelBotRequest) (*cozex.TravelBotResponse, error)
	CallRainRouteWorkflow(ctx context.Context, req cozex.RainRouteWorkflowRequest) (*cozex.RainRouteWorkflowResponse, error)
}

type mobileOrderService interface {
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
}

type mobileDriverTrackingService interface {
	ReportDriverLocation(ctx context.Context, req *driverv1.ReportDriverLocationRequest) (*driverv1.DriverLocationReply, error)
	ReplayDriverTrack(ctx context.Context, req *driverv1.ReplayDriverTrackRequest) (*driverv1.ReplayDriverTrackReply, error)
}

type mobileAIChatRequest struct {
	SessionID string `json:"sessionId"`
	Text      string `json:"text"`
	UserID    int64  `json:"userId"`
	UserRole  string `json:"userRole"`
}

type mobileRainRouteRequest struct {
	SessionID   string `json:"sessionId"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	City        string `json:"city"`
	Weather     string `json:"weather"`
	Avoid       string `json:"avoid"`
	Preference  string `json:"preference"`
	UserRole    string `json:"userRole"`
}

type mobileChatWithRouteRequest struct {
	Chat  mobileAIChatRequest    `json:"chat"`
	Route mobileRainRouteRequest `json:"route"`
}

type mobileFloodReportRequest struct {
	ReporterID   int64   `json:"reporterId"`
	ReporterRole string  `json:"reporterRole"`
	City         string  `json:"city"`
	LocationText string  `json:"locationText"`
	PhotoURL     string  `json:"photoUrl"`
	DepthCM      float64 `json:"depthCm"`
	Confidence   float64 `json:"confidence"`
}

type mobileLocationReportRequest struct {
	DriverID   int64   `json:"driverId"`
	OrderID    any     `json:"orderId"`
	City       string  `json:"city"`
	FleetID    string  `json:"fleetId"`
	HotZone    string  `json:"hotZone"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Speed      float64 `json:"speed"`
	Heading    float64 `json:"heading"`
	ReportedAt string  `json:"reportedAt"`
}

type mobileDriverTravelAdviceRequest struct {
	Mode             string  `json:"mode"`
	OrderID          any     `json:"orderId"`
	OrderStatus      string  `json:"orderStatus"`
	StartAddress     string  `json:"startAddress"`
	EndAddress       string  `json:"endAddress"`
	DriverLat        float64 `json:"driverLat"`
	DriverLng        float64 `json:"driverLng"`
	RouteDistanceKm  float64 `json:"routeDistanceKm"`
	EstimatedMinutes float64 `json:"estimatedMinutes"`
	WeatherText      string  `json:"weatherText"`
	Scene            string  `json:"scene"`
}

func (r *mobileLocationReportRequest) normalizedCoordinates() (float64, float64, bool) {
	lat := r.Lat
	if lat == 0 {
		lat = r.Latitude
	}
	lng := r.Lng
	if lng == 0 {
		lng = r.Longitude
	}
	valid := !math.IsNaN(lat) &&
		!math.IsNaN(lng) &&
		!math.IsInf(lat, 0) &&
		!math.IsInf(lng, 0) &&
		lat >= -90 &&
		lat <= 90 &&
		lng >= -180 &&
		lng <= 180
	return lat, lng, valid
}

type mobileFloodReport struct {
	ReportNo    string    `json:"reportNo"`
	ReporterID  int64     `json:"reporterId"`
	UserRole    string    `json:"userRole"`
	City        string    `json:"city"`
	Location    string    `json:"locationText"`
	PhotoURL    string    `json:"photoUrl"`
	DepthCM     float64   `json:"depthCm"`
	Confidence  float64   `json:"confidence"`
	AuditStatus string    `json:"auditStatus"`
	CreatedAt   time.Time `json:"createdAt"`
}

type mobileTrackStore struct {
	mu      sync.RWMutex
	floods  []mobileFloodReport
	nowFunc func() time.Time
}

func registerMobileAIDispatchRoutes(srv *khttp.Server, orderSvc *service.OrderService, driverSvc *service.DriverService, passengerSvc mobilePassengerProfileService) {
	registerMobileAIDispatchRoutesWithDeps(srv, orderSvc, driverSvc, newMobileCozeClient(), newMobileTrackStore(), mobileOrderContactProfiles{
		passenger: passengerSvc,
		driver:    driverSvc,
	})
}

type mobileOrderContactProfiles struct {
	passenger mobilePassengerProfileService
	driver    mobileDriverProfileService
}

func registerMobileAIDispatchRoutesWithDeps(srv *khttp.Server, orderSvc mobileOrderService, driverSvc mobileDriverTrackingService, aiClient mobileAIClient, trackStore *mobileTrackStore, profiles ...mobileOrderContactProfiles) {
	if trackStore == nil {
		trackStore = newMobileTrackStore()
	}
	contactProfiles := mobileOrderContactProfiles{}
	if len(profiles) > 0 {
		contactProfiles = profiles[0]
	}
	router := srv.Route("/")
	router.POST("/api/travel/route-info", func(ctx khttp.Context) error {
		req := new(mobileRainRouteRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if strings.TrimSpace(req.UserRole) == "" {
			req.UserRole = "passenger"
		}
		if strings.TrimSpace(req.Origin) == "" || strings.TrimSpace(req.Destination) == "" || strings.TrimSpace(req.City) == "" {
			return errors.New("origin, destination and city are required")
		}
		return returnData(ctx, map[string]any{
			"origin":      strings.TrimSpace(req.Origin),
			"destination": strings.TrimSpace(req.Destination),
			"city":        strings.TrimSpace(req.City),
			"weather":     strings.TrimSpace(req.Weather),
			"available":   true,
			"source":      "gateway-service",
			"riskHints": []string{
				"avoid flooded roads when heavy rain is reported",
				"prefer main roads with recent traffic updates",
			},
		}, nil)
	})
	router.POST("/api/v1/ai/chat", func(ctx khttp.Context) error {
		req := new(mobileAIChatRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if req.UserID == 0 {
			req.UserID = UserIDFromRequest(ctx.Request())
		}
		if err := validateMobileAIChat(*req); err != nil {
			return err
		}
		res, err := aiClient.CallTravelBot(ctx, cozex.TravelBotRequest{
			Text:      strings.TrimSpace(req.Text),
			SessionID: defaultMobileSessionID(req.SessionID, req.UserID),
		})
		if err != nil {
			return err
		}
		return returnData(ctx, map[string]any{
			"sessionId": defaultMobileSessionID(req.SessionID, req.UserID),
			"userId":    req.UserID,
			"userRole":  normalizeMobileRole(req.UserRole),
			"answer":    res.RawBody,
		}, nil)
	})
	router.POST("/api/v1/ai/rain-route", func(ctx khttp.Context) error {
		req := new(mobileRainRouteRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if err := validateMobileRainRoute(*req); err != nil {
			return err
		}
		res, err := aiClient.CallRainRouteWorkflow(ctx, toCozeRainRoute(*req))
		if err != nil {
			return err
		}
		return returnData(ctx, map[string]any{
			"sessionId": defaultMobileSessionID(req.SessionID, UserIDFromRequest(ctx.Request())),
			"rawResult": res.RawBody,
			"city":      strings.TrimSpace(req.City),
		}, nil)
	})
	router.POST("/api/v1/ai/chat-with-route", func(ctx khttp.Context) error {
		req := new(mobileChatWithRouteRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if err := validateMobileRainRoute(req.Route); err != nil {
			return err
		}
		routeRes, err := aiClient.CallRainRouteWorkflow(ctx, toCozeRainRoute(req.Route))
		if err != nil {
			return err
		}
		chat := req.Chat
		if strings.TrimSpace(chat.SessionID) == "" {
			chat.SessionID = req.Route.SessionID
		}
		if chat.UserID == 0 {
			chat.UserID = UserIDFromRequest(ctx.Request())
		}
		if strings.TrimSpace(chat.UserRole) == "" {
			chat.UserRole = req.Route.UserRole
		}
		chat.Text = strings.TrimSpace(chat.Text)
		if chat.Text == "" {
			chat.Text = "请结合雨天路线结果给出出行建议"
		}
		chat.Text += "\nRoute context: " + routeRes.RawBody
		if err := validateMobileAIChat(chat); err != nil {
			return err
		}
		chatRes, err := aiClient.CallTravelBot(ctx, cozex.TravelBotRequest{
			Text:      chat.Text,
			SessionID: defaultMobileSessionID(chat.SessionID, chat.UserID),
		})
		if err != nil {
			return err
		}
		return returnData(ctx, map[string]any{
			"route": map[string]any{"rawResult": routeRes.RawBody},
			"chat":  map[string]any{"answer": chatRes.RawBody, "sessionId": defaultMobileSessionID(chat.SessionID, chat.UserID)},
		}, nil)
	})
	router.POST("/api/v1/ai/flood-report", func(ctx khttp.Context) error {
		req := new(mobileFloodReportRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if req.ReporterID == 0 {
			req.ReporterID = UserIDFromRequest(ctx.Request())
		}
		report, err := trackStore.addFloodReport(*req)
		return returnData(ctx, report, err)
	})
	router.GET("/api/v1/passenger/orders", func(ctx khttp.Context) error {
		query := ctx.Query()
		reply, err := orderSvc.ListOrders(ctx, currentUserID(ctx.Request()), parseMobileOrderStatus(query.Get("status")), int32(parseInt(query.Get("page"))), int32(parseInt(query.Get("pageSize"))))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, mobileOrderListResponse(reply), nil)
	})
	router.GET("/api/v1/passenger/orders/{id}/track", func(ctx khttp.Context) error {
		query := ctx.Query()
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID, err := parseMobileInt64Value(query.Get("driverId"))
		if err != nil {
			return returnBadRequest(ctx, "driverId is invalid")
		}
		if orderSvc != nil {
			detail, err := orderSvc.GetOrderDetail(ctx, orderID, UserIDFromRequest(ctx.Request()))
			if err != nil {
				return returnData(ctx, nil, err)
			}
			if detail.GetOrder().GetDriverId() > 0 {
				driverID = detail.GetOrder().GetDriverId()
			}
		}
		if driverSvc == nil {
			return returnUnavailable(ctx, "driver tracking service unavailable")
		}
		reply, err := driverSvc.ReplayDriverTrack(ctx, &driverv1.ReplayDriverTrackRequest{
			DriverId: driverID,
			OrderId:  orderID,
			Page:     int32(parseMobilePage(query.Get("page"))),
			PageSize: int32(parseMobilePageSize(query.Get("pageSize"))),
		})
		if err != nil {
			return err
		}
		points := mobileTrackPointsFromProto(reply.GetItems())
		zap.L().Info("passenger order track replay",
			zap.Int64("order_id", orderID),
			zap.Int64("driver_id", driverID),
			zap.Int("points", len(points)),
			zap.Int64("user_id", UserIDFromRequest(ctx.Request())),
		)
		return returnData(ctx, map[string]any{
			"orderId":   int64String(orderID),
			"order_id":  int64String(orderID),
			"driverId":  int64String(driverID),
			"driver_id": int64String(driverID),
			"points":    points,
			"list":      points,
			"total":     reply.GetTotal(),
		}, nil)
	})
	router.GET("/api/v1/passenger/orders/{id}", func(ctx khttp.Context) error {
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		reply, err := orderSvc.GetOrderDetail(ctx, orderID, currentUserID(ctx.Request()))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		payload := mobileOrderDetailResponse(reply)
		enrichMobileOrderContactPayload(ctx, payload, contactProfiles.passenger, contactProfiles.driver)
		return returnData(ctx, payload, nil)
	})
	router.GET("/api/v1/driver/ai-alerts", func(ctx khttp.Context) error {
		return returnData(ctx, map[string]any{"list": trackStore.listFloodAlerts(), "total": len(trackStore.listFloodAlerts())}, nil)
	})
	router.POST("/api/v1/driver/ai/travel-advice", func(ctx khttp.Context) error {
		req := new(mobileDriverTravelAdviceRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		mode := normalizeMobileTravelAdviceMode(req.Mode)
		var orderID int64
		startAddress := strings.TrimSpace(req.StartAddress)
		endAddress := strings.TrimSpace(req.EndAddress)
		if mode == "order" {
			var err error
			orderID, err = parseMobileInt64Value(req.OrderID)
			if err != nil {
				return returnBadRequest(ctx, "orderId is invalid")
			}
			if orderID <= 0 {
				return returnBadRequest(ctx, "orderId is required")
			}
			if orderSvc != nil {
				detail, err := orderSvc.GetOrderDetail(ctx, orderID, driverID)
				if err != nil {
					return returnData(ctx, nil, err)
				}
				if order := detail.GetOrder(); order != nil {
					if order.GetDriverId() > 0 && order.GetDriverId() != driverID {
						return returnBadRequest(ctx, "order does not belong to current driver")
					}
					if startAddress == "" {
						startAddress = order.GetOrigin()
					}
					if endAddress == "" {
						endAddress = order.GetDestination()
					}
					if strings.TrimSpace(req.OrderStatus) == "" {
						req.OrderStatus = mobileOrderStatus(order.GetStatus())
					}
				}
			}
			if startAddress == "" || endAddress == "" {
				return returnBadRequest(ctx, "startAddress and endAddress are required")
			}
		} else if !validMobileCoordinates(req.DriverLat, req.DriverLng) {
			return returnBadRequest(ctx, "driver location is required for nearby warning")
		}
		scene := strings.TrimSpace(req.Scene)
		if scene == "" {
			if mode == "nearby" {
				scene = "idle_warning"
			} else {
				scene = "before_departure"
			}
		}
		advice, err := callMobileTravelAgent(ctx, map[string]any{
			"mode":             mode,
			"orderId":          int64String(orderID),
			"driverId":         int64String(driverID),
			"orderStatus":      strings.TrimSpace(req.OrderStatus),
			"startAddress":     startAddress,
			"endAddress":       endAddress,
			"driverLat":        req.DriverLat,
			"driverLng":        req.DriverLng,
			"routeDistanceKm":  req.RouteDistanceKm,
			"estimatedMinutes": req.EstimatedMinutes,
			"weatherText":      strings.TrimSpace(req.WeatherText),
			"scene":            scene,
			"userRole":         "driver",
		})
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, advice, nil)
	})
	router.GET("/api/v1/driver/stats", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		reply, err := orderSvc.PendingOrders(ctx, driverID, 0, 1, 1)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		start, end := mobileTodayRange()
		income, err := orderSvc.GetDriverIncome(ctx, driverID, start, end, 1, 100)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, map[string]any{
			"todayOrders":      income.GetTodayOrders(),
			"today_orders":     income.GetTodayOrders(),
			"todayIncome":      income.GetTodayIncome(),
			"today_income":     income.GetTodayIncome(),
			"pendingWithdraw":  income.GetPendingWithdraw(),
			"pending_withdraw": income.GetPendingWithdraw(),
			"pending":          reply.GetTotal(),
		}, nil)
	})
	router.GET("/api/v1/driver/income", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		query := ctx.Query()
		start, end := mobileTodayRange()
		if value := strings.TrimSpace(query.Get("start_time")); value != "" {
			start = value
		}
		if value := strings.TrimSpace(query.Get("end_time")); value != "" {
			end = value
		}
		reply, err := orderSvc.GetDriverIncome(ctx, driverID, start, end, int32(parseMobilePage(query.Get("page"))), int32(parseMobilePageSize(query.Get("pageSize"))))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, mobileDriverIncomeResponse(reply), nil)
	})
	router.GET("/api/v1/driver/orders/available", func(ctx khttp.Context) error {
		query := ctx.Query()
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		reply, err := orderSvc.PendingOrders(ctx, driverID, int64(parseInt(query.Get("tripId"))), int32(parseMobilePage(query.Get("page"))), int32(parseMobilePageSize(query.Get("pageSize"))))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, mobilePendingOrderListResponseWithDetails(ctx, orderSvc, driverID, reply), nil)
	})
	router.GET("/api/v1/driver/orders", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		query := ctx.Query()
		reply, err := orderSvc.ListDriverOrders(ctx, driverID, parseMobileOrderStatus(query.Get("status")), int32(parseMobilePage(query.Get("page"))), int32(parseMobilePageSize(query.Get("pageSize"))))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, mobileOrderListResponse(reply), nil)
	})
	router.GET("/api/v1/driver/orders/{id}", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		reply, err := orderSvc.GetOrderDetail(ctx, orderID, driverID)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		payload := mobileOrderDetailResponse(reply)
		enrichMobileOrderContactPayload(ctx, payload, contactProfiles.passenger, contactProfiles.driver)
		return returnData(ctx, payload, nil)
	})
	router.POST("/api/v1/driver/orders/{id}/accept", func(ctx khttp.Context) error {
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID := currentUserID(ctx.Request())
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.AcceptOrder(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "mobile_accept_order", driverID, orderID, idempotencyKey, "paid", "accepted", err)
		return returnMessage(ctx, "accepted", err)
	})
	router.POST("/api/v1/driver/orders/{id}/reject", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		actionReq := orderActionFromRequest(ctx)
		err = orderSvc.RejectOrder(ctx, orderID, driverID, actionReq.idempotencyKey(), actionReq.rejectReason())
		logGatewayOrderAction(ctx, "mobile_reject_order", driverID, orderID, actionReq.idempotencyKey(), "paid", "rejected", err)
		return returnMessage(ctx, "rejected", err)
	})
	router.POST("/api/v1/driver/orders/{id}/start-pickup", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.StartPickup(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "mobile_start_pickup", driverID, orderID, idempotencyKey, "accepted", "picking_up", err)
		return returnMessage(ctx, "picking_up", err)
	})
	router.POST("/api/v1/driver/orders/{id}/start-delivery", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.StartDelivery(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "mobile_start_delivery", driverID, orderID, idempotencyKey, "picking_up", "delivering", err)
		return returnMessage(ctx, "delivering", err)
	})
	router.POST("/api/v1/driver/orders/{id}/complete", func(ctx khttp.Context) error {
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if orderSvc == nil {
			return returnUnavailable(ctx, "order service unavailable")
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.CompleteOrder(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "mobile_complete_order", driverID, orderID, idempotencyKey, "delivering", "completed", err)
		return returnMessage(ctx, "completed", err)
	})
	router.POST("/api/v1/driver/location/report", func(ctx khttp.Context) error {
		req := new(mobileLocationReportRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		req.DriverID = UserIDFromRequest(ctx.Request())
		if req.DriverID <= 0 {
			return errors.New("current driver user is required")
		}
		if driverSvc == nil {
			return returnUnavailable(ctx, "driver tracking service unavailable")
		}
		orderID, err := parseMobileInt64Value(req.OrderID)
		if err != nil {
			return returnBadRequest(ctx, "orderId is invalid")
		}
		if orderID <= 0 {
			return returnBadRequest(ctx, "orderId is required")
		}
		lat, lng, ok := req.normalizedCoordinates()
		if !ok {
			return returnBadRequest(ctx, "location is invalid")
		}
		reply, err := driverSvc.ReportDriverLocation(ctx, &driverv1.ReportDriverLocationRequest{
			DriverId:   req.DriverID,
			OrderId:    orderID,
			Latitude:   lat,
			Longitude:  lng,
			Speed:      req.Speed,
			Heading:    req.Heading,
			ReportedAt: strings.TrimSpace(req.ReportedAt),
		})
		if err != nil {
			return err
		}
		zap.L().Info("driver location reported",
			gatewayLogFields(ctx.Request(), zap.Int64("user_id", req.DriverID), zap.Int64("driver_id", req.DriverID), zap.Int64("order_id", orderID), zap.Float64("latitude", lat), zap.Float64("longitude", lng), zap.String("status_before", "picking_up|delivering"), zap.String("status_after", "location_reported"))...)
		return returnData(ctx, mobileTrackPointFromProto(reply.GetLocation()), nil)
	})
	router.GET("/api/v1/driver/track/replay", func(ctx khttp.Context) error {
		query := ctx.Query()
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		orderID, err := parseMobileInt64Value(query.Get("orderId"))
		if err != nil {
			return returnBadRequest(ctx, "orderId is invalid")
		}
		if orderID <= 0 {
			return returnBadRequest(ctx, "orderId is required")
		}
		if driverSvc == nil {
			return returnUnavailable(ctx, "driver tracking service unavailable")
		}
		reply, err := driverSvc.ReplayDriverTrack(ctx, &driverv1.ReplayDriverTrackRequest{
			DriverId: driverID,
			OrderId:  orderID,
			Page:     int32(parseMobilePage(query.Get("page"))),
			PageSize: int32(parseMobilePageSize(query.Get("pageSize"))),
		})
		if err != nil {
			return err
		}
		points := mobileTrackPointsFromProto(reply.GetItems())
		zap.L().Info("driver track replay",
			zap.Int64("driver_id", driverID),
			zap.Int64("order_id", orderID),
			zap.Int("points", len(points)),
		)
		return returnData(ctx, map[string]any{
			"driverId":  int64String(driverID),
			"driver_id": int64String(driverID),
			"orderId":   int64String(orderID),
			"order_id":  int64String(orderID),
			"points":    points,
			"list":      points,
			"total":     reply.GetTotal(),
		}, nil)
	})
	router.GET("/api/v1/driver/location/history", func(ctx khttp.Context) error {
		query := ctx.Query()
		driverID := UserIDFromRequest(ctx.Request())
		if driverID <= 0 {
			return errors.New("current driver user is required")
		}
		if driverSvc == nil {
			return returnUnavailable(ctx, "driver tracking service unavailable")
		}
		orderID, err := parseMobileInt64Value(query.Get("orderId"))
		if err != nil {
			return returnBadRequest(ctx, "orderId is invalid")
		}
		if orderID <= 0 {
			return returnBadRequest(ctx, "orderId is required")
		}
		reply, err := driverSvc.ReplayDriverTrack(ctx, &driverv1.ReplayDriverTrackRequest{
			DriverId: driverID,
			OrderId:  orderID,
			Page:     int32(parseMobilePage(query.Get("page"))),
			PageSize: int32(parseMobilePageSize(query.Get("pageSize"))),
		})
		if err != nil {
			return err
		}
		points := mobileTrackPointsFromProto(reply.GetItems())
		return returnData(ctx, map[string]any{
			"driverId":  int64String(driverID),
			"driver_id": int64String(driverID),
			"orderId":   int64String(orderID),
			"order_id":  int64String(orderID),
			"points":    points,
			"list":      points,
			"total":     reply.GetTotal(),
		}, nil)
	})
}

func returnUnavailable(ctx khttp.Context, msg string) error {
	return ctx.Returns(map[string]any{"code": http.StatusServiceUnavailable, "data": nil, "msg": msg}, nil)
}

func callMobileTravelAgent(ctx context.Context, payload map[string]any) (map[string]any, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv(aiTravelAgentURLEnv)), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8011"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/travel-advice", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var decoded map[string]any
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("AI travel agent request failed")
	}
	if code, ok := decoded["code"].(float64); ok && code == 0 {
		if data, ok := decoded["data"].(map[string]any); ok {
			return data, nil
		}
	}
	return normalizeMobileTravelAdvice(decoded), nil
}

func normalizeMobileTravelAdvice(payload map[string]any) map[string]any {
	if payload == nil {
		payload = map[string]any{}
	}
	result := map[string]any{
		"riskLevel":      stringValue(payload, "riskLevel", "medium"),
		"summary":        stringValue(payload, "summary", ""),
		"weatherAdvice":  stringSliceValue(payload, "weatherAdvice"),
		"routeAdvice":    stringSliceValue(payload, "routeAdvice"),
		"safetyAdvice":   stringSliceValue(payload, "safetyAdvice"),
		"displayText":    stringValue(payload, "displayText", ""),
		"mode":           stringValue(payload, "mode", "order"),
		"trafficAdvice":  stringSliceValue(payload, "trafficAdvice"),
		"dispatchAdvice": stringSliceValue(payload, "dispatchAdvice"),
		"nearbyTraffic":  payload["nearbyTraffic"],
	}
	if result["summary"] == "" {
		result["summary"] = result["displayText"]
	}
	if result["displayText"] == "" {
		result["displayText"] = result["summary"]
	}
	return result
}

func normalizeMobileTravelAdviceMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), "nearby") {
		return "nearby"
	}
	return "order"
}

func validMobileCoordinates(lat, lng float64) bool {
	return !math.IsNaN(lat) &&
		!math.IsNaN(lng) &&
		!math.IsInf(lat, 0) &&
		!math.IsInf(lng, 0) &&
		lat >= -90 &&
		lat <= 90 &&
		lng >= -180 &&
		lng <= 180 &&
		lat != 0 &&
		lng != 0
}

func stringValue(payload map[string]any, key, fallback string) string {
	if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func stringSliceValue(payload map[string]any, key string) []string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return []string{}
	}
	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		items := make([]string, 0, len(values))
		for _, item := range values {
			if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
				items = append(items, text)
			}
		}
		return items
	case string:
		if text := strings.TrimSpace(values); text != "" {
			return []string{text}
		}
	}
	return []string{}
}

func mobilePendingOrderListResponseWithDetails(ctx context.Context, orderSvc mobileOrderService, driverID int64, reply *orderv1.PendingOrdersReply) map[string]any {
	response := mobilePendingOrderListResponse(reply)
	items, ok := response["items"].([]map[string]any)
	if !ok || orderSvc == nil {
		return response
	}
	for index, item := range reply.GetItems() {
		if item == nil || index >= len(items) {
			continue
		}
		detail, err := orderSvc.GetOrderDetail(ctx, item.GetId(), driverID)
		if err != nil || detail.GetOrder() == nil {
			if err != nil {
				zap.L().Warn("driver pending order detail enrich failed", zap.Int64("driver_id", driverID), zap.Int64("order_id", item.GetId()), zap.Error(err))
			}
			continue
		}
		order := mobileOrderItemResponse(detail.GetOrder())
		for _, key := range []string{"driverId", "driver_id", "origin", "destination", "departTime", "depart_time"} {
			if value, exists := order[key]; exists {
				items[index][key] = value
			}
		}
	}
	response["items"] = items
	response["list"] = items
	return response
}

func newMobileCozeClient() mobileAIClient {
	projectID := cozex.DefaultTravelBotProjectID
	if parsed, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(cozeTravelBotProjectIDEnv)), 10, 64); err == nil && parsed > 0 {
		projectID = parsed
	}
	return cozex.NewClient(cozex.Config{
		TravelBotURL:           os.Getenv(cozeTravelBotURLEnv),
		RainRouteWorkflowURL:   os.Getenv(cozeRainRouteWorkflowURLEnv),
		TravelBotToken:         os.Getenv(cozeTravelBotTokenEnv),
		RainRouteWorkflowToken: os.Getenv(cozeRainRouteWorkflowTokenEnv),
		TravelBotProjectID:     projectID,
		TravelBotSessionID:     os.Getenv(cozeTravelBotSessionIDEnv),
	}, nil)
}

func newMobileTrackStore() *mobileTrackStore {
	return &mobileTrackStore{nowFunc: time.Now}
}

func mobileTrackPointsFromProto(items []*driverv1.DriverLocationPoint) []map[string]any {
	points := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if point := mobileTrackPointFromProto(item); point != nil {
			points = append(points, point)
		}
	}
	return points
}

func mobileTrackPointFromProto(item *driverv1.DriverLocationPoint) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":         int64String(item.GetId()),
		"driverId":   int64String(item.GetDriverId()),
		"driver_id":  int64String(item.GetDriverId()),
		"orderId":    int64String(item.GetOrderId()),
		"order_id":   int64String(item.GetOrderId()),
		"latitude":   item.GetLatitude(),
		"longitude":  item.GetLongitude(),
		"lat":        item.GetLatitude(),
		"lng":        item.GetLongitude(),
		"speed":      item.GetSpeed(),
		"heading":    item.GetHeading(),
		"online":     true,
		"reportedAt": item.GetReportedAt(),
		"createdAt":  item.GetCreatedAt(),
	}
}

func parseMobileInt64Value(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return 0, nil
		}
		return strconv.ParseInt(v, 10, 64)
	case float64:
		if v == 0 {
			return 0, nil
		}
		if math.Trunc(v) != v {
			return 0, errors.New("not an integer")
		}
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, errors.New("unsupported orderId type")
	}
}

func (s *mobileTrackStore) addFloodReport(req mobileFloodReportRequest) (*mobileFloodReport, error) {
	if strings.TrimSpace(req.City) == "" || strings.TrimSpace(req.LocationText) == "" {
		return nil, errors.New("city and locationText are required")
	}
	if req.DepthCM < 0 {
		return nil, errors.New("depthCm must be non-negative")
	}
	if req.Confidence < 0 || req.Confidence > 100 {
		return nil, errors.New("confidence must be between 0 and 100")
	}
	role := normalizeMobileRole(req.ReporterRole)
	if role == "" {
		role = "passenger"
	}
	report := mobileFloodReport{
		ReportNo:    "FLOOD-MOBILE-" + strconv.FormatInt(s.now().UnixNano(), 10),
		ReporterID:  req.ReporterID,
		UserRole:    role,
		City:        strings.TrimSpace(req.City),
		Location:    strings.TrimSpace(req.LocationText),
		PhotoURL:    strings.TrimSpace(req.PhotoURL),
		DepthCM:     req.DepthCM,
		Confidence:  req.Confidence,
		AuditStatus: floodAuditStatus(req.Confidence),
		CreatedAt:   s.now(),
	}
	s.mu.Lock()
	s.floods = append(s.floods, report)
	s.mu.Unlock()
	return &report, nil
}

func (s *mobileTrackStore) listFloodAlerts() []mobileFloodReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]mobileFloodReport{}, s.floods...)
}

func (s *mobileTrackStore) now() time.Time {
	if s == nil || s.nowFunc == nil {
		return time.Now()
	}
	return s.nowFunc()
}

func validateMobileAIChat(req mobileAIChatRequest) error {
	if strings.TrimSpace(req.Text) == "" {
		return errors.New("text is required")
	}
	if normalizeMobileRole(req.UserRole) == "" {
		return errors.New("userRole must be passenger, driver or admin")
	}
	return nil
}

func validateMobileRainRoute(req mobileRainRouteRequest) error {
	if strings.TrimSpace(req.Origin) == "" || strings.TrimSpace(req.Destination) == "" || strings.TrimSpace(req.City) == "" {
		return errors.New("origin, destination and city are required")
	}
	if normalizeMobileRole(req.UserRole) == "" {
		return errors.New("userRole must be passenger, driver or admin")
	}
	return nil
}

func toCozeRainRoute(req mobileRainRouteRequest) cozex.RainRouteWorkflowRequest {
	return cozex.RainRouteWorkflowRequest{
		Origin:      strings.TrimSpace(req.Origin),
		Destination: strings.TrimSpace(req.Destination),
		City:        strings.TrimSpace(req.City),
		Weather:     strings.TrimSpace(req.Weather),
		Avoid:       strings.TrimSpace(req.Avoid),
		Preference:  strings.TrimSpace(req.Preference),
		UserRole:    normalizeMobileRole(req.UserRole),
	}
}

func defaultMobileSessionID(sessionID string, userID int64) string {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		return sessionID
	}
	if userID > 0 {
		return "mobile-" + strconv.FormatInt(userID, 10)
	}
	return "mobile-anonymous"
}

func normalizeMobileRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "passenger", "driver", "admin":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return ""
	}
}

func floodAuditStatus(confidence float64) string {
	if confidence >= 80 {
		return "auto_confirmed"
	}
	return "pending_manual_review"
}

func parseMobilePage(value string) int {
	page := parseInt(value)
	if page <= 0 {
		return 1
	}
	return page
}

func parseMobilePageSize(value string) int {
	pageSize := parseInt(value)
	if pageSize <= 0 {
		return 100
	}
	if pageSize > 500 {
		return 500
	}
	return pageSize
}

func mobileTodayRange() (string, string) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start.Format(time.RFC3339), start.AddDate(0, 0, 1).Format(time.RFC3339)
}
