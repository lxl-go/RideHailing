package carpool

import "time"

type DriverProfile struct {
	ID                  int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:司机ID"`
	Name                string    `json:"name" gorm:"column:name;type:varchar(64);not null;default:'';comment:司机姓名"`
	Phone               string    `json:"phone" gorm:"column:phone;type:varchar(32);not null;default:'';comment:手机号"`
	AvatarURL           string    `json:"avatarUrl" gorm:"column:avatar_url;type:varchar(255);not null;default:'';comment:头像"`
	ServiceStatus       int       `json:"serviceStatus" gorm:"column:service_status;type:tinyint;not null;default:1;comment:服务状态"`
	CertificationStatus int       `json:"certificationStatus" gorm:"column:certification_status;type:tinyint;not null;default:1;comment:认证状态 1草稿 2审核中 3通过 4驳回"`
	CreatedAt           time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt           time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (DriverProfile) TableName() string {
	return "driver_profile"
}

type DriverCertification struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement:false;comment:认证ID"`
	DriverID         int64     `json:"driverId" gorm:"column:driver_id;type:bigint;not null;uniqueIndex:uk_driver_certification;comment:司机ID"`
	RealName         string    `json:"realName" gorm:"column:real_name;type:varchar(64);not null;default:'';comment:真实姓名"`
	IDCardNo         string    `json:"idCardNo" gorm:"column:id_card_no;type:varchar(32);not null;default:'';comment:身份证号"`
	LicenseNo        string    `json:"licenseNo" gorm:"column:license_no;type:varchar(64);not null;default:'';comment:驾驶证号"`
	LicenseType      string    `json:"licenseType" gorm:"column:license_type;type:varchar(32);not null;default:'';comment:准驾车型"`
	City             string    `json:"city" gorm:"column:city;type:varchar(64);not null;default:'';comment:所属城市"`
	VehicleLicenseNo string    `json:"vehicleLicenseNo" gorm:"column:vehicle_license_no;type:varchar(64);not null;default:'';comment:行驶证号"`
	VehiclePhotoURL  string    `json:"vehiclePhotoUrl" gorm:"column:vehicle_photo_url;type:varchar(255);not null;default:'';comment:车辆照片"`
	FacePhotoURL     string    `json:"facePhotoUrl" gorm:"column:face_photo_url;type:varchar(255);not null;default:'';comment:人脸照片"`
	Status           int       `json:"status" gorm:"column:status;type:tinyint;not null;default:2;comment:认证状态 1草稿 2审核中 3通过 4驳回"`
	RejectReason     string    `json:"rejectReason" gorm:"column:reject_reason;type:varchar(255);not null;default:'';comment:驳回原因"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (DriverCertification) TableName() string {
	return "driver_certification"
}
