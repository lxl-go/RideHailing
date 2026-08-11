package carpool

import (
	"time"

	"gorm.io/gorm"
)

type CertificationAudit struct {
	ID                  int64          `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:主键ID"`
	UserID              int64          `json:"userId" gorm:"column:user_id;type:bigint;not null;index:idx_user_role;comment:用户ID"`
	UserRole            int            `json:"userRole" gorm:"column:user_role;type:tinyint;not null;default:2;index:idx_user_role;comment:角色 1乘客 2司机"`
	CertType            int            `json:"certType" gorm:"column:cert_type;type:tinyint;not null;index:idx_cert_status;comment:证件类型 1身份证 2驾驶证 3行驶证"`
	CertNumber          string         `json:"certNumber" gorm:"column:cert_number;type:varchar(256);not null;comment:证件号AES加密"`
	RealName            string         `json:"realName" gorm:"column:real_name;type:varchar(64);not null;comment:真实姓名"`
	DriverLicenseNo     string         `json:"driverLicenseNo" gorm:"column:driver_license_no;type:varchar(64);not null;default:'';comment:驾驶证号"`
	LicenseType         string         `json:"licenseType" gorm:"column:license_type;type:varchar(32);not null;default:'';comment:准驾车型"`
	City                string         `json:"city" gorm:"column:city;type:varchar(64);not null;default:'';comment:所属城市"`
	FrontImageURL       string         `json:"frontImageUrl" gorm:"column:front_image_url;type:varchar(512);not null;comment:证件正面照"`
	BackImageURL        string         `json:"backImageUrl" gorm:"column:back_image_url;type:varchar(512);default:'';comment:证件反面照"`
	HandheldImageURL    string         `json:"handheldImageUrl" gorm:"column:handheld_image_url;type:varchar(512);default:'';comment:手持证件照"`
	Status              int            `json:"status" gorm:"column:status;type:tinyint;not null;default:0;index:idx_cert_status;comment:0待审核 1通过 2驳回 3补充"`
	ReviewerID          int64          `json:"reviewerId" gorm:"column:reviewer_id;type:bigint;default:0;comment:审核人ID"`
	RejectReason        string         `json:"rejectReason" gorm:"column:reject_reason;type:varchar(512);default:'';comment:驳回原因"`
	SubmitCount         int            `json:"submitCount" gorm:"column:submit_count;type:int;not null;default:0;comment:提交审核次数"`
	ReviewedAt          *time.Time     `json:"reviewedAt" gorm:"column:reviewed_at;datetime;default null;comment:审核完成时间"`
	ReviewDurationHours int            `json:"reviewDurationHours" gorm:"column:review_duration_hours;type:int;not null;default:0;comment:审核耗时(小时)"`
	CreatedAt           time.Time      `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt           time.Time      `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:软删除时间"`
}

func (CertificationAudit) TableName() string {
	return "carpool_certification_audit"
}
