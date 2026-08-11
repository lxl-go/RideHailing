package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/require"

	gatewaybiz "ride-hailing/services/gateway-service/internal/biz"
	gatewaydata "ride-hailing/services/gateway-service/internal/data"
	gatewayservice "ride-hailing/services/gateway-service/internal/service"
	reviewv1 "ride-hailing/services/review-service/api/review/v1"
	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

func TestMobileCompatibilityReadRoutesUseRealServices(t *testing.T) {
	srv, tripClient, reviewClient := newMobileCompatibilityTestServer()

	coupons := doMobileCompatJSONWithUser(srv, http.MethodGet, "/carpool/coupons?page=2&page_size=10", "", 1001)
	require.Equal(t, http.StatusOK, coupons.Code)
	couponPayload := decodeGatewayBody(t, coupons)
	require.Equal(t, float64(0), couponPayload["code"])
	couponData := couponPayload["data"].(map[string]any)
	require.Equal(t, float64(1), couponData["total"])
	coupon := couponData["list"].([]any)[0].(map[string]any)
	require.Equal(t, "CPN-001", coupon["couponNo"])
	require.Equal(t, float64(8.5), coupon["amount"])
	require.Equal(t, int64(1001), tripClient.listCouponsReq.UserId)
	require.Equal(t, int32(2), tripClient.listCouponsReq.Page)
	require.Equal(t, int32(10), tripClient.listCouponsReq.PageSize)

	demands := doMobileCompatJSONWithUser(srv, http.MethodGet, "/carpool/trips/demands/mine?status=1&page=1&page_size=20", "", 1001)
	require.Equal(t, http.StatusOK, demands.Code)
	demandPayload := decodeGatewayBody(t, demands)
	require.Equal(t, float64(0), demandPayload["code"])
	demandData := demandPayload["data"].(map[string]any)
	require.Equal(t, float64(1), demandData["total"])
	demand := demandData["items"].([]any)[0].(map[string]any)
	require.Equal(t, "1001", demand["passengerId"])
	require.Equal(t, "A", demand["origin"])
	require.Equal(t, int64(1001), tripClient.listMyDemandsReq.PassengerId)
	require.Equal(t, int32(1), tripClient.listMyDemandsReq.Status)

	reviews := doMobileCompatJSONWithUser(srv, http.MethodGet, "/carpool/reviews/mine/5001", "", 1001)
	require.Equal(t, http.StatusOK, reviews.Code)
	reviewPayload := decodeGatewayBody(t, reviews)
	require.Equal(t, float64(0), reviewPayload["code"])
	reviewData := reviewPayload["data"].(map[string]any)
	require.Equal(t, true, reviewData["hasReview"])
	review := reviewData["review"].(map[string]any)
	require.Equal(t, "5001", review["orderId"])
	require.Equal(t, float64(5), review["rating"])
	require.Equal(t, int64(5001), reviewClient.getMyReviewReq.OrderId)
	require.Equal(t, int64(1001), reviewClient.getMyReviewReq.FromUserId)
}

func TestMobileCompatibilityActionRoutesUseRealServices(t *testing.T) {
	srv, tripClient, _ := newMobileCompatibilityTestServer()

	claim := doMobileCompatJSONWithUserAndIdempotency(srv, http.MethodPost, "/carpool/coupons/claim", `{"coupon_id":"CPN-001"}`, 1001, "idem-claim")
	require.Equal(t, http.StatusOK, claim.Code)
	claimPayload := decodeGatewayBody(t, claim)
	require.Equal(t, float64(0), claimPayload["code"])
	claimData := claimPayload["data"].(map[string]any)
	require.Equal(t, true, claimData["duplicated"])
	require.Equal(t, int64(1001), tripClient.claimCouponReq.UserId)
	require.Equal(t, "CPN-001", tripClient.claimCouponReq.CouponNo)
	require.Equal(t, "idem-claim", tripClient.claimCouponReq.IdempotencyKey)

	publish := doMobileCompatJSONWithUser(srv, http.MethodPost, "/carpool/trips/demands", `{"origin":"A","destination":"B","depart_time":"2026-08-04 09:00:00","seats":2,"budget":30.5,"remark":"near gate"}`, 1001)
	require.Equal(t, http.StatusOK, publish.Code)
	publishPayload := decodeGatewayBody(t, publish)
	require.Equal(t, float64(0), publishPayload["code"])
	publishData := publishPayload["data"].(map[string]any)
	require.Equal(t, "1001", publishData["passengerId"])
	require.Equal(t, "A", publishData["origin"])
	require.Equal(t, int64(1001), tripClient.publishDemandReq.PassengerId)
	require.Equal(t, "B", tripClient.publishDemandReq.Destination)

	cancel := doMobileCompatJSONWithUser(srv, http.MethodPost, "/carpool/trips/demands/701/cancel", `{}`, 1001)
	require.Equal(t, http.StatusOK, cancel.Code)
	cancelPayload := decodeGatewayBody(t, cancel)
	require.Equal(t, float64(0), cancelPayload["code"])
	require.Equal(t, int64(701), tripClient.cancelDemandReq.Id)
	require.Equal(t, int64(1001), tripClient.cancelDemandReq.PassengerId)

	deleted := doMobileCompatJSONWithUser(srv, http.MethodDelete, "/carpool/trips/801", `{}`, 1001)
	require.Equal(t, http.StatusOK, deleted.Code)
	deletePayload := decodeGatewayBody(t, deleted)
	require.Equal(t, float64(0), deletePayload["code"])
	require.Equal(t, int64(801), tripClient.deleteTripReq.Id)
	require.Equal(t, int64(1001), tripClient.deleteTripReq.DriverId)
}

