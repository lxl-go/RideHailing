package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	orderv1 "ride-hailing/services/order-service/api/order/v1"
)

func TestMobileOrderItemResponseUsesFrontendSafeIDAndMoney(t *testing.T) {
	item := &orderv1.OrderItem{
		Id:          9007199254740993,
		TripId:      9007199254740995,
		PassengerId: 9007199254740997,
		DriverId:    9007199254740999,
		TotalPrice:  39.8,
		Status:      0,
	}

	got := mobileOrderItemResponse(item)

	require.Equal(t, "9007199254740993", got["id"])
	require.Equal(t, "9007199254740995", got["tripId"])
	require.Equal(t, "9007199254740997", got["passengerId"])
	require.Equal(t, "9007199254740999", got["driverId"])
	require.Equal(t, "39.80", got["totalPriceText"])
	require.Equal(t, "39.80", got["total_price_text"])
	require.Equal(t, int64(3980), got["totalPriceCents"])
	require.Equal(t, int64(3980), got["total_price_cents"])
}

func TestParseMobileOrderStatusDoesNotMixPaymentAndFulfillmentStates(t *testing.T) {
	require.Equal(t, int32(0), parseMobileOrderStatus("pending"))
	require.Equal(t, int32(1), parseMobileOrderStatus("accepted"))
	require.Equal(t, int32(5), parseMobileOrderStatus("picking_up"))
	require.Equal(t, int32(6), parseMobileOrderStatus("delivering"))
	require.Equal(t, int32(4), parseMobileOrderStatus("paid"))
	require.Equal(t, int32(-1), parseMobileOrderStatus("in_progress"))
}

func TestMobileOrderStatusIncludesDriverFulfillmentStates(t *testing.T) {
	require.Equal(t, "picking_up", mobileOrderStatus(5))
	require.Equal(t, "delivering", mobileOrderStatus(6))
}
