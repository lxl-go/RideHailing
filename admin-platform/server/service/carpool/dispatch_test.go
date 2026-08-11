package carpool

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

func newDispatchServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&carpoolModel.OrderMain{},
		&testCarpoolTrip{},
		&testCarpoolOrder{},
		&testCarpoolPayment{},
		&carpoolModel.OrderStatusHistory{},
		&carpoolModel.OrderDispatchAudit{},
		&carpoolModel.DispatchConfig{},
		&carpoolModel.DispatchConfigVersion{},
		&carpoolModel.DriverLocationPoint{},
		&carpoolModel.RealtimeMessage{},
	))
	global.GVA_DB = db
	return db
}

func createDispatchRealOrder(t *testing.T, db *gorm.DB, id uint64, depart time.Time, status string) string {
	t.Helper()
	tripID := id + 1000
	require.NoError(t, db.Create(&testCarpoolTrip{
		ID: tripID, DriverID: 2000 + id%1000, StartLocation: "A", EndLocation: "B",
		DepartureTime: depart, AvailableSeats: 2, PricePerSeat: 80, Status: "published",
	}).Error)
	require.NoError(t, db.Create(&testCarpoolOrder{
		ID: id, TripID: tripID, PassengerID: 1000 + id%1000, SeatsBooked: 1,
		TotalPrice: 80, Status: status, CreatedAt: depart.Add(-time.Hour), UpdatedAt: depart.Add(-time.Hour),
	}).Error)
	return strconv.FormatUint(id, 10)
}

func TestDispatchServiceListOrdersSortsEarlierScheduleFirst(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	now := time.Date(2026, 7, 30, 9, 0, 0, 0, time.Local)
	lateOrderID := createDispatchRealOrder(t, db, 6102, now.Add(2*time.Hour), "paid")
	earlyOrderID := createDispatchRealOrder(t, db, 6101, now.Add(30*time.Minute), "paid")
	service := DispatchService{}

	list, total, err := service.ListOrders(context.Background(), carpoolReq.DispatchOrderSearch{Status: "paid"})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Equal(t, earlyOrderID, list[0].OrderNo)
	require.Equal(t, lateOrderID, list[1].OrderNo)
}

func TestDispatchServiceScoresRealCarpoolOrderWithStringIDs(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&testCarpoolTrip{},
		&testCarpoolOrder{},
		&testCarpoolPayment{},
	))
	depart := time.Date(2026, 8, 4, 12, 0, 0, 0, time.Local)
	bigOrderID := "9007199254740993"
	bigDriverID := "9007199254740995"
	require.NoError(t, db.Create(&testCarpoolTrip{
		ID: 7301, DriverID: 9007199254740996, StartLocation: "Hongqiao", EndLocation: "Pudong",
		DepartureTime: depart, AvailableSeats: 2, PricePerSeat: 88, Status: "published",
	}).Error)
	require.NoError(t, db.Create(&testCarpoolOrder{
		ID: 9007199254740993, TripID: 7301, PassengerID: 9007199254740994,
		SeatsBooked: 1, TotalPrice: 88, Status: "paid", CreatedAt: depart.Add(-time.Hour), UpdatedAt: depart.Add(-time.Hour),
	}).Error)

	service := DispatchService{}
	decision, err := service.ScoreDrivers(context.Background(), carpoolReq.DispatchScoreRequest{
		OrderID: bigOrderID,
		City:    "Shanghai",
		Candidates: []carpoolReq.DriverCandidate{
			{DriverID: bigDriverID, DriverName: "Large ID Driver", City: "Shanghai", Online: true, DistanceKM: 2, Rating: 4.9, IdleMinutes: 30},
		},
		IdempotencyKey: "dispatch-large-id",
	})
	require.NoError(t, err)
	require.Equal(t, bigOrderID, decision.OrderID)
	require.Equal(t, bigDriverID, decision.SelectedDriverID)
}

