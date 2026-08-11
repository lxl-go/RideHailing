package carpool

import (
	"time"

	"gorm.io/gorm"
)

type VehicleInfo struct {
	ID              int64          `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:主键ID"`
	DriverID        int64          `json:"driverId" gorm:"column:driver_id;type:bigint;not null;index:idx_driver_status;comment:司机用户ID"`
	PlateNumber     string         `json:"plateNumber" gorm:"column:plate_number;type:varchar(32);not null;uniqueIndex:uk_plate_number;comment:车牌号"`
	Brand           string         `json:"brand" gorm:"column:brand;type:varchar(64);not null;comment:品牌"`
	Model           string         `json:"model" gorm:"column:model;type:varchar(64);not null;comment:车型"`
	Color           string         `json:"color" gorm:"column:color;type:varchar(32);not null;comment:颜色"`
	Seats           int            `json:"seats" gorm:"column:seats;type:int;not null;default:4;comment:座位数"`
	YearCheckDate   *time.Time     `json:"yearCheckDate" gorm:"column:year_check_date;datetime;default null;comment:年检到期日"`
	InsuranceExpire *time.Time     `json:"insuranceExpire" gorm:"column:insurance_expire;datetime;default null;comment:保险到期日"`
	Status          int            `json:"status" gorm:"column:status;type:tinyint;not null;default:0;index:idx_driver_status;comment:0待审核 1通过 2驳回"`
	ReviewerID      int64          `json:"reviewerId" gorm:"column:reviewer_id;type:bigint;default:0;comment:审核人ID"`
	RejectReason    string         `json:"rejectReason" gorm:"column:reject_reason;type:varchar(512);default:'';comment:驳回原因"`
	CreatedAt       time.Time      `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:软删除时间"`
}

func (VehicleInfo) TableName() string {
	return "carpool_vehicle_info"
}
