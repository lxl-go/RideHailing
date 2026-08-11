package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"ride-hailing/pkg/alipayx"
	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
	"ride-hailing/services/gateway-service/internal/conf"
	"ride-hailing/services/gateway-service/internal/data"
	"ride-hailing/services/gateway-service/internal/service"
	orderv1 "ride-hailing/services/order-service/api/order/v1"
	passengerv1 "ride-hailing/services/passenger-service/api/passenger/v1"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, alipayCfg *conf.Alipay, amapCfg *conf.Amap, authSvc *service.AuthService, tripSvc *service.TripService, orderSvc *service.OrderService, reviewSvc *service.ReviewService, passengerSvc *service.PassengerService, driverSvc *service.DriverService) *khttp.Server {
	opts := []khttp.ServerOption{}
	if c != nil && c.Http != nil {
		if c.Http.Addr != "" {
			opts = append(opts, khttp.Address(c.Http.Addr))
		}
		if d, err := time.ParseDuration(c.Http.Timeout); err == nil && d > 0 {
			opts = append(opts, khttp.Timeout(d))
		}
	}
	opts = append(opts, khttp.Filter(NewCORSFilter(), NewAuthFilter(authCfg)))
	srv := khttp.NewServer(opts...)
	permissionChecker := NewCachedPermissionChecker(authSvc, permissionPolicyFromConfig(authCfg))
	registerAuthRoutes(srv, authSvc, passengerSvc, driverSvc)
	registerTripRoutes(srv, permissionChecker, tripSvc)
	registerOrderRoutes(srv, permissionChecker, orderSvc, passengerSvc, driverSvc, alipayCfg)
	registerReviewRoutes(srv, permissionChecker, reviewSvc)
	registerPassengerRoutes(srv, permissionChecker, passengerSvc)
	registerDriverRoutes(srv, permissionChecker, driverSvc)
	registerMobileAIDispatchRoutes(srv, orderSvc, driverSvc, passengerSvc)
	registerOrderChatRoutes(srv, newOrderChatHub(), orderSvc)
	registerMapRoutes(srv, amapCfg)
	registerPaymentRoutes(srv, alipayCfg, orderSvc)
	registerMobileCompatibilityRoutes(srv, tripSvc, reviewSvc)
	return srv
}

