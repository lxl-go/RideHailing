package carpool

import "time"

type PersonProfile struct {
	ID                uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:person id" json:"id"`
	PersonNo          string    `gorm:"column:person_no;type:varchar(64);not null;uniqueIndex:uk_person_no;comment:person number" json:"personNo"`
	PersonType        string    `gorm:"column:person_type;type:varchar(16);not null;index:idx_person_type_status;comment:staff driver passenger" json:"personType"`
	Name              string    `gorm:"column:name;type:varchar(64);not null;index:idx_person_keyword;comment:name" json:"name"`
	Phone             string    `gorm:"column:phone;type:varchar(32);not null;comment:phone" json:"-"`
	PhoneHash         string    `gorm:"column:phone_hash;type:varchar(64);not null;uniqueIndex:uk_person_phone_hash;comment:phone hash" json:"-"`
	Email             string    `gorm:"column:email;type:varchar(128);not null;default:'';comment:email" json:"email"`
	IDCardNo          string    `gorm:"column:id_card_no;type:varchar(32);not null;comment:id card" json:"-"`
	IDCardHash        string    `gorm:"column:id_card_hash;type:varchar(64);not null;uniqueIndex:uk_person_id_card_hash;comment:id card hash" json:"-"`
	DriverLicenseNo   string    `gorm:"column:driver_license_no;type:varchar(64);not null;default:'';comment:driver license" json:"driverLicenseNo"`
	VehicleNo         string    `gorm:"column:vehicle_no;type:varchar(32);not null;default:'';comment:vehicle number" json:"vehicleNo"`
	VehicleType       string    `gorm:"column:vehicle_type;type:varchar(32);not null;default:'';comment:vehicle type" json:"vehicleType"`
	CommonAddress     string    `gorm:"column:common_address;type:varchar(255);not null;default:'';comment:common address" json:"commonAddress"`
	PaymentPreference string    `gorm:"column:payment_preference;type:varchar(32);not null;default:'';comment:payment preference" json:"paymentPreference"`
	Rating            float64   `gorm:"column:rating;type:decimal(3,2);not null;default:5;comment:rating" json:"rating"`
	Status            string    `gorm:"column:status;type:varchar(16);not null;default:'enabled';index:idx_person_type_status;comment:enabled disabled deleted" json:"status"`
	RegisterDate      time.Time `gorm:"column:register_date;not null;comment:register date" json:"registerDate"`
	DisabledReason    string    `gorm:"column:disabled_reason;type:varchar(255);not null;default:'';comment:disabled reason" json:"disabledReason"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;comment:updated time" json:"updatedAt"`
}

func (PersonProfile) TableName() string {
	return "person_profile"
}

type PersonRole struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:person role id" json:"id"`
	PersonID  uint64    `gorm:"column:person_id;type:bigint;not null;uniqueIndex:uk_person_role;index:idx_person_role_person;comment:person id" json:"personId"`
	RoleCode  string    `gorm:"column:role_code;type:varchar(32);not null;uniqueIndex:uk_person_role;index:idx_person_role_code;comment:role code" json:"roleCode"`
	RoleName  string    `gorm:"column:role_name;type:varchar(64);not null;comment:role name" json:"roleName"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (PersonRole) TableName() string {
	return "person_role"
}

type PersonImportBatch struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:batch id" json:"id"`
	BatchNo      string    `gorm:"column:batch_no;type:varchar(64);not null;uniqueIndex:uk_person_import_batch;comment:batch number" json:"batchNo"`
	SourceType   string    `gorm:"column:source_type;type:varchar(16);not null;default:'json';comment:source type" json:"sourceType"`
	Total        int       `gorm:"column:total;type:int;not null;default:0;comment:total rows" json:"total"`
	SuccessCount int       `gorm:"column:success_count;type:int;not null;default:0;comment:success rows" json:"successCount"`
	ErrorCount   int       `gorm:"column:error_count;type:int;not null;default:0;comment:error rows" json:"errorCount"`
	Status       string    `gorm:"column:status;type:varchar(16);not null;default:'success';index:idx_person_import_status;comment:success failed preview" json:"status"`
	Operator     string    `gorm:"column:operator;type:varchar(64);not null;default:'';comment:operator" json:"operator"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (PersonImportBatch) TableName() string {
	return "person_import_batch"
}

type PersonImportError struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:error id" json:"id"`
	BatchNo   string    `gorm:"column:batch_no;type:varchar(64);not null;index:idx_person_import_error_batch;comment:batch number" json:"batchNo"`
	RowNo     int       `gorm:"column:row_no;type:int;not null;comment:row number" json:"rowNo"`
	Field     string    `gorm:"column:field;type:varchar(64);not null;comment:error field" json:"field"`
	Message   string    `gorm:"column:message;type:varchar(255);not null;comment:error message" json:"message"`
	RawData   string    `gorm:"column:raw_data;type:text;comment:raw data json" json:"rawData"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:created time" json:"createdAt"`
}

func (PersonImportError) TableName() string {
	return "person_import_error"
}