func TestMobileCompatibilityActionRoutesRejectInvalidIDs(t *testing.T) {
	srv := khttp.NewServer()
	tripSvc := gatewayservice.NewTripService(gatewaybiz.NewTripUsecase(&fakeTripClient{}))
	reviewSvc := gatewayservice.NewReviewService(gatewaybiz.NewReviewUsecase(&fakeReviewClient{}))
	registerMobileCompatibilityRoutes(srv, tripSvc, reviewSvc)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/carpool/trips/demands/abc/cancel"},
		{http.MethodDelete, "/carpool/trips/abc"},
	}

	for _, tc := range cases {
		res := doMobileCompatJSONWithUser(srv, tc.method, tc.path, `{}`, 1001)
		require.Equal(t, http.StatusBadRequest, res.Code, tc.path)
		payload := decodeGatewayBody(t, res)
		require.NotEqual(t, float64(0), payload["code"], tc.path)
	}
}

func TestMobileCompatibilityRecommendTripsRequiresDemand(t *testing.T) {
	srv, tripClient, _ := newMobileCompatibilityTestServer()
	tripClient.listMyDemandsReply = &tripv1.ListMyDemandsReply{Total: 0}
	tripClient.searchTripsReply = &tripv1.SearchTripsReply{
		Total: 1,
		Items: []*tripv1.TripItem{{Id: 801, Origin: "A", Destination: "B", Status: 1}},
	}

	res := doMobileCompatJSONWithUser(srv, http.MethodGet, "/carpool/trips/demands/recommendations?page=1&page_size=5", "", 1001)
	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	require.Equal(t, float64(0), data["total"])
	require.Empty(t, data["items"])
	require.Equal(t, int64(1001), tripClient.listMyDemandsReq.PassengerId)
	require.Equal(t, int32(1), tripClient.listMyDemandsReq.Status)
	require.Equal(t, 0, tripClient.searchTripsCalls)
}