func registerAuthRoutes(srv *khttp.Server, authSvc *service.AuthService, passengerSvc *service.PassengerService, driverSvc *service.DriverService) {
	router := srv.Route("/")
	router.POST("/carpool/auth/sms/send", func(ctx khttp.Context) error {
		req := new(authv1.SendLoginCodeRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := authSvc.SendLoginCode(ctx, req)
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/auth/login", func(ctx khttp.Context) error {
		req := new(authv1.LoginRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := authSvc.Login(ctx, req)
		if err == nil && reply != nil && reply.GetUserId() > 0 {
			principal := strings.TrimSpace(req.GetPrincipal())
			switch req.GetRole() {
			case "passenger":
				if _, seedErr := passengerSvc.EnsurePassenger(ctx, reply.GetUserId(), principal); seedErr != nil {
					zap.L().Warn("seed passenger phone failed", zap.Int64("user_id", reply.GetUserId()), zap.Error(seedErr))
				}
			case "driver":
				if _, seedErr := driverSvc.EnsureDriver(ctx, reply.GetUserId(), principal); seedErr != nil {
					zap.L().Warn("seed driver phone failed", zap.Int64("user_id", reply.GetUserId()), zap.Error(seedErr))
				}
			}
		}
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/auth/refresh", func(ctx khttp.Context) error {
		req := new(authv1.RefreshTokenRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := authSvc.RefreshToken(ctx, req.RefreshToken)
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/auth/logout", func(ctx khttp.Context) error {
		req := new(authv1.LogoutRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := authSvc.Logout(ctx, req.RefreshToken)
		return returnData(ctx, reply, err)
	})
}

func registerTripRoutes(srv *khttp.Server, permissions permissionChecker, tripSvc *service.TripService) {
	router := srv.Route("/")
	router.GET("/carpool/trips", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:search") {
			return nil
		}
		query := ctx.Query()
		reply, err := tripSvc.SearchTrips(ctx, data.SearchTripsRequest{
			Origin:      query.Get("origin"),
			Destination: query.Get("destination"),
			DepartDate:  query.Get("depart_date"),
			Page:        int32(parseInt(query.Get("page"))),
			PageSize:    int32(parseInt(query.Get("page_size"))),
		})
		return returnData(ctx, mobileSearchTripListResponse(reply), err)
	})
	router.GET("/carpool/trips/mine", func(ctx khttp.Context) error {
		query := ctx.Query()
		driverID := currentUserID(ctx.Request())
		reply, err := tripSvc.ListDriverTrips(ctx, &tripv1.ListDriverTripsRequest{
			Status:   int32(parseInt(query.Get("status"))),
			Page:     int32(parseInt(query.Get("page"))),
			PageSize: int32(parseInt(query.Get("page_size"))),
			DriverId: driverID,
		})
		return returnData(ctx, mobileDriverTripListResponse(reply), err)
	})
	router.GET("/carpool/trips/{id}", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:view_detail") {
			return nil
		}
		tripID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil || tripID <= 0 {
			return returnBadRequest(ctx, "琛岀▼ID鏍煎紡涓嶆纭紝璇峰埛鏂板悗閲嶈瘯")
		}
		reply, err := tripSvc.GetTripDetail(ctx, tripID)
		return returnData(ctx, mobileTripDetailResponse(reply), err)
	})
	router.POST("/carpool/trips", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:publish") {
			return nil
		}
		req := new(tripv1.PublishTripRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		req.DriverId = currentUserID(ctx.Request())
		reply, err := tripSvc.PublishTrip(ctx, req)
		return returnData(ctx, mobilePublishTripResponse(reply), err)
	})
	router.DELETE("/carpool/trips/{id}", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:update_status_self") {
			return nil
		}
		tripID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil || tripID <= 0 {
			return returnBadRequest(ctx, "琛岀▼ID鏍煎紡涓嶆纭紝璇峰埛鏂板悗閲嶈瘯")
		}
		err = tripSvc.DeleteTrip(ctx, &tripv1.DeleteTripRequest{Id: tripID, DriverId: currentUserID(ctx.Request())})
		return returnMessage(ctx, "deleted", err)
	})
	router.PUT("/carpool/trips/{id}/status", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:update_status_self") {
			return nil
		}
		req := new(tripv1.UpdateTripStatusRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		tripID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil || tripID <= 0 {
			return returnBadRequest(ctx, "琛岀▼ID鏍煎紡涓嶆纭紝璇峰埛鏂板悗閲嶈瘯")
		}
		req.Id = tripID
		return returnMessage(ctx, "updated", tripSvc.UpdateTripStatus(ctx, req))
	})
	router.POST("/carpool/trips/locations/validate", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:publish") {
			return nil
		}
		req := new(tripv1.ValidateLocationRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := tripSvc.ValidateLocation(ctx, req)
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/trips/price-preview", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:publish") {
			return nil
		}
		req := new(tripv1.PreviewTripPriceRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := tripSvc.PreviewTripPrice(ctx, req)
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/trips/locations/suggest", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "trip:publish") {
			return nil
		}
		req := new(tripv1.SuggestLocationsRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		reply, err := tripSvc.SuggestLocations(ctx, req)
		return returnData(ctx, reply, err)
	})
}

