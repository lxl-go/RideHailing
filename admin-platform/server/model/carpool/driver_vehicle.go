package carpool

import "time"

type DriverVehicle struct {
	ID          int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID    int64     `json:"driverId" gorm:"column:driver_id;type:bigint;not null;index:idx_driver_vehicle"`
	PlateNo     string    `json:"plateNo" gorm:"column:plate_no;type:varchar(32);not null;default:''"`
	Brand       string    `json:"brand" gorm:"column:brand;type:varchar(64);not null;default:''"`
	Model       string    `json:"model" gorm:"column:model;type:varchar(64);not null;default:''"`
	Color       string    `json:"color" gorm:"column:color;type:varchar(32);not null;default:''"`
	VehicleType string    `json:"vehicleType" gorm:"column:vehicle_type;type:varchar(32);not null;default:''"`
	Seats       int       `json:"seats" gorm:"column:seats;type:int;not null;default:4"`
	Status      int       `json:"status" gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (DriverVehicle) TableName() string {
	return "driver_vehicle"
}
