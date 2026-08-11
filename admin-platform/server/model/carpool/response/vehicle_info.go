package response

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"ride-hailing/admin-server/model/carpool"
)

type VehicleInfoResponse struct {
	ID              string         `json:"id"`
	DriverID        string         `json:"driverId"`
	PlateNumber     string         `json:"plateNumber"`
	Brand           string         `json:"brand"`
	Model           string         `json:"model"`
	Color           string         `json:"color"`
	Seats           int            `json:"seats"`
	YearCheckDate   *time.Time     `json:"yearCheckDate"`
	InsuranceExpire *time.Time     `json:"insuranceExpire"`
	Status          int            `json:"status"`
	ReviewerID      string         `json:"reviewerId"`
	RejectReason    string         `json:"rejectReason"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `json:"-"`
}

func NewVehicleInfoResponse(info carpool.VehicleInfo) VehicleInfoResponse {
	return VehicleInfoResponse{
		ID:              strconv.FormatInt(info.ID, 10),
		DriverID:        strconv.FormatInt(info.DriverID, 10),
		PlateNumber:     info.PlateNumber,
		Brand:           info.Brand,
		Model:           info.Model,
		Color:           info.Color,
		Seats:           info.Seats,
		YearCheckDate:   info.YearCheckDate,
		InsuranceExpire: info.InsuranceExpire,
		Status:          info.Status,
		ReviewerID:      strconv.FormatInt(info.ReviewerID, 10),
		RejectReason:    info.RejectReason,
		CreatedAt:       info.CreatedAt,
		UpdatedAt:       info.UpdatedAt,
		DeletedAt:       info.DeletedAt,
	}
}

func NewVehicleInfoResponses(items []carpool.VehicleInfo) []VehicleInfoResponse {
	responses := make([]VehicleInfoResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewVehicleInfoResponse(item))
	}
	return responses
}