func registerOrderRoutes(srv *khttp.Server, permissions permissionChecker, orderSvc *service.OrderService, passengerSvc mobilePassengerProfileService, driverSvc mobileDriverProfileService, alipayCfg *conf.Alipay) {
	router := srv.Route("/")
	router.POST("/carpool/orders", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:create") {
			return nil
		}
		req := new(orderv1.CreateOrderRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		passengerID := currentUserID(ctx.Request())
		reply, err := orderSvc.CreateOrder(ctx, req.TripId, passengerID, req.SeatsBooked)
		if err != nil {
			logGatewayOrderAction(ctx, "create_order", passengerID, 0, "", "new", "created", err)
			return returnData(ctx, nil, err)
		}
		logGatewayOrderAction(ctx, "create_order", passengerID, 0, "", "new", "created", nil)
		return returnData(ctx, mobileCreateOrderResponse(reply), nil)
	})
	router.POST("/carpool/orders/{id}/cancel", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:cancel_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		passengerID := currentUserID(ctx.Request())
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.CancelOrder(ctx, orderID, passengerID, idempotencyKey)
		logGatewayOrderAction(ctx, "cancel_order", passengerID, orderID, idempotencyKey, "created|accepted|picking_up", "cancelled", err)
		return returnMessage(ctx, "cancelled", err)
	})
	router.GET("/carpool/orders", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:list_passenger_self") {
			return nil
		}
		query := ctx.Query()
		passengerID := currentUserID(ctx.Request())
		reply, err := orderSvc.ListOrders(ctx, passengerID, parseMobileOrderStatus(query.Get("status")), int32(parseInt(query.Get("page"))), int32(parseInt(query.Get("page_size"))))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, mobileOrderListResponse(reply), nil)
	})
	router.GET("/carpool/orders/pending", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:list_driver_pending") {
			return nil
		}
		query := ctx.Query()
		driverID := currentUserID(ctx.Request())
		reply, err := orderSvc.PendingOrders(ctx, driverID, int64(parseInt(query.Get("trip_id"))), int32(parseInt(query.Get("page"))), int32(parseInt(query.Get("page_size"))))
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, mobilePendingOrderListResponse(reply), nil)
	})
	router.GET("/carpool/orders/{id}", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:view_passenger_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		passengerID := currentUserID(ctx.Request())
		reply, err := orderSvc.GetOrderDetail(ctx, orderID, passengerID)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		payload := mobileOrderDetailResponse(reply)
		enrichMobileOrderContactPayload(ctx, payload, passengerSvc, driverSvc)
		return returnData(ctx, payload, nil)
	})
	router.POST("/carpool/orders/{id}/accept", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:accept_driver_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID := currentUserID(ctx.Request())
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.AcceptOrder(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "accept_order", driverID, orderID, idempotencyKey, "paid", "accepted", err)
		return returnMessage(ctx, "accepted", err)
	})
	router.POST("/carpool/orders/{id}/reject", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:reject_driver_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID := currentUserID(ctx.Request())
		actionReq := orderActionFromRequest(ctx)
		err = orderSvc.RejectOrder(ctx, orderID, driverID, actionReq.idempotencyKey(), actionReq.rejectReason())
		logGatewayOrderAction(ctx, "reject_order", driverID, orderID, actionReq.idempotencyKey(), "paid", "rejected", err)
		return returnMessage(ctx, "rejected", err)
	})
	router.POST("/carpool/orders/{id}/start-pickup", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:accept_driver_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID := currentUserID(ctx.Request())
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.StartPickup(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "start_pickup", driverID, orderID, idempotencyKey, "accepted", "picking_up", err)
		return returnMessage(ctx, "picking_up", err)
	})
	router.POST("/carpool/orders/{id}/start-delivery", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:accept_driver_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID := currentUserID(ctx.Request())
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.StartDelivery(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "start_delivery", driverID, orderID, idempotencyKey, "picking_up", "delivering", err)
		return returnMessage(ctx, "delivering", err)
	})
	router.POST("/carpool/orders/{id}/complete", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "order:accept_driver_self") {
			return nil
		}
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		driverID := currentUserID(ctx.Request())
		idempotencyKey := idempotencyKeyFromRequest(ctx)
		err = orderSvc.CompleteOrder(ctx, orderID, driverID, idempotencyKey)
		logGatewayOrderAction(ctx, "complete_order", driverID, orderID, idempotencyKey, "delivering", "completed", err)
		return returnMessage(ctx, "completed", err)
	})
	router.POST("/carpool/orders/{id}/pay", func(ctx khttp.Context) error {
		passengerID := currentUserID(ctx.Request())
		orderID, err := parseOrderIDParam(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, invalidOrderIDMessage)
		}
		payment, err := orderSvc.CreatePayment(ctx, orderID, passengerID, "alipay_sandbox")
		if err != nil {
			zap.L().Warn("prepare alipay failed to create payment", gatewayLogFields(ctx.Request(), zap.Int64("order_id", orderID), zap.Int64("user_id", passengerID), zap.Int64("passenger_id", passengerID), zap.Error(err))...)
			return returnData(ctx, nil, err)
		}
		client, err := newAlipayClient(alipayCfg)
		if err != nil {
			zap.L().Error("alipay config unavailable", zap.Error(err))
			return returnUnavailable(ctx, "alipay config unavailable")
		}
		form, err := client.CreateWapPay(ctx, alipayx.WapPayRequest{
			OutTradeNo:  payment.GetOutTradeNo(),
			Subject:     "RideHailing Order " + strconv.FormatInt(orderID, 10),
			TotalAmount: payment.GetTotalAmount(),
		})
		if err != nil {
			zap.L().Error("create alipay wap pay failed", gatewayLogFields(ctx.Request(), zap.Int64("order_id", orderID), zap.Int64("user_id", passengerID), zap.String("payment_no", payment.GetOutTradeNo()), zap.Error(err))...)
			return returnData(ctx, nil, err)
		}
		zap.L().Info("alipay wap pay prepared", gatewayLogFields(ctx.Request(), zap.Int64("order_id", orderID), zap.Int64("user_id", passengerID), zap.Int64("passenger_id", passengerID), zap.String("payment_no", payment.GetOutTradeNo()), zap.String("status_before", "created"), zap.String("status_after", "paying"))...)
		return returnData(ctx, map[string]any{
			"orderId":    int64String(orderID),
			"order_id":   int64String(orderID),
			"paymentId":  int64String(payment.GetPaymentId()),
			"payment_id": int64String(payment.GetPaymentId()),
			"paymentNo":  payment.GetOutTradeNo(),
			"payment_no": payment.GetOutTradeNo(),
			"channel":    "alipay_sandbox",
			"amount":     payment.GetTotalAmount(),
			"payForm":    form,
			"pay_form":   form,
		}, nil)
	})
}

