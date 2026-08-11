package server

import (
	"context"
	"net/http"
	"testing"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/require"

	"ride-hailing/pkg/amapx"
)

type fakeAmapRouteClient struct {
	geocodeReq struct{ address, city string }
	regeoReq   struct{ lat, lng float64 }
	routeReq   struct{ origin, destination, city, strategy string }
	weatherReq string
	staticReq  amapx.StaticMapRequest
	staticErr  error
}

func (f *fakeAmapRouteClient) Geocode(_ context.Context, address, city string) (*amapx.Location, error) {
	f.geocodeReq.address = address
	f.geocodeReq.city = city
	return &amapx.Location{Latitude: 31.249, Longitude: 121.462, FormattedAddress: address}, nil
}

func (f *fakeAmapRouteClient) Regeo(_ context.Context, lat, lng float64) (*amapx.Location, error) {
	f.regeoReq.lat = lat
	f.regeoReq.lng = lng
	return &amapx.Location{Latitude: lat, Longitude: lng, FormattedAddress: "People Square"}, nil
}

func (f *fakeAmapRouteClient) RouteByAddress(_ context.Context, originAddress, destinationAddress, city, strategy string) (*amapx.Route, *amapx.Location, *amapx.Location, error) {
	f.routeReq.origin = originAddress
	f.routeReq.destination = destinationAddress
	f.routeReq.city = city
	f.routeReq.strategy = strategy
	route := &amapx.Route{
		Origin:          amapx.Location{Latitude: 31.200, Longitude: 121.400, FormattedAddress: originAddress},
		Destination:     amapx.Location{Latitude: 31.300, Longitude: 121.500, FormattedAddress: destinationAddress},
		DistanceMeters:  12345,
		DurationSeconds: 1800,
		Polyline: []amapx.RoutePoint{
			{Latitude: 31.200, Longitude: 121.400},
			{Latitude: 31.250, Longitude: 121.450},
		},
	}
	return route, &route.Origin, &route.Destination, nil
}

func (f *fakeAmapRouteClient) Weather(_ context.Context, city string) (*amapx.Weather, error) {
	f.weatherReq = city
	return &amapx.Weather{City: city, Weather: "Sunny", Temperature: "29"}, nil
}

func (f *fakeAmapRouteClient) StaticMap(_ context.Context, req amapx.StaticMapRequest) (*amapx.StaticMap, error) {
	f.staticReq = req
	if f.staticErr != nil {
		return nil, f.staticErr
	}
	return &amapx.StaticMap{ContentType: "image/png", Data: []byte{0x89, 0x50, 0x4e, 0x47}}, nil
}

func TestMapRoutesProxyAmapAPIs(t *testing.T) {
	srv := khttp.NewServer()
	fake := &fakeAmapRouteClient{}
	registerMapRoutesWithClient(srv, fake)

	res := doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/geocode?address=Shanghai+Station&city=Shanghai", "")
	require.Equal(t, http.StatusOK, res.Code)
	payload := decodeGatewayBody(t, res)
	require.Equal(t, float64(0), payload["code"])
	require.Equal(t, "Shanghai Station", payload["data"].(map[string]any)["formattedAddress"])
	require.Equal(t, "Shanghai Station", fake.geocodeReq.address)

	res = doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/regeo?lat=31.2304&lng=121.4737", "")
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "People Square", decodeGatewayBody(t, res)["data"].(map[string]any)["formattedAddress"])

	res = doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/route?origin=Origin&destination=Destination&city=Shanghai", "")
	require.Equal(t, http.StatusOK, res.Code)
	route := decodeGatewayBody(t, res)["data"].(map[string]any)
	require.Equal(t, float64(12345), route["distanceMeters"])
	require.Len(t, route["polyline"].([]any), 2)

	res = doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/weather?city=Shanghai", "")
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "Sunny", decodeGatewayBody(t, res)["data"].(map[string]any)["weather"])
	require.Equal(t, "Shanghai", fake.weatherReq)
}

func TestMapRoutesRejectBadCoordinates(t *testing.T) {
	srv := khttp.NewServer()
	registerMapRoutesWithClient(srv, &fakeAmapRouteClient{})

	res := doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/regeo?lat=abc&lng=121.4737", "")

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestMapRoutesStaticMapProxy(t *testing.T) {
	srv := khttp.NewServer()
	fake := &fakeAmapRouteClient{}
	registerMapRoutesWithClient(srv, fake)

	res := doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/static?origin=116.378517,39.865246&destination=116.321212,39.894582&size=750*400&zoom=12&scale=2", "")

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "image/png", res.Header().Get("Content-Type"))
	require.Len(t, res.Body.Bytes(), 4)
	require.Len(t, fake.staticReq.Markers, 2)
	require.Equal(t, "起", fake.staticReq.Markers[0].Label)
	require.Equal(t, "终", fake.staticReq.Markers[1].Label)
	require.Equal(t, 12, fake.staticReq.Zoom)
	require.Equal(t, "750*400", fake.staticReq.Size)
}

func TestMapRoutesStaticMapRejectsInvalidCoords(t *testing.T) {
	srv := khttp.NewServer()
	registerMapRoutesWithClient(srv, &fakeAmapRouteClient{})

	res := doGatewayJSON(srv, http.MethodGet, "/api/v1/maps/static?origin=abc&destination=116.321212,39.894582", "")

	require.Equal(t, http.StatusBadRequest, res.Code)
}
