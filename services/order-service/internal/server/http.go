package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	orderv1 "ride-hailing/services/order-service/api/order/v1"
	"ride-hailing/services/order-service/internal/conf"
	"ride-hailing/services/order-service/internal/service"
)

func NewHTTPServer(c *conf.Server, orderSvc *service.OrderService) *khttp.Server {
	opts := []khttp.ServerOption{}
	if c != nil && c.Http != nil {
		if c.Http.Addr != "" {
			opts = append(opts, khttp.Address(c.Http.Addr))
		}
		if d, err := time.ParseDuration(c.Http.Timeout); err == nil && d > 0 {
			opts = append(opts, khttp.Timeout(d))
		}
	}
	srv := khttp.NewServer(opts...)
	orderv1.RegisterOrderServiceHTTPServer(srv, orderSvc)
	registerOrderCompatibilityRoutes(srv, orderSvc)
	return srv
}

func registerOrderCompatibilityRoutes(srv *khttp.Server, orderSvc *service.OrderService) {
	router := srv.Route("/")
	router.POST("/v1/driver/orders/{id}/complete", func(ctx khttp.Context) error {
		driverID, err := parseInt64Field(ctx, "driver_id", "driverId")
		if err != nil {
			return writeOrderBadRequest(ctx, "driver_id is required")
		}
		id, _ := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
		if _, err := orderSvc.CompleteOrder(ctx, &orderv1.CompleteOrderRequest{
			Id:       id,
			DriverId: driverID,
		}); err != nil {
			return err
		}
		return ctx.Returns(map[string]any{"code": 0, "msg": "completed", "data": nil}, nil)
	})
}

func parseInt64Field(ctx khttp.Context, names ...string) (int64, error) {
	var body map[string]any
	if err := ctx.Bind(&body); err != nil {
		return 0, err
	}
	for _, name := range names {
		raw, ok := body[name]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			return int64(v), nil
		case string:
			return strconv.ParseInt(v, 10, 64)
		}
	}
	return 0, strconv.ErrSyntax
}

func writeOrderBadRequest(ctx khttp.Context, msg string) error {
	ctx.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Response().WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(ctx.Response()).Encode(map[string]any{"code": http.StatusBadRequest, "msg": msg, "data": nil})
}
