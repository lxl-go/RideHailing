package carpool

import (
	"context"
	"encoding/json"
	"testing"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestShuttleServiceCreateListPublish(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&carpoolModel.ShuttleLine{}))
	global.GVA_DB = db

	service := ShuttleService{}
	ctx := context.Background()

	line, err := service.CreateLine(ctx, carpoolReq.ShuttleLinePayload{
		LineCode:    "S100",
		LineName:    "S100 Commuter Express",
		Route:       "A Station -> B Station -> C Station",
		DepartTime:  "08:00",
		ArriveTime:  "08:40",
		VehicleType: "Medium Bus",
		TotalSeats:  28,
		RemainSeats: 28,
		SortNo:      1,
	})
	require.NoError(t, err)
	require.NotZero(t, line.ID)

	list, total, err := service.ListLines(ctx, carpoolReq.ShuttleLineSearch{
		Keyword: "S100",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "S100", list[0].LineCode)
	require.EqualValues(t, 0, list[0].Status)

	err = service.PublishLines(ctx, []uint64{line.ID})
	require.NoError(t, err)

	detail, err := service.GetLine(ctx, line.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, detail.Status)
}

func TestShuttleServiceNormalizesStations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&carpoolModel.ShuttleLine{}))
	global.GVA_DB = db

	service := ShuttleService{}
	ctx := context.Background()

	line, err := service.CreateLine(ctx, carpoolReq.ShuttleLinePayload{
		LineCode: "S200",
		LineName: "Airport Shuttle",
		Stations: []carpoolReq.ShuttleStation{
			{Name: "Terminal", OffsetMin: 30, Time: "08:30"},
			{Name: "Origin", OffsetMin: 0, Time: "08:00"},
			{Name: "terminal", OffsetMin: 30, Time: "08:30"},
			{Name: "Middle", OffsetMin: 10, Time: "08:10"},
		},
	})
	require.NoError(t, err)

	var stations []carpoolReq.ShuttleStation
	require.NoError(t, json.Unmarshal([]byte(line.Stations), &stations))
	require.Len(t, stations, 3)
	require.Equal(t, "Origin", stations[0].Name)
	require.Equal(t, 0, stations[0].OffsetMin)
	require.Equal(t, "Middle", stations[1].Name)
	require.Equal(t, 10, stations[1].OffsetMin)
	require.Equal(t, "Terminal", stations[2].Name)
	require.Equal(t, 30, stations[2].OffsetMin)
}

func TestShuttleServiceRejectsInvalidStations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&carpoolModel.ShuttleLine{}))
	global.GVA_DB = db

	service := ShuttleService{}
	ctx := context.Background()

	_, err = service.CreateLine(ctx, carpoolReq.ShuttleLinePayload{
		LineCode: "S201",
		LineName: "Invalid Shuttle",
		Stations: []carpoolReq.ShuttleStation{
			{Name: "Middle", OffsetMin: 10},
			{Name: "Terminal", OffsetMin: 30},
		},
	})
	require.ErrorContains(t, err, "first station offsetMin must be 0")

	_, err = service.CreateLine(ctx, carpoolReq.ShuttleLinePayload{
		LineCode: "S202",
		LineName: "Invalid Shuttle",
		Stations: []carpoolReq.ShuttleStation{
			{Name: "Origin", OffsetMin: 0},
			{Name: " ", OffsetMin: 10},
		},
	})
	require.ErrorContains(t, err, "station name is required")

	_, err = service.CreateLine(ctx, carpoolReq.ShuttleLinePayload{
		LineCode: "S203",
		LineName: "Invalid Shuttle",
		Stations: []carpoolReq.ShuttleStation{
			{Name: "Origin", OffsetMin: 0},
			{Name: "Back", OffsetMin: -1},
		},
	})
	require.ErrorContains(t, err, "station offsetMin cannot be negative")
}
