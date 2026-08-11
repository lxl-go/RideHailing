package data

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	tripv1 "ride-hailing/services/trip-service/api/trip/v1"
)

func TestTripHTTPClientSearchTripsForwardsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/trips", r.URL.Path)
		require.Equal(t, "A", r.URL.Query().Get("origin"))
		require.Equal(t, "B", r.URL.Query().Get("destination"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":1,"items":[{"id":101,"origin":"A","destination":"B"}]}`))
	}))
	defer server.Close()

	client := NewTripHTTPClient(server.URL)
	reply, err := client.SearchTrips(t.Context(), SearchTripsRequest{
		Origin:      "A",
		Destination: "B",
		Page:        1,
		PageSize:    20,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), reply.Total)
	require.Len(t, reply.Items, 1)
	require.Equal(t, int64(101), reply.Items[0].Id)
}

func TestTripHTTPClientCouponMethodsForwardContracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/coupons":
			require.Equal(t, "1001", r.URL.Query().Get("user_id"))
			require.Equal(t, "2", r.URL.Query().Get("page"))
			require.Equal(t, "10", r.URL.Query().Get("page_size"))
			_, _ = w.Write([]byte(`{"total":1,"items":[{"id":601,"coupon_no":"CPN-001","name":"commute","face_value":8.5}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/coupons/claim":
			var req tripv1.ClaimCouponRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			require.Equal(t, int64(1001), req.UserId)
			require.Equal(t, "CPN-001", req.CouponNo)
			require.Equal(t, "idem-001", req.IdempotencyKey)
			_, _ = w.Write([]byte(`{"duplicated":true,"coupon":{"id":601,"coupon_no":"CPN-001","claimed":true}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := NewTripHTTPClient(server.URL)
	coupons, err := client.ListCoupons(t.Context(), &tripv1.ListCouponsRequest{UserId: 1001, Page: 2, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), coupons.Total)
	require.Equal(t, "CPN-001", coupons.Items[0].CouponNo)

	claim, err := client.ClaimCoupon(t.Context(), &tripv1.ClaimCouponRequest{
		UserId:         1001,
		CouponNo:       "CPN-001",
		IdempotencyKey: "idem-001",
	})
	require.NoError(t, err)
	require.True(t, claim.Duplicated)
	require.Equal(t, int64(601), claim.Coupon.Id)
}

func TestTripHTTPClientDemandMethodsForwardContracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/trips/demands":
			var req tripv1.PublishDemandRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			require.Equal(t, int64(1001), req.PassengerId)
			require.Equal(t, "A", req.Origin)
			require.Equal(t, "B", req.Destination)
			_, _ = w.Write([]byte(`{"demand":{"id":701,"passenger_id":1001,"origin":"A","destination":"B","status":1}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/trips/demands/mine":
			require.Equal(t, "1001", r.URL.Query().Get("passenger_id"))
			require.Equal(t, "1", r.URL.Query().Get("status"))
			require.Equal(t, "1", r.URL.Query().Get("page"))
			require.Equal(t, "20", r.URL.Query().Get("page_size"))
			_, _ = w.Write([]byte(`{"total":1,"items":[{"id":701,"passenger_id":1001,"origin":"A","destination":"B","status":1}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/trips/demands/701/cancel":
			var req tripv1.CancelDemandRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			require.Equal(t, int64(701), req.Id)
			require.Equal(t, int64(1001), req.PassengerId)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/driver/trips/801":
			require.Equal(t, "2001", r.URL.Query().Get("driver_id"))
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Empty(t, string(body))
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := NewTripHTTPClient(server.URL)
	published, err := client.PublishDemand(t.Context(), &tripv1.PublishDemandRequest{
		PassengerId: 1001,
		Origin:      "A",
		Destination: "B",
	})
	require.NoError(t, err)
	require.Equal(t, int64(701), published.Demand.Id)

	demands, err := client.ListMyDemands(t.Context(), &tripv1.ListMyDemandsRequest{
		PassengerId: 1001,
		Status:      1,
		Page:        1,
		PageSize:    20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), demands.Total)

	require.NoError(t, client.CancelDemand(t.Context(), &tripv1.CancelDemandRequest{Id: 701, PassengerId: 1001}))
	require.NoError(t, client.DeleteTrip(t.Context(), &tripv1.DeleteTripRequest{Id: 801, DriverId: 2001}))
}