func TestMobileCompatibilityRecommendTripsSortsByDemandMatch(t *testing.T) {
	srv, tripClient, _ := newMobileCompatibilityTestServer()
	demandDepart := time.Now().Add(24 * time.Hour)
	tripClient.listMyDemandsReply = &tripv1.ListMyDemandsReply{
		Total: 1,
		Items: []*tripv1.TripDemandItem{{
			Id:          701,
			PassengerId: 1001,
			Origin:      "人民广场",
			Destination: "虹桥火车站",
			DepartTime:  demandDepart.Format("2006-01-02 15:04:05"),
			Seats:       2,
			Budget:      80,
			Status:      1,
		}},
	}
	tripClient.searchTripsReply = &tripv1.SearchTripsReply{
		Total: 4,
		Items: []*tripv1.TripItem{
			{
				Id:             802,
				DriverId:       12,
				Origin:         "张江",
				Destination:    "浦东机场",
				DepartTime:     demandDepart.Add(2 * time.Hour).Format("2006-01-02 15:04:05"),
				SeatsAvailable: 3,
				Price:          120,
				Status:         1,
			},
			{
				Id:             803,
				DriverId:       13,
				Origin:         "人民广场",
				Destination:    "虹桥火车站",
				DepartTime:     demandDepart.Add(20 * time.Minute).Format("2006-01-02 15:04:05"),
				SeatsAvailable: 2,
				Price:          78,
				Status:         1,
			},
			{
				Id:             804,
				DriverId:       14,
				Origin:         "人民广场",
				Destination:    "虹桥火车站",
				DepartTime:     time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05"),
				SeatsAvailable: 2,
				Price:          76,
				Status:         1,
			},
			{
				Id:             805,
				DriverId:       15,
				Origin:         "人民广场",
				Destination:    "虹桥火车站",
				DepartTime:     demandDepart.Add(10 * time.Minute).Format("2006-01-02 15:04:05"),
				SeatsAvailable: 2,
				Price:          79,
				Status:         0,
			},
		},
	}

	res := doMobileCompatJSONWithUser(srv, http.MethodGet, "/carpool/trips/demands/recommendations?page=1&page_size=5", "", 1001)
	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	data := payload["data"].(map[string]any)
	items := data["items"].([]any)
	require.Len(t, items, 2)
	first := items[0].(map[string]any)
	second := items[1].(map[string]any)
	require.Equal(t, "803", first["id"])
	require.Greater(t, first["matchScore"].(float64), second["matchScore"].(float64))
	require.NotEmpty(t, first["matchReason"])
	require.Equal(t, "人民广场", tripClient.searchTripsReq.Origin)
	require.Equal(t, "虹桥火车站", tripClient.searchTripsReq.Destination)
	require.Equal(t, demandDepart.Format("2006-01-02"), tripClient.searchTripsReq.DepartDate)
	require.Equal(t, int32(1), tripClient.searchTripsReq.Page)
	require.Equal(t, int32(5), tripClient.searchTripsReq.PageSize)
}

func newMobileCompatibilityTestServer() (*khttp.Server, *fakeTripClient, *fakeReviewClient) {
	srv := khttp.NewServer()
	tripClient := &fakeTripClient{}
	reviewClient := &fakeReviewClient{}
	tripSvc := gatewayservice.NewTripService(gatewaybiz.NewTripUsecase(tripClient))
	reviewSvc := gatewayservice.NewReviewService(gatewaybiz.NewReviewUsecase(reviewClient))
	registerMobileCompatibilityRoutes(srv, tripSvc, reviewSvc)
	return srv, tripClient, reviewClient
}

func doMobileCompatJSONWithUser(handler http.Handler, method, path, body string, userID int64) *httptest.ResponseRecorder {
	return doMobileCompatJSONWithUserAndIdempotency(handler, method, path, body, userID, "")
}