func newAlipayClient(cfg *conf.Alipay) (*alipayx.Client, error) {
	if cfg == nil {
		cfg = &conf.Alipay{}
	}
	return alipayx.NewClient(alipayx.Config{
		AppID:           configOrEnv(cfg.AppID, "ALIPAY_APP_ID"),
		PrivateKey:      configOrEnv(cfg.PrivateKey, "ALIPAY_PRIVATE_KEY"),
		AlipayPublicKey: configOrEnv(cfg.AlipayPublicKey, "ALIPAY_PUBLIC_KEY"),
		Production:      cfg.Production,
		NotifyURL:       configOrEnv(cfg.NotifyURL, "ALIPAY_NOTIFY_URL"),
		ReturnURL:       configOrEnv(cfg.ReturnURL, "ALIPAY_RETURN_URL"),
	})
}

func configOrEnv(value string, envName string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		value = ""
	}
	if value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv(envName))
}

func formatAlipayAmount(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func registerReviewRoutes(srv *khttp.Server, permissions permissionChecker, reviewSvc *service.ReviewService) {
	router := srv.Route("/")
	router.POST("/carpool/reviews", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "review:submit") {
			return nil
		}
		req := new(reviewv1.SubmitReviewRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		req.FromUserId = currentUserID(ctx.Request())
		reply, err := reviewSvc.SubmitReview(ctx, req)
		return returnData(ctx, reply, err)
	})
}

func registerPassengerRoutes(srv *khttp.Server, permissions permissionChecker, passengerSvc *service.PassengerService) {
	router := srv.Route("/")
	router.GET("/carpool/passengers/me", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "passenger:profile:view_self") {
			return nil
		}
		userID := currentUserID(ctx.Request())
		reply, err := passengerSvc.EnsurePassenger(ctx, userID, "")
		return returnData(ctx, reply, err)
	})
	router.PUT("/carpool/passengers/me", func(ctx khttp.Context) error {
		req := new(passengerv1.UpdatePassengerRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "passenger:profile:update_self") {
			return nil
		}
		req.Id = currentUserID(ctx.Request())
		reply, err := passengerSvc.UpdatePassenger(ctx, req)
		return returnData(ctx, reply, err)
	})
}