func TestDispatchServiceExcludesDriverWithOverlappingServiceWindow(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	depart := time.Date(2026, 7, 30, 10, 0, 0, 0, time.Local)
	orderID := createDispatchRealOrder(t, db, 6201, depart, "paid")
	service := DispatchService{}

	decision, err := service.ScoreDrivers(context.Background(), carpoolReq.DispatchScoreRequest{
		OrderID: orderID,
		City:    "Shanghai",
		Candidates: []carpoolReq.DriverCandidate{
			{DriverID: "2001", DriverName: "Busy Driver", City: "Shanghai", Online: true, DistanceKM: 1, Rating: 5, IdleMinutes: 50, ServiceWindows: []carpoolReq.ServiceWindow{{Start: depart.Add(-30 * time.Minute), End: depart.Add(30 * time.Minute)}}},
			{DriverID: "2002", DriverName: "Free Driver", City: "Shanghai", Online: true, DistanceKM: 3, Rating: 4.8, IdleMinutes: 40},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "2002", decision.SelectedDriverID)
	require.Contains(t, decision.ExcludedReasons["2001"], "service window overlaps")
}

func TestDispatchServiceDayNightWeightSelectionChangesScore(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	service := DispatchService{}
	dayDepart := time.Date(2026, 7, 30, 10, 0, 0, 0, time.Local)
	nightDepart := time.Date(2026, 7, 30, 23, 0, 0, 0, time.Local)
	dayOrderID := createDispatchRealOrder(t, db, 6301, dayDepart, "paid")
	nightOrderID := createDispatchRealOrder(t, db, 6302, nightDepart, "paid")

	candidates := []carpoolReq.DriverCandidate{
		{DriverID: "2001", DriverName: "Near Driver", City: "Shanghai", Online: true, DistanceKM: 1, Rating: 4.2, IdleMinutes: 20},
		{DriverID: "2002", DriverName: "High Rating Driver", City: "Shanghai", Online: true, DistanceKM: 7, Rating: 5, IdleMinutes: 20},
	}
	dayDecision, err := service.ScoreDrivers(context.Background(), carpoolReq.DispatchScoreRequest{OrderID: dayOrderID, City: "Shanghai", Candidates: candidates})
	require.NoError(t, err)
	nightDecision, err := service.ScoreDrivers(context.Background(), carpoolReq.DispatchScoreRequest{OrderID: nightOrderID, City: "Shanghai", Candidates: candidates})
	require.NoError(t, err)

	require.Equal(t, "2001", dayDecision.SelectedDriverID)
	require.Equal(t, "2002", nightDecision.SelectedDriverID)
	require.Contains(t, dayDecision.ScoreDetail, "distance")
	require.Contains(t, nightDecision.ScoreDetail, "rating")
}

func TestDispatchServiceCandidatePoolFiltersByCityFleetAndHotZone(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	depart := time.Date(2026, 7, 30, 11, 0, 0, 0, time.Local)
	orderID := createDispatchRealOrder(t, db, 6401, depart, "paid")
	service := DispatchService{}

	decision, err := service.ScoreDrivers(context.Background(), carpoolReq.DispatchScoreRequest{
		OrderID: orderID,
		City:    "Shanghai",
		FleetID: "fleet-a",
		HotZone: "hongqiao",
		Candidates: []carpoolReq.DriverCandidate{
			{DriverID: "2001", City: "Hangzhou", FleetID: "fleet-a", HotZone: "hongqiao", Online: true, DistanceKM: 1, Rating: 5, IdleMinutes: 30},
			{DriverID: "2002", City: "Shanghai", FleetID: "fleet-b", HotZone: "hongqiao", Online: true, DistanceKM: 1, Rating: 5, IdleMinutes: 30},
			{DriverID: "2003", City: "Shanghai", FleetID: "fleet-a", HotZone: "pudong", Online: true, DistanceKM: 1, Rating: 5, IdleMinutes: 30},
			{DriverID: "2004", City: "Shanghai", FleetID: "fleet-a", HotZone: "hongqiao", Online: true, DistanceKM: 4, Rating: 4.9, IdleMinutes: 30},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "2004", decision.SelectedDriverID)
	require.Len(t, decision.Candidates, 1)
	require.Contains(t, decision.ExcludedReasons["2001"], "city mismatch")
	require.Contains(t, decision.ExcludedReasons["2002"], "fleet mismatch")
	require.Contains(t, decision.ExcludedReasons["2003"], "hot zone mismatch")
}

func TestDispatchServiceSelectsHighestScoreAndPersistsScoreDetail(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	depart := time.Date(2026, 7, 30, 14, 0, 0, 0, time.Local)
	orderID := createDispatchRealOrder(t, db, 6501, depart, "paid")
	service := DispatchService{}

	decision, err := service.ScoreDrivers(context.Background(), carpoolReq.DispatchScoreRequest{
		OrderID: orderID,
		City:    "Shanghai",
		Candidates: []carpoolReq.DriverCandidate{
			{DriverID: "2001", City: "Shanghai", Online: true, DistanceKM: 9, Rating: 4.4, IdleMinutes: 5},
			{DriverID: "2002", City: "Shanghai", Online: true, DistanceKM: 2, Rating: 4.9, IdleMinutes: 60},
		},
		IdempotencyKey: "score-once",
	})
	require.NoError(t, err)
	require.Equal(t, "2002", decision.SelectedDriverID)
	require.Greater(t, decision.Score, 0.0)
	require.Contains(t, decision.ScoreDetail, "idle")

	var audit carpoolModel.OrderDispatchAudit
	require.NoError(t, db.Where("idempotency_key = ?", "score-once").First(&audit).Error)
	require.Equal(t, orderID, audit.OrderNo)
	require.EqualValues(t, 2002, audit.DriverID)
	require.Contains(t, audit.ScoreDetail, "distance")
}

func TestDispatchServiceDuplicateIdempotencyKeyDoesNotCreateDuplicateAudit(t *testing.T) {
	db := newDispatchServiceTestDB(t)
	depart := time.Date(2026, 7, 30, 15, 0, 0, 0, time.Local)
	orderID := createDispatchRealOrder(t, db, 6601, depart, "paid")
	service := DispatchService{}
	req := carpoolReq.DispatchScoreRequest{
		OrderID: orderID,
		City:    "Shanghai",
		Candidates: []carpoolReq.DriverCandidate{
			{DriverID: "2001", City: "Shanghai", Online: true, DistanceKM: 2, Rating: 4.8, IdleMinutes: 30},
		},
		IdempotencyKey: "dispatch-idem-001",
	}

	first, err := service.ScoreDrivers(context.Background(), req)
	require.NoError(t, err)
	second, err := service.ScoreDrivers(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, first.AuditNo, second.AuditNo)

	var count int64
	require.NoError(t, db.Model(&carpoolModel.OrderDispatchAudit{}).Where("idempotency_key = ?", "dispatch-idem-001").Count(&count).Error)
	require.EqualValues(t, 1, count)
}