func doMobileCompatJSONWithUserAndIdempotency(handler http.Handler, method, path, body string, userID int64, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if userID > 0 {
		req = withCurrentUser(req, CurrentUser{UserID: userID})
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

type fakeTripClient struct {
	listCouponsReq     *tripv1.ListCouponsRequest
	claimCouponReq     *tripv1.ClaimCouponRequest
	searchTripsReq     gatewaydata.SearchTripsRequest
	searchTripsCalls   int
	searchTripsReply   *tripv1.SearchTripsReply
	publishDemandReq   *tripv1.PublishDemandRequest
	listMyDemandsReq   *tripv1.ListMyDemandsRequest
	listMyDemandsReply *tripv1.ListMyDemandsReply
	cancelDemandReq    *tripv1.CancelDemandRequest
	deleteTripReq      *tripv1.DeleteTripRequest
}

func (c *fakeTripClient) SearchTrips(_ context.Context, req gatewaydata.SearchTripsRequest) (*tripv1.SearchTripsReply, error) {
	c.searchTripsReq = req
	c.searchTripsCalls++
	if c.searchTripsReply != nil {
		return c.searchTripsReply, nil
	}
	return &tripv1.SearchTripsReply{}, nil
}

func (c *fakeTripClient) GetTripDetail(context.Context, int64) (*tripv1.GetTripDetailReply, error) {
	return &tripv1.GetTripDetailReply{}, nil
}

func (c *fakeTripClient) PublishTrip(context.Context, *tripv1.PublishTripRequest) (*tripv1.PublishTripReply, error) {
	return &tripv1.PublishTripReply{}, nil
}

func (c *fakeTripClient) ListDriverTrips(context.Context, *tripv1.ListDriverTripsRequest) (*tripv1.ListDriverTripsReply, error) {
	return &tripv1.ListDriverTripsReply{}, nil
}

func (c *fakeTripClient) UpdateTripStatus(context.Context, *tripv1.UpdateTripStatusRequest) error {
	return nil
}

func (c *fakeTripClient) ListCoupons(_ context.Context, req *tripv1.ListCouponsRequest) (*tripv1.ListCouponsReply, error) {
	c.listCouponsReq = req
	return &tripv1.ListCouponsReply{
		Total: 1,
		Items: []*tripv1.CouponItem{{
			Id:              601,
			CouponNo:        "CPN-001",
			CouponCode:      "SAVE85",
			Name:            "commute coupon",
			CouponType:      "amount",
			FaceValue:       8.5,
			ThresholdAmount: 20,
			Status:          "available",
			Claimed:         false,
		}},
	}, nil
}

func (c *fakeTripClient) ClaimCoupon(_ context.Context, req *tripv1.ClaimCouponRequest) (*tripv1.ClaimCouponReply, error) {
	c.claimCouponReq = req
	return &tripv1.ClaimCouponReply{
		Duplicated: true,
		Coupon: &tripv1.CouponItem{
			Id:         601,
			CouponNo:   req.CouponNo,
			CouponCode: "SAVE85",
			Name:       "commute coupon",
			Status:     "claimed",
			Claimed:    true,
		},
	}, nil
}

func (c *fakeTripClient) PublishDemand(_ context.Context, req *tripv1.PublishDemandRequest) (*tripv1.PublishDemandReply, error) {
	c.publishDemandReq = req
	return &tripv1.PublishDemandReply{Demand: &tripv1.TripDemandItem{
		Id:          701,
		PassengerId: req.PassengerId,
		Origin:      req.Origin,
		Destination: req.Destination,
		DepartTime:  req.DepartTime,
		Seats:       req.Seats,
		Budget:      req.Budget,
		Remark:      req.Remark,
		Status:      1,
	}}, nil
}

func (c *fakeTripClient) ListMyDemands(_ context.Context, req *tripv1.ListMyDemandsRequest) (*tripv1.ListMyDemandsReply, error) {
	c.listMyDemandsReq = req
	if c.listMyDemandsReply != nil {
		return c.listMyDemandsReply, nil
	}
	return &tripv1.ListMyDemandsReply{
		Total: 1,
		Items: []*tripv1.TripDemandItem{{
			Id:          701,
			PassengerId: req.PassengerId,
			Origin:      "A",
			Destination: "B",
			Status:      req.Status,
		}},
	}, nil
}

func (c *fakeTripClient) CancelDemand(_ context.Context, req *tripv1.CancelDemandRequest) error {
	c.cancelDemandReq = req
	return nil
}

func (c *fakeTripClient) DeleteTrip(_ context.Context, req *tripv1.DeleteTripRequest) error {
	c.deleteTripReq = req
	return nil
}

func (c *fakeTripClient) ValidateLocation(_ context.Context, req *tripv1.ValidateLocationRequest) (*tripv1.ValidateLocationReply, error) {
	loc := req.GetLocation()
	if loc == nil {
		return nil, nil
	}
	return &tripv1.ValidateLocationReply{Location: &tripv1.LocationInput{Name: loc.Name}}, nil
}

func (c *fakeTripClient) PreviewTripPrice(_ context.Context, req *tripv1.PreviewTripPriceRequest) (*tripv1.PreviewTripPriceReply, error) {
	return &tripv1.PreviewTripPriceReply{}, nil
}

func (c *fakeTripClient) SuggestLocations(_ context.Context, req *tripv1.SuggestLocationsRequest) (*tripv1.SuggestLocationsReply, error) {
	return &tripv1.SuggestLocationsReply{
		Locations: []*tripv1.LocationInput{{Name: req.GetKeyword(), FormattedAddress: req.GetKeyword()}},
	}, nil
}

type fakeReviewClient struct {
	getMyReviewReq *reviewv1.GetMyReviewRequest
}

func (c *fakeReviewClient) SubmitReview(context.Context, *reviewv1.SubmitReviewRequest) (*reviewv1.SubmitReviewReply, error) {
	return &reviewv1.SubmitReviewReply{}, nil
}

func (c *fakeReviewClient) GetMyReview(_ context.Context, req *reviewv1.GetMyReviewRequest) (*reviewv1.GetMyReviewReply, error) {
	c.getMyReviewReq = req
	return &reviewv1.GetMyReviewReply{
		HasReview: true,
		Review: &reviewv1.ReviewItem{
			Id:         901,
			OrderId:    req.OrderId,
			FromUserId: req.FromUserId,
			ToUserId:   2002,
			Rating:     5,
			Content:    "good",
			CreatedAt:  "2026-08-04 10:00:00",
		},
	}, nil
}
