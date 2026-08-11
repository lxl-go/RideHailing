package data

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderHTTPClientCreateOrderPostsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/orders", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order_id":5001,"total_price":39.8}`))
	}))
	defer server.Close()

	client := NewOrderHTTPClient(server.URL)
	reply, err := client.CreateOrder(t.Context(), 1001, 3001, 2)

	require.NoError(t, err)
	require.Equal(t, int64(5001), reply.OrderId)
	require.Equal(t, 39.8, reply.TotalPrice)
}

func TestOrderHTTPClientCompleteOrderPostsContractBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/driver/orders/5001/complete", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, float64(5001), body["id"])
		require.Equal(t, float64(2001), body["driver_id"])
		require.Equal(t, "complete-5001", body["idempotency_key"])
		require.Equal(t, "complete-5001", r.Header.Get("Idempotency-Key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewOrderHTTPClient(server.URL)
	err := client.CompleteOrder(t.Context(), 5001, 2001, "complete-5001")

	require.NoError(t, err)
}

func TestOrderHTTPClientReturnsUpstreamHTTPErrorWithBodyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/orders", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"seat inventory not enough"}`))
	}))
	defer server.Close()

	client := NewOrderHTTPClient(server.URL)
	_, err := client.CreateOrder(t.Context(), 1001, 3001, 2)

	require.Error(t, err)
	var upstreamErr *UpstreamHTTPError
	require.True(t, errors.As(err, &upstreamErr))
	require.Equal(t, http.StatusBadRequest, upstreamErr.StatusCode)
	require.Equal(t, "seat inventory not enough", upstreamErr.Message)
}
