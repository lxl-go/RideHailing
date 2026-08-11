package server

import (
	"strconv"

	driverv1 "ride-hailing/services/driver-service/api/driver/v1"
)

type vehicleListResponse struct {
	Items []vehicleResponse `json:"items"`
	List  []vehicleResponse `json:"list"`
}

type vehicleReplyResponse struct {
	Vehicle *vehicleResponse `json:"vehicle"`
}

type vehicleResponse struct {
	ID              string `json:"id"`
	DriverID        string `json:"driverId"`
	DriverIDAlt     string `json:"driver_id"`
	PlateNo         string `json:"plateNo"`
	PlateNoAlt      string `json:"plate_no"`
	Brand           string `json:"brand"`
	Model           string `json:"model"`
	Color           string `json:"color"`
	VehicleType     string `json:"vehicleType"`
	VehicleTypeAlt  string `json:"vehicle_type"`
	Seats           int32  `json:"seats"`
	Status          int32  `json:"status"`
	CreatedAt       string `json:"createdAt"`
	CreatedAtAlt    string `json:"created_at"`
	UpdatedAt       string `json:"updatedAt"`
	UpdatedAtAlt    string `json:"updated_at"`
	AuditID         string `json:"auditId"`
	AuditIDAlt      string `json:"audit_id"`
	ReviewStatus    int32  `json:"reviewStatus"`
	ReviewStatusAlt int32  `json:"review_status"`
	RejectReason    string `json:"rejectReason"`
	RejectReasonAlt string `json:"reject_reason"`
	Source          string `json:"source"`
	CanEdit         bool   `json:"canEdit"`
	CanEditAlt      bool   `json:"can_edit"`
	CanDelete       bool   `json:"canDelete"`
	CanDeleteAlt    bool   `json:"can_delete"`
}

func safeVehicleListReply(reply *driverv1.ListVehiclesReply) vehicleListResponse {
	if reply == nil {
		return vehicleListResponse{Items: []vehicleResponse{}, List: []vehicleResponse{}}
	}
	items := make([]vehicleResponse, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, safeVehicle(item))
	}
	return vehicleListResponse{Items: items, List: items}
}

func safeVehicleReply(reply *driverv1.VehicleReply) vehicleReplyResponse {
	if reply == nil || reply.Vehicle == nil {
		return vehicleReplyResponse{}
	}
	vehicle := safeVehicle(reply.Vehicle)
	return vehicleReplyResponse{Vehicle: &vehicle}
}

func safeVehicle(vehicle *driverv1.DriverVehicle) vehicleResponse {
	if vehicle == nil {
		return vehicleResponse{}
	}
	driverID := int64String(vehicle.DriverId)
	plateNo := vehicle.PlateNo
	vehicleType := vehicle.VehicleType
	createdAt := vehicle.CreatedAt
	updatedAt := vehicle.UpdatedAt
	auditID := int64String(vehicle.AuditId)
	rejectReason := vehicle.RejectReason
	return vehicleResponse{
		ID:              int64String(vehicle.Id),
		DriverID:        driverID,
		DriverIDAlt:     driverID,
		PlateNo:         plateNo,
		PlateNoAlt:      plateNo,
		Brand:           vehicle.Brand,
		Model:           vehicle.Model,
		Color:           vehicle.Color,
		VehicleType:     vehicleType,
		VehicleTypeAlt:  vehicleType,
		Seats:           vehicle.Seats,
		Status:          vehicle.Status,
		CreatedAt:       createdAt,
		CreatedAtAlt:    createdAt,
		UpdatedAt:       updatedAt,
		UpdatedAtAlt:    updatedAt,
		AuditID:         auditID,
		AuditIDAlt:      auditID,
		ReviewStatus:    vehicle.ReviewStatus,
		ReviewStatusAlt: vehicle.ReviewStatus,
		RejectReason:    rejectReason,
		RejectReasonAlt: rejectReason,
		Source:          vehicle.Source,
		CanEdit:         vehicle.CanEdit,
		CanEditAlt:      vehicle.CanEdit,
		CanDelete:       vehicle.CanDelete,
		CanDeleteAlt:    vehicle.CanDelete,
	}
}

func int64String(value int64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}
