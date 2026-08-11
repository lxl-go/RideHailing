package response

import (
	"strconv"
	"time"

	"ride-hailing/admin-server/model/carpool"
)

type TripResponse struct {
	ID                   string     `json:"id"`
	PublisherID          string     `json:"publisherId"`
	PublisherRole        int        `json:"publisherRole"`
	TripType             int        `json:"tripType"`
	OriginName           string     `json:"originName"`
	OriginLat            float64    `json:"originLat"`
	OriginLng            float64    `json:"originLng"`
	DestName             string     `json:"destName"`
	DestLat              float64    `json:"destLat"`
	DestLng              float64    `json:"destLng"`
	DepartureTime        time.Time  `json:"departureTime"`
	ArriveTime           time.Time  `json:"arriveTime"`
	SeatsTotal           int        `json:"seatsTotal"`
	SeatsAvailable       int        `json:"seatsAvailable"`
	ShareCost            float64    `json:"shareCost"`
	BaggageInfo          string     `json:"baggageInfo"`
	PetAllowed           int        `json:"petAllowed"`
	Remarks              string     `json:"remarks"`
	Status               int        `json:"status"`
	RejectReason         string     `json:"rejectReason"`
	AuditOperatorID      string     `json:"auditOperatorId"`
	AuditTime            *time.Time `json:"auditTime"`
	RouteDistanceMeters  int        `json:"routeDistanceMeters"`
	RouteDurationSeconds int        `json:"routeDurationSeconds"`
	IsDeleted            bool       `json:"isDeleted"`
	MatchedOrderID       string     `json:"matchedOrderId"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func NewTripResponse(trip carpool.Trip) TripResponse {
	return TripResponse{
		ID:                   strconv.FormatInt(trip.ID, 10),
		PublisherID:          strconv.FormatInt(trip.PublisherID, 10),
		PublisherRole:        trip.PublisherRole,
		TripType:             trip.TripType,
		OriginName:           trip.OriginName,
		OriginLat:            trip.OriginLat,
		OriginLng:            trip.OriginLng,
		DestName:             trip.DestName,
		DestLat:              trip.DestLat,
		DestLng:              trip.DestLng,
		DepartureTime:        trip.DepartureTime,
		ArriveTime:           trip.ArriveTime,
		SeatsTotal:           trip.SeatsTotal,
		SeatsAvailable:       trip.SeatsAvailable,
		ShareCost:            trip.ShareCost,
		BaggageInfo:          trip.BaggageInfo,
		PetAllowed:           trip.PetAllowed,
		Remarks:              trip.Remarks,
		Status:               trip.Status,
		RejectReason:         trip.RejectReason,
		AuditOperatorID:      strconv.FormatInt(trip.AuditOperatorID, 10),
		AuditTime:            trip.AuditTime,
		RouteDistanceMeters:  trip.RouteDistanceMeters,
		RouteDurationSeconds: trip.RouteDurationSeconds,
		IsDeleted:            trip.IsDeleted,
		MatchedOrderID:       strconv.FormatInt(trip.MatchedOrderID, 10),
		CreatedAt:            trip.CreatedAt,
		UpdatedAt:            trip.UpdatedAt,
	}
}

func NewTripResponses(items []carpool.Trip) []TripResponse {
	responses := make([]TripResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewTripResponse(item))
	}
	return responses
}
