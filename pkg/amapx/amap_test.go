package amapx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeocodeParsesLocationAndSendsKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v3/geocode/geo", r.URL.Path)
		require.Equal(t, "test-key", r.URL.Query().Get("key"))
		require.Equal(t, "Shanghai Station", r.URL.Query().Get("address"))
		_, _ = w.Write([]byte(`{
			"status":"1",
			"info":"OK",
			"geocodes":[{
				"formatted_address":"Shanghai Station",
				"province":"Shanghai",
				"city":"Shanghai",
				"district":"Jing'an",
				"adcode":"310106",
				"location":"121.462000,31.249000"
			}]
		}`))
	}))
	defer srv.Close()

	client, err := NewClient(Config{WebKey: "test-key", BaseURL: srv.URL})
	require.NoError(t, err)

	point, err := client.Geocode(context.Background(), "Shanghai Station", "Shanghai")

	require.NoError(t, err)
	require.Equal(t, 31.249, point.Latitude)
	require.Equal(t, 121.462, point.Longitude)
	require.Equal(t, "Shanghai Station", point.FormattedAddress)
	require.Equal(t, "310106", point.AdCode)
}

func TestRouteByAddressGeocodesThenBuildsDrivingRoute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/geocode/geo":
			if r.URL.Query().Get("address") == "Origin" {
				_, _ = w.Write([]byte(`{"status":"1","geocodes":[{"formatted_address":"Origin","location":"121.400000,31.200000"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"status":"1","geocodes":[{"formatted_address":"Destination","location":"121.500000,31.300000"}]}`))
		case "/v3/direction/driving":
			require.Equal(t, "121.400000,31.200000", r.URL.Query().Get("origin"))
			require.Equal(t, "121.500000,31.300000", r.URL.Query().Get("destination"))
			_, _ = w.Write([]byte(`{
				"status":"1",
				"route":{
					"paths":[{
						"distance":"12345",
						"duration":"1800",
						"steps":[
							{"polyline":"121.400000,31.200000;121.450000,31.250000"},
							{"polyline":"121.450000,31.250000;121.500000,31.300000"}
						]
					}]
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client, err := NewClient(Config{WebKey: "test-key", BaseURL: srv.URL})
	require.NoError(t, err)

	route, origin, destination, err := client.RouteByAddress(context.Background(), "Origin", "Destination", "Shanghai", "")

	require.NoError(t, err)
	require.Equal(t, "Origin", origin.FormattedAddress)
	require.Equal(t, "Destination", destination.FormattedAddress)
	require.Equal(t, int64(12345), route.DistanceMeters)
	require.Equal(t, int64(1800), route.DurationSeconds)
	require.Len(t, route.Polyline, 3)
	require.Equal(t, 31.3, route.Polyline[2].Latitude)
}

func TestRegeoAndWeatherParseBaseResponses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/geocode/regeo":
			require.Equal(t, "121.473700,31.230400", r.URL.Query().Get("location"))
			_, _ = w.Write([]byte(`{
				"status":"1",
				"regeocode":{
					"formatted_address":"People Square",
					"addressComponent":{"province":"Shanghai","city":"Shanghai","district":"Huangpu","adcode":"310101"}
				}
			}`))
		case "/v3/weather/weatherInfo":
			_, _ = w.Write([]byte(`{
				"status":"1",
				"lives":[{
					"province":"Shanghai",
					"city":"Shanghai",
					"adcode":"310000",
					"weather":"Sunny",
					"temperature":"29",
					"winddirection":"East",
					"windpower":"3",
					"humidity":"61",
					"reporttime":"2026-08-04 10:00:00"
				}]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client, err := NewClient(Config{WebKey: "test-key", BaseURL: srv.URL})
	require.NoError(t, err)

	regeo, err := client.Regeo(context.Background(), 31.2304, 121.4737)
	require.NoError(t, err)
	require.Equal(t, "People Square", regeo.FormattedAddress)

	weather, err := client.Weather(context.Background(), "310000")
	require.NoError(t, err)
	require.Equal(t, "Sunny", weather.Weather)
	require.Equal(t, "2026-08-04 10:00:00", weather.ReportTime)
}
