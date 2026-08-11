package data

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassengerHTTPClientGetPassengerUsesProfilePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/passengers/1001", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"passenger":{"id":1001,"nickname":"Alice","status":1}}`))
	}))
	defer server.Close()

	client := NewPassengerHTTPClient(server.URL)
	reply, err := client.GetPassenger(t.Context(), 1001)

	require.NoError(t, err)
	require.Equal(t, int64(1001), reply.Passenger.Id)
	require.Equal(t, "Alice", reply.Passenger.Nickname)
}