func registerDriverRoutes(srv *khttp.Server, permissions permissionChecker, driverSvc *service.DriverService) {
	router := srv.Route("/")
	router.GET("/carpool/drivers/me", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:profile:view_self") {
			return nil
		}
		userID := currentUserID(ctx.Request())
		reply, err := driverSvc.EnsureDriver(ctx, userID, "")
		return returnData(ctx, reply, err)
	})
	router.PUT("/carpool/drivers/me", func(ctx khttp.Context) error {
		req := new(driverv1.UpdateDriverRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:profile:update_self") {
			return nil
		}
		req.Id = currentUserID(ctx.Request())
		reply, err := driverSvc.UpdateDriver(ctx, req)
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/drivers/certification", func(ctx khttp.Context) error {
		req := new(driverv1.SubmitCertificationRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:certification:submit_self") {
			return nil
		}
		req.Id = currentUserID(ctx.Request())
		reply, err := driverSvc.SubmitCertification(ctx, req)
		return returnData(ctx, reply, err)
	})
	router.GET("/carpool/drivers/certification", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:certification:view_self") {
			return nil
		}
		userID := currentUserID(ctx.Request())
		reply, err := driverSvc.GetCertification(ctx, userID)
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/drivers/vehicles", func(ctx khttp.Context) error {
		req := new(driverv1.SaveVehicleRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:vehicle:manage_self") {
			return nil
		}
		req.Id = currentUserID(ctx.Request())
		reply, err := driverSvc.SaveVehicle(ctx, req)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, safeVehicleReply(reply), nil)
	})
	router.PUT("/carpool/drivers/vehicles/{id}", func(ctx khttp.Context) error {
		req := new(driverv1.UpdateVehicleRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:vehicle:manage_self") {
			return nil
		}
		vehicleID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, "车辆ID格式不正确，请刷新后重试")
		}
		req.DriverId = currentUserID(ctx.Request())
		req.VehicleId = vehicleID
		reply, err := driverSvc.UpdateVehicle(ctx, req)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, safeVehicleReply(reply), nil)
	})
	router.DELETE("/carpool/drivers/vehicles/{id}", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:vehicle:manage_self") {
			return nil
		}
		vehicleID, err := parseInt64Param(ctx.Vars().Get("id"))
		if err != nil {
			return returnBadRequest(ctx, "车辆ID格式不正确，请刷新后重试")
		}
		req := &driverv1.DeleteVehicleRequest{
			DriverId:  currentUserID(ctx.Request()),
			VehicleId: vehicleID,
		}
		reply, err := driverSvc.DeleteVehicle(ctx, req)
		return returnData(ctx, reply, err)
	})
	router.GET("/carpool/drivers/vehicles", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:vehicle:list_self") {
			return nil
		}
		userID := currentUserID(ctx.Request())
		reply, err := driverSvc.ListVehicles(ctx, userID)
		if err != nil {
			return returnData(ctx, nil, err)
		}
		return returnData(ctx, safeVehicleListReply(reply), nil)
	})
	router.GET("/carpool/drivers/messages", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:profile:view_self") {
			return nil
		}
		reply, err := driverSvc.ListMessages(ctx, currentUserID(ctx.Request()))
		return returnData(ctx, reply, err)
	})
	router.POST("/carpool/drivers/messages/{id}/ack", func(ctx khttp.Context) error {
		if !requirePermission(ctx.Response(), ctx.Request(), permissions, "driver:profile:view_self") {
			return nil
		}
		req := &driverv1.AckMessageRequest{
			DriverId:  currentUserID(ctx.Request()),
			MessageId: int64(parseInt(ctx.Vars().Get("id"))),
		}
		reply, err := driverSvc.AckMessage(ctx, req)
		return returnData(ctx, reply, err)
	})
}

func returnData(ctx khttp.Context, payload any, err error) error {
	if err != nil {
		return returnGatewayError(ctx, err)
	}
	return ctx.Returns(map[string]any{"code": 0, "data": payload, "msg": "success"}, nil)
}

func returnMessage(ctx khttp.Context, msg string, err error) error {
	if err != nil {
		return returnGatewayError(ctx, err)
	}
	return ctx.Returns(map[string]any{"code": 0, "data": nil, "msg": msg}, nil)
}

type orderActionRequest struct {
	IdempotencyKey      string `json:"idempotency_key"`
	IdempotencyKeyCamel string `json:"idempotencyKey"`
	RejectReason        string `json:"reject_reason"`
	RejectReasonCamel   string `json:"rejectReason"`
}

func (r *orderActionRequest) idempotencyKey() string {
	if r == nil {
		return ""
	}
	if key := strings.TrimSpace(r.IdempotencyKey); key != "" {
		return key
	}
	return strings.TrimSpace(r.IdempotencyKeyCamel)
}

func (r *orderActionRequest) rejectReason() string {
	if r == nil {
		return ""
	}
	if reason := strings.TrimSpace(r.RejectReason); reason != "" {
		return reason
	}
	return strings.TrimSpace(r.RejectReasonCamel)
}

