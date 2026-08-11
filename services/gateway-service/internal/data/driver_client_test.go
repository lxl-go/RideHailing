package data

import (
	"net/http"
	"net/http/httptest"
	"testing"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"

	"github.com/stretchr/testify/require"
)

func TestDriverHTTPClientSubmitCertificationUsesDriverPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/drivers/2001/certification", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"certification":{"id":3001,"driver_id":2001,"status":2}}`))
	}))
	defer server.Close()

	client := NewDriverHTTPClient(server.URL)
	reply, err := client.SubmitCertification(t.Context(), &driverv1.SubmitCertificationRequest{
		Id:        2001,
		RealName:  "Bob",
		LicenseNo: "DL001",
	})

	require.NoError(t, err)
	require.Equal(t, int64(3001), reply.Certification.Id)
	require.Equal(t, int32(2), reply.Certification.Status)
}

func TestDriverHTTPClientParsesProtoJSONInt64Strings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/drivers/3/certification", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"certification":{"id":"2084079732630102016","driverId":"3","realName":"Bob","licenseNo":"DL001","licenseType":"C1","status":2}}`))
	}))
	defer server.Close()

	client := NewDriverHTTPClient(server.URL)
	reply, err := client.GetCertification(t.Context(), 3)

	require.NoError(t, err)
	require.Equal(t, int64(2084079732630102016), reply.Certification.Id)
	require.Equal(t, int64(3), reply.Certification.DriverId)
	require.Equal(t, "C1", reply.Certification.LicenseType)
}

func TestDriverHTTPClientPreservesCertificationUpstreamHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/drivers/2001/certification", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":403,"message":"real-name authentication failed"}`))
	}))
	defer server.Close()

	client := NewDriverHTTPClient(server.URL)
	_, err := client.SubmitCertification(t.Context(), &driverv1.SubmitCertificationRequest{
		Id:          2001,
		RealName:    "徐乐",
		IdCardNo:    "652901196611026716",
		LicenseNo:   "123",
		LicenseType: "C1",
	})

	require.Error(t, err)
	upstreamErr, ok := err.(*UpstreamHTTPError)
	require.True(t, ok)
	require.Equal(t, http.StatusForbidden, upstreamErr.StatusCode)
	require.Equal(t, "real-name authentication failed", upstreamErr.Message)
}
