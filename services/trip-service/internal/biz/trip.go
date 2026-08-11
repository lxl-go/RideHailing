package biz

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

const (
	TripStatusPending    = 10
	TripStatusApproved   = 20
	TripStatusRejected   = 30
	TripStatusCancelled  = 5
	TripStatusRecruiting = TripStatusPending
)

const tripMinimumLeadTime = 15 * time.Minute
const tripConflictBuffer = 30 * time.Minute

const (
	DemandStatusPending = iota + 1
	DemandStatusCancelled
)

const CouponStatusUnused = "unused"

type Trip struct {
	ID                   int64
	DriverID             int64
	PublisherID          int64
	PublisherRole        int
	TripType             int
	Origin               string
	OriginName           string
	OriginLat            float64
	OriginLng            float64
	Destination          string
	DestName             string
	DestLat              float64
	DestLng              float64
	DepartTime           time.Time
	ArriveTime           time.Time
	DepartureTime        time.Time
	SeatsTotal           int
	SeatsAvailable       int
	Price                float64
	ShareCost            float64
	Status               int
	RejectReason         string
	AuditOperatorID      int64
	AuditTime            *time.Time
	RouteDistanceMeters  int
	RouteDurationSeconds int
	IsDeleted            bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Coupon struct {
	ID              int64
	CouponNo        string
	CouponCode      string
	Name            string
	CouponType      string
	FaceValue       float64
	DiscountRate    float64
	ThresholdAmount float64
	ValidFrom       time.Time
	ValidTo         time.Time
	Status          string
	Claimed         bool
}

type TripDemand struct {
	ID          int64
	PassengerID int64
	Origin      string
	Destination string
	DepartTime  time.Time
	Seats       int
	Budget      float64
	Remark      string
	Status      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PublishTripCommand struct {
	RequestID            string
	DriverID             int64
	Origin               string
	OriginName           string
	OriginLat            float64
	OriginLng            float64
	Destination          string
	DestName             string
	DestLat              float64
	DestLng              float64
	DepartTime           time.Time
	ArriveTime           time.Time
	SeatsTotal           int
	Price                float64
	RouteDistanceMeters  int
	RouteDurationSeconds int
}

type ClaimCouponCommand struct {
	PassengerID    int64
	CouponNo       string
	IdempotencyKey string
}

type PublishDemandCommand struct {
	PassengerID int64
	Origin      string
	Destination string
	DepartTime  time.Time
	Seats       int
	Budget      float64
	Remark      string
}

type DeleteTripCommand struct {
	TripID   int64
	DriverID int64
}

type TripUsecase struct {
	node *snowflake.Node
	log  *zap.Logger
	repo TripRepo
}

func NewTripUsecase(node *snowflake.Node, log *zap.Logger, repo TripRepo) *TripUsecase {
	return &TripUsecase{node: node, log: log, repo: repo}
}

func (uc *TripUsecase) SearchTrips(ctx context.Context, origin, destination, departDate string, page, pageSize int) ([]Trip, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.SearchTrips(ctx, strings.TrimSpace(origin), strings.TrimSpace(destination), strings.TrimSpace(departDate), page, pageSize)
}

func (uc *TripUsecase) GetTripDetail(ctx context.Context, id int64) (*Trip, error) {
	if id <= 0 {
		return nil, ErrInvalidTrip
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *TripUsecase) PublishTrip(ctx context.Context, cmd PublishTripCommand) (*Trip, error) {
	acquirer, supportsIdempotency := uc.repo.(interface {
		AcquirePublishRequest(context.Context, int64, string) (bool, error)
		ReleasePublishRequest(context.Context, int64, string)
	})
	if supportsIdempotency && strings.TrimSpace(cmd.RequestID) == "" {
		return nil, ErrInvalidTrip
	}
	resultCache, supportsResultCache := uc.repo.(interface {
		GetPublishedTripResult(context.Context, int64, string) (*Trip, error)
		SavePublishedTripResult(context.Context, int64, string, *Trip) error
	})
	if supportsResultCache && strings.TrimSpace(cmd.RequestID) != "" {
		if cached, err := resultCache.GetPublishedTripResult(ctx, cmd.DriverID, cmd.RequestID); err != nil {
			return nil, err
		} else if cached != nil {
			return cached, nil
		}
	}
	acquiredRequest := false
	if supportsIdempotency {
		acquired, err := acquirer.AcquirePublishRequest(ctx, cmd.DriverID, cmd.RequestID)
		if err != nil {
			return nil, err
		}
		if !acquired {
			if supportsResultCache {
				if cached, cacheErr := resultCache.GetPublishedTripResult(ctx, cmd.DriverID, cmd.RequestID); cacheErr != nil {
					return nil, cacheErr
				} else if cached != nil {
					return cached, nil
				}
			}
			return nil, ErrDuplicateTripRequest
		}
		acquiredRequest = true
		defer acquirer.ReleasePublishRequest(ctx, cmd.DriverID, cmd.RequestID)
	}
	if cmd.DriverID <= 0 || strings.TrimSpace(cmd.Origin) == "" || strings.TrimSpace(cmd.Destination) == "" || cmd.SeatsTotal < 1 || cmd.SeatsTotal > 6 || cmd.DepartTime.Before(time.Now().Add(tripMinimumLeadTime)) {
		if acquiredRequest {
			acquirer.ReleasePublishRequest(ctx, cmd.DriverID, cmd.RequestID)
		}
		return nil, ErrInvalidTrip
	}
	if conflictRepo, ok := uc.repo.(interface {
		HasApprovedTimeConflict(context.Context, int64, time.Time) (bool, error)
	}); ok {
		conflict, err := conflictRepo.HasApprovedTimeConflict(ctx, cmd.DriverID, cmd.DepartTime)
		if err != nil {
			return nil, err
		}
		if conflict {
			return nil, ErrTripTimeConflict
		}
	}
	origin := firstNonEmpty(cmd.OriginName, cmd.Origin)
	destination := firstNonEmpty(cmd.DestName, cmd.Destination)
	trip := &Trip{
		ID:                   uc.node.Generate().Int64(),
		DriverID:             cmd.DriverID,
		PublisherID:          cmd.DriverID,
		PublisherRole:        1,
		TripType:             1,
		Origin:               origin,
		OriginName:           origin,
		OriginLat:            cmd.OriginLat,
		OriginLng:            cmd.OriginLng,
		Destination:          destination,
		DestName:             destination,
		DestLat:              cmd.DestLat,
		DestLng:              cmd.DestLng,
		DepartTime:           cmd.DepartTime,
		ArriveTime:           cmd.ArriveTime,
		DepartureTime:        cmd.DepartTime,
		SeatsTotal:           cmd.SeatsTotal,
		SeatsAvailable:       cmd.SeatsTotal,
		Price:                cmd.Price,
		ShareCost:            cmd.Price,
		RouteDistanceMeters:  cmd.RouteDistanceMeters,
		RouteDurationSeconds: cmd.RouteDurationSeconds,
		Status:               TripStatusPending,
	}
	if err := uc.repo.Create(ctx, trip); err != nil {
		uc.log.Error("publish trip failed", zap.Error(err))
		if acquiredRequest {
			acquirer.ReleasePublishRequest(ctx, cmd.DriverID, cmd.RequestID)
		}
		return nil, err
	}
	if supportsResultCache && strings.TrimSpace(cmd.RequestID) != "" {
		if err := resultCache.SavePublishedTripResult(ctx, cmd.DriverID, cmd.RequestID, trip); err != nil {
			return nil, err
		}
	}
	return trip, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (uc *TripUsecase) ListDriverTrips(ctx context.Context, driverID int64, status int, page, pageSize int) ([]Trip, int64, error) {
	if driverID <= 0 {
		return nil, 0, ErrInvalidTrip
	}
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.ListByDriver(ctx, driverID, status, page, pageSize)
}

func (uc *TripUsecase) UpdateTripStatus(ctx context.Context, id int64, status int) error {
	if id <= 0 || (status != TripStatusApproved && status != TripStatusRejected && status != TripStatusCancelled) {
		return ErrInvalidTrip
	}
	return uc.repo.UpdateStatus(ctx, id, status)
}

func (uc *TripUsecase) ListCoupons(ctx context.Context, passengerID int64, page, pageSize int) ([]Coupon, int64, error) {
	if passengerID <= 0 {
		return nil, 0, ErrInvalidCoupon
	}
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.ListCoupons(ctx, passengerID, page, pageSize)
}

func (uc *TripUsecase) ClaimCoupon(ctx context.Context, cmd ClaimCouponCommand) (*Coupon, bool, error) {
	if cmd.PassengerID <= 0 || strings.TrimSpace(cmd.CouponNo) == "" {
		return nil, false, ErrInvalidCoupon
	}
	return uc.repo.ClaimCoupon(ctx, cmd.PassengerID, strings.TrimSpace(cmd.CouponNo), strings.TrimSpace(cmd.IdempotencyKey))
}

func (uc *TripUsecase) PublishDemand(ctx context.Context, cmd PublishDemandCommand) (*TripDemand, error) {
	if cmd.PassengerID <= 0 || strings.TrimSpace(cmd.Origin) == "" || strings.TrimSpace(cmd.Destination) == "" || cmd.DepartTime.IsZero() || cmd.Seats <= 0 {
		return nil, ErrInvalidDemand
	}
	demand := &TripDemand{
		ID:          uc.node.Generate().Int64(),
		PassengerID: cmd.PassengerID,
		Origin:      strings.TrimSpace(cmd.Origin),
		Destination: strings.TrimSpace(cmd.Destination),
		DepartTime:  cmd.DepartTime,
		Seats:       cmd.Seats,
		Budget:      cmd.Budget,
		Remark:      strings.TrimSpace(cmd.Remark),
		Status:      DemandStatusPending,
		CreatedAt:   time.Now(),
	}
	if err := uc.repo.CreateDemand(ctx, demand); err != nil {
		uc.log.Error("publish demand failed", zap.Error(err))
		return nil, err
	}
	return demand, nil
}

func (uc *TripUsecase) ListMyDemands(ctx context.Context, passengerID int64, status int, page, pageSize int) ([]TripDemand, int64, error) {
	if passengerID <= 0 {
		return nil, 0, ErrInvalidDemand
	}
	page, pageSize = normalizePage(page, pageSize)
	return uc.repo.ListDemandsByPassenger(ctx, passengerID, status, page, pageSize)
}

func (uc *TripUsecase) CancelDemand(ctx context.Context, id int64, passengerID int64) error {
	if id <= 0 || passengerID <= 0 {
		return ErrInvalidDemand
	}
	return uc.repo.CancelDemand(ctx, id, passengerID)
}

func (uc *TripUsecase) DeleteDriverTrip(ctx context.Context, cmd DeleteTripCommand) error {
	if cmd.TripID <= 0 || cmd.DriverID <= 0 {
		return ErrInvalidTrip
	}
	return uc.repo.DeleteDriverTrip(ctx, cmd.TripID, cmd.DriverID)
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