func orderActionFromRequest(ctx khttp.Context) *orderActionRequest {
	req := new(orderActionRequest)
	if ctx == nil {
		return req
	}
	if err := ctx.Bind(req); err != nil {
		return req
	}
	return req
}

func idempotencyKeyFromRequest(ctx khttp.Context) string {
	if ctx == nil || ctx.Request() == nil {
		return ""
	}
	if key := strings.TrimSpace(ctx.Request().Header.Get("Idempotency-Key")); key != "" {
		return key
	}
	return orderActionFromRequest(ctx).idempotencyKey()
}

func requestTraceID(req *http.Request) string {
	if req == nil {
		return ""
	}
	for _, key := range []string{"X-Trace-Id", "X-Request-Id", "Traceparent"} {
		if value := strings.TrimSpace(req.Header.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func gatewayLogFields(req *http.Request, fields ...zap.Field) []zap.Field {
	if traceID := requestTraceID(req); traceID != "" {
		fields = append([]zap.Field{zap.String("trace_id", traceID)}, fields...)
	}
	return fields
}

func logGatewayOrderAction(ctx khttp.Context, action string, userID, orderID int64, idempotencyKey, statusBefore, statusAfter string, err error) {
	fields := gatewayLogFields(ctx.Request(),
		zap.String("action", action),
		zap.Int64("user_id", userID),
		zap.Int64("order_id", orderID),
		zap.String("idempotency_key", idempotencyKey),
		zap.String("status_before", statusBefore),
		zap.String("status_after", statusAfter),
	)
	if err != nil {
		zap.L().Warn("gateway order action failed", append(fields, zap.Error(err))...)
		return
	}
	zap.L().Info("gateway order action completed", fields...)
}

func returnGatewayError(ctx khttp.Context, err error) error {
	statusCode, msg := gatewayErrorStatusAndMessage(err)
	zap.L().Warn("gateway request failed", zap.Int("status_code", statusCode), zap.String("message", msg), zap.Error(err))
	return writeGatewayJSON(ctx, statusCode, msg)
}

func gatewayErrorStatusAndMessage(err error) (int, string) {
	var upstreamErr *data.UpstreamHTTPError
	if errors.As(err, &upstreamErr) {
		statusCode := upstreamErr.StatusCode
		if statusCode < http.StatusBadRequest {
			statusCode = http.StatusInternalServerError
		}
		return statusCode, fallbackChineseMessage(upstreamErr.Message, statusCode)
	}
	if st, ok := grpcstatus.FromError(err); ok {
		statusCode := httpStatusFromGRPCCode(st.Code())
		return statusCode, fallbackChineseMessage(st.Message(), statusCode)
	}
	return http.StatusInternalServerError, fallbackChineseMessage("", http.StatusInternalServerError)
}

func httpStatusFromGRPCCode(code codes.Code) int {
	switch code {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func fallbackChineseMessage(msg string, statusCode int) string {
	msg = strings.TrimSpace(msg)
	if msg != "" && msg != "internal error" {
		return msg
	}
	switch statusCode {
	case http.StatusBadRequest:
		return "请求参数不正确，请检查后重试"
	case http.StatusUnauthorized:
		return "登录已过期，请重新登录"
	case http.StatusForbidden:
		return "暂无权限执行该操作"
	case http.StatusNotFound:
		return "订单不存在或已被处理，请刷新后重试"
	case http.StatusServiceUnavailable:
		return "服务暂时不可用，请稍后重试"
	default:
		return "服务开小差了，请稍后重试"
	}
}

func returnBadRequest(ctx khttp.Context, msg string) error {
	return writeGatewayJSON(ctx, http.StatusBadRequest, msg)
}

func writeGatewayJSON(ctx khttp.Context, statusCode int, msg string) error {
	ctx.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Response().WriteHeader(statusCode)
	return json.NewEncoder(ctx.Response()).Encode(map[string]any{
		"code": statusCode,
		"data": nil,
		"msg":  msg,
	})
}

const invalidOrderIDMessage = "订单ID格式不正确，请刷新后重试"

func parseOrderIDParam(value string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid order id")
	}
	return id, nil
}

func parseInt64Param(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseInt(value string) int {
	n, _ := strconv.Atoi(value)
	return n
}
