package server

import (
	"encoding/json"
	"testing"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"

	"github.com/stretchr/testify/require"
)

func TestVehicleListResponseKeepsSnowflakeIDsAsStrings(t *testing.T) {
	reply := &driverv1.ListVehiclesReply{
		Items: []*driverv1.DriverVehicle{{
			Id:           2084245881267298304,
			DriverId:     3,
			AuditId:      2084245881267298304,
			PlateNo:      "京A12345",
			Brand:        "比亚迪",
			Model:        "秦PLUS",
			Color:        "白色",
			VehicleType:  "新能源",
			Seats:        5,
			Status:       1,
			ReviewStatus: 3,
			CanEdit:      true,
			CanDelete:    true,
			Source:       "vehicle",
			CreatedAt:    "2026-08-04 10:12:17",
			UpdatedAt:    "2026-08-04 10:12:17",
		}},
	}

	payload := safeVehicleListReply(reply)

	require.Len(t, payload.Items, 1)
	item := payload.Items[0]
	require.Equal(t, "2084245881267298304", item.ID)
	require.Equal(t, "3", item.DriverID)
	require.Equal(t, "2084245881267298304", item.AuditID)
	require.Equal(t, "京A12345", item.PlateNo)
	require.True(t, item.CanEdit)
	require.True(t, item.CanDelete)

	body, err := json.Marshal(payload)
	require.NoError(t, err)
	var encoded map[string]any
	require.NoError(t, json.Unmarshal(body, &encoded))
	items, ok := encoded["items"].([]any)
	require.True(t, ok)
	encodedItem, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "2084245881267298304", encodedItem["id"])
	require.Equal(t, "2084245881267298304", encodedItem["auditId"])
}

func TestVehicleReplyKeepsSnowflakeIDsAsStrings(t *testing.T) {
	reply := &driverv1.VehicleReply{
		Vehicle: &driverv1.DriverVehicle{
			Id:       2084245881267298304,
			DriverId: 3,
			AuditId:  2084245881267298304,
		},
	}

	payload := safeVehicleReply(reply)

	require.NotNil(t, payload.Vehicle)
	require.Equal(t, "2084245881267298304", payload.Vehicle.ID)
	require.Equal(t, "3", payload.Vehicle.DriverID)
	require.Equal(t, "2084245881267298304", payload.Vehicle.AuditID)
}

func TestParseInt64ParamAcceptsSnowflakeID(t *testing.T) {
	id, err := parseInt64Param("2084245881267298304")

	require.NoError(t, err)
	require.Equal(t, int64(2084245881267298304), id)
}
