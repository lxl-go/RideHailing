package carpool

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"ride-hailing/admin-server/global"
	carpoolModel "ride-hailing/admin-server/model/carpool"
	carpoolReq "ride-hailing/admin-server/model/carpool/request"
)

// PersonService 读写真实用户数据（auth_user + driver_profile/passenger_profile + driver_vehicle/driver_certification）。
// 不再使用 person_profile/person_role，管理端的乘客/司机数据与移动端同源。
type PersonService struct{}

type PersonDTO struct {
	carpoolModel.PersonProfile
	PhoneMasked  string                    `json:"phoneMasked"`
	IDCardMasked string                    `json:"idCardMasked"`
	Roles        []carpoolModel.PersonRole `json:"roles"`
}

type PersonStats struct {
	Total    int64 `json:"total"`
	Enabled  int64 `json:"enabled"`
	Disabled int64 `json:"disabled"`
	Active   int64 `json:"active"`
}

type PersonImportPreview struct {
	BatchNo      string                           `json:"batchNo"`
	Total        int                              `json:"total"`
	SuccessCount int                              `json:"successCount"`
	ErrorCount   int                              `json:"errorCount"`
	Errors       []carpoolModel.PersonImportError `json:"errors"`
}

// ---- 真实表模型（镜像移动端各 service 的表结构） ----

type authUserRow struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Principal string    `gorm:"column:principal"`
	Role      string    `gorm:"column:role"`
	Status    int       `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (authUserRow) TableName() string { return "auth_user" }

type driverProfileRow struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	Name                string    `gorm:"column:name"`
	Phone               string    `gorm:"column:phone"`
	AvatarURL           string    `gorm:"column:avatar_url"`
	ServiceStatus       int       `gorm:"column:service_status"`
	CertificationStatus int       `gorm:"column:certification_status"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (driverProfileRow) TableName() string { return "driver_profile" }

type passengerProfileRow struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	Nickname          string    `gorm:"column:nickname"`
	Phone             string    `gorm:"column:phone"`
	AvatarURL         string    `gorm:"column:avatar_url"`
	CommonAddress     string    `gorm:"column:common_address"`
	PaymentPreference string    `gorm:"column:payment_preference"`
	Status            int       `gorm:"column:status"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (passengerProfileRow) TableName() string { return "passenger_profile" }

type driverVehicleRow struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID    int64     `gorm:"column:driver_id"`
	PlateNo     string    `gorm:"column:plate_no"`
	Brand       string    `gorm:"column:brand"`
	Model       string    `gorm:"column:model"`
	Color       string    `gorm:"column:color"`
	VehicleType string    `gorm:"column:vehicle_type"`
	Seats       int       `gorm:"column:seats"`
	Status      int       `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (driverVehicleRow) TableName() string { return "driver_vehicle" }

type driverCertRow struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:false"`
	DriverID  int64     `gorm:"column:driver_id"`
	RealName  string    `gorm:"column:real_name"`
	IDCardNo  string    `gorm:"column:id_card_no"`
	LicenseNo string    `gorm:"column:license_no"`
	LicenseType string  `gorm:"column:license_type"`
	City      string    `gorm:"column:city"`
	Status    int       `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (driverCertRow) TableName() string { return "driver_certification" }

// ---- 常量 / 工具 ----

var personPhoneRE = regexp.MustCompile(`^1[3-9]\d{9}$`)

var TypeRolesMap = map[string][]string{
	"driver":    {"driver"},
	"passenger": {"passenger"},
	"staff":     {"staff"},
}

func personRolesForType(personType string) []string {
	if personType == "" {
		return nil
	}
	return TypeRolesMap[personType]
}

func statusToAuth(status string) int {
	if status == "disabled" || status == "deleted" {
		return 0
	}
	return 1
}

func authToStatus(status int) string {
	if status == 1 {
		return "enabled"
	}
	return "disabled"
}

func roleFromCode(code string) string {
	switch code {
	case "driver", "carpool_driver", "shuttle_driver", "pickup_driver":
		return "driver"
	case "passenger":
		return "passenger"
	case "staff":
		return "staff"
	}
	return code
}

func personRolesForRole(role string) []carpoolModel.PersonRole {
	switch role {
	case "driver":
		return []carpoolModel.PersonRole{{RoleCode: "carpool_driver", RoleName: "顺风车司机"}}
	case "passenger":
		return []carpoolModel.PersonRole{{RoleCode: "passenger", RoleName: "乘客"}}
	case "staff":
		return []carpoolModel.PersonRole{{RoleCode: "staff", RoleName: "员工"}}
	}
	return nil
}

func personRoleCodes(payload carpoolReq.PersonPayload) []string {
	// 允许前端传 roles；没有则按 personType 推导
	if len(payload.Roles) > 0 {
		return payload.Roles
	}
	return TypeRolesMap[payload.PersonType]
}

func validateAuthPerson(payload carpoolReq.PersonPayload) error {
	if payload.PersonType != "staff" && payload.PersonType != "driver" && payload.PersonType != "passenger" {
		return errors.New("invalid personType")
	}
	if !personPhoneRE.MatchString(payload.Phone) {
		return errors.New("invalid phone")
	}
	return nil
}

// ---- 读取：真实表 ----

func (s *PersonService) ListPersons(ctx context.Context, search carpoolReq.PersonSearch) ([]PersonDTO, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&authUserRow{})
	if roles := personRolesForType(search.PersonType); len(roles) > 0 {
		db = db.Where("role IN ?", roles)
	}
	if search.Status != "" {
		db = db.Where("status = ?", statusToAuth(search.Status))
	}
	if search.RoleCode != "" {
		db = db.Where("role = ?", roleFromCode(search.RoleCode))
	}
	if search.Keyword != "" {
		rowSub := personIdsByKeyword(ctx, search.Keyword)
		if len(rowSub) == 0 {
			return []PersonDTO{}, 0, nil
		}
		db = db.Where("id IN ?", rowSub)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := personLimitOffset(search.Page, search.PageSize)
	var users []authUserRow
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	list := make([]PersonDTO, 0, len(users))
	for _, u := range users {
		dto, err := s.buildDTO(ctx, u)
		if err != nil {
			continue
		}
		list = append(list, dto)
	}
	return list, total, nil
}

func personIdsByKeyword(ctx context.Context, keyword string) []int64 {
	kw := "%" + keyword + "%"
	set := map[int64]bool{}
	var ids []int64
	global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("principal LIKE ?", kw).Pluck("id", &ids)
	for _, id := range ids {
		set[id] = true
	}
	ids = ids[:0]
	global.GVA_DB.WithContext(ctx).Model(&driverProfileRow{}).Where("name LIKE ? OR phone LIKE ?", kw, kw).Pluck("id", &ids)
	for _, id := range ids {
		set[id] = true
	}
	ids = ids[:0]
	global.GVA_DB.WithContext(ctx).Model(&passengerProfileRow{}).Where("nickname LIKE ? OR phone LIKE ?", kw, kw).Pluck("id", &ids)
	for _, id := range ids {
		set[id] = true
	}
	result := make([]int64, 0, len(set))
	for id := range set {
		result = append(result, id)
	}
	return result
}

func (s *PersonService) GetPersonDetail(ctx context.Context, id uint64) (*PersonDTO, error) {
	var u authUserRow
	if err := global.GVA_DB.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	dto, err := s.buildDTO(ctx, u)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *PersonService) buildDTO(ctx context.Context, u authUserRow) (PersonDTO, error) {
	roles := personRolesForRole(u.Role)
	profile := carpoolModel.PersonProfile{
		ID:             uint64(u.ID),
		PersonNo:       fmt.Sprintf("U%d", u.ID),
		PersonType:         u.Role,
		Name:               u.Principal,
		DriverLicenseNo:    "",
		VehicleNo:          "",
		VehicleType:        "",
		CommonAddress:      "",
		PaymentPreference:  "",
		Rating:             5,
		Status:             authToStatus(u.Status),
		RegisterDate:       u.CreatedAt,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}
	dto := PersonDTO{
		PersonProfile: profile,
		PhoneMasked:   maskPhone(u.Principal),
		Roles:         roles,
	}
	if u.Role == "driver" {
		var dp driverProfileRow
		if err := global.GVA_DB.WithContext(ctx).Where("id = ?", u.ID).First(&dp).Error; err == nil {
			dto.PersonProfile.Name = fallbackPersonStr(dp.Name, u.Principal)
		}
		var v driverVehicleRow
		if err := global.GVA_DB.WithContext(ctx).Where("driver_id = ?", u.ID).Order("status ASC, id ASC").First(&v).Error; err == nil {
			dto.PersonProfile.VehicleNo = v.PlateNo
			dto.PersonProfile.VehicleType = v.VehicleType
		}
		var c driverCertRow
		if err := global.GVA_DB.WithContext(ctx).Where("driver_id = ?", u.ID).Order("created_at ASC").First(&c).Error; err == nil {
			dto.IDCardMasked = maskIDCard(c.IDCardNo)
			dto.PersonProfile.Name = fallbackPersonStr(c.RealName, dto.PersonProfile.Name)
			dto.PersonProfile.DriverLicenseNo = c.LicenseNo
		}
	} else {
		var pp passengerProfileRow
		if err := global.GVA_DB.WithContext(ctx).Where("id = ?", u.ID).First(&pp).Error; err == nil {
			dto.PersonProfile.Name = fallbackPersonStr(pp.Nickname, u.Principal)
			dto.PersonProfile.CommonAddress = pp.CommonAddress
			dto.PersonProfile.PaymentPreference = pp.PaymentPreference
		}
	}
	return dto, nil
}

func (s *PersonService) GetStats(ctx context.Context, personType string) (*PersonStats, error) {
	roles := personRolesForType(personType)
	base := global.GVA_DB.WithContext(ctx).Model(&authUserRow{})
	if len(roles) > 0 {
		base = base.Where("role IN ?", roles)
	}
	result := &PersonStats{}
	if err := base.Count(&result.Total).Error; err != nil {
		return nil, err
	}
	if err := base.Where("status = 1").Count(&result.Enabled).Error; err != nil {
		return nil, err
	}
	if err := base.Where("status = 0").Count(&result.Disabled).Error; err != nil {
		return nil, err
	}
	result.Active = result.Enabled
	return result, nil
}

// ---- 写入：真实表 ----

func (s *PersonService) CreatePerson(ctx context.Context, payload carpoolReq.PersonPayload) (*PersonDTO, error) {
	if err := validateAuthPerson(payload); err != nil {
		return nil, err
	}
	role := payload.PersonType
	if len(payload.Roles) > 0 {
		role = payload.Roles[0]
	}
	role = roleFromCode(role)
	principal := strings.TrimSpace(payload.Phone)
	var exist int64
	if err := global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("principal = ?", principal).Count(&exist).Error; err != nil {
		return nil, err
	}
	if exist > 0 {
		return nil, errors.New("account already exists")
	}
	var user authUserRow
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user = authUserRow{Principal: principal, Role: role, Status: 1}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if role == "driver" {
			dp := driverProfileRow{ID: user.ID, Name: payload.Name, Phone: payload.Phone}
			if err := tx.Create(&dp).Error; err != nil {
				return err
			}
		} else {
			pp := passengerProfileRow{ID: user.ID, Nickname: fallbackPersonStr(payload.Name, payload.Phone), Phone: payload.Phone, PaymentPreference: payload.PaymentPreference}
			if err := tx.Create(&pp).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetPersonDetail(ctx, uint64(user.ID))
}

func (s *PersonService) UpdatePerson(ctx context.Context, id uint64, payload carpoolReq.PersonPayload) (*PersonDTO, error) {
	var user authUserRow
	if err := global.GVA_DB.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	name := strings.TrimSpace(payload.Name)
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if user.Role == "driver" {
			updates := map[string]any{}
			if name != "" {
				updates["name"] = name
			}
			if payload.Phone != "" {
				updates["phone"] = payload.Phone
			}
			if len(updates) > 0 {
				if err := tx.Model(&driverProfileRow{}).Where("id = ?", id).Updates(updates).Error; err != nil {
					return err
				}
			}
			if payload.IDCardNo != "" {
				if err := tx.Model(&driverCertRow{}).Where("driver_id = ?", id).Updates(map[string]interface{}{"id_card_no": payload.IDCardNo, "real_name": name}).Error; err != nil {
					return err
				}
			}
		} else {
			updates := map[string]interface{}{}
			if name != "" {
				updates["nickname"] = name
			}
			if payload.Phone != "" {
				updates["phone"] = payload.Phone
			}
			if payload.PaymentPreference != "" {
				updates["payment_preference"] = payload.PaymentPreference
			}
			if len(updates) != 0 {
				if err := tx.Model(&passengerProfileRow{}).Where("id = ?", id).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetPersonDetail(ctx, id)
}

func (s *PersonService) AssignRoles(ctx context.Context, req carpoolReq.PersonRoleAssign) (*PersonDTO, error) {
	role := ""
	if len(req.Roles) > 0 {
		role = req.Roles[0]
	}
	if role != "" {
		if err := global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("id = ?", req.PersonID).Update("role", roleFromCode(role)).Error; err != nil {
			return nil, err
		}
	}
	return s.GetPersonDetail(ctx, req.PersonID)
}

func (s *PersonService) BatchUpdateStatus(ctx context.Context, req carpoolReq.PersonBatchStatus) error {
	if len(req.IDs) == 0 {
		return errors.New("ids is required")
	}
	authStatus := statusToAuth(req.Status)
	err := global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("id IN ?", req.IDs).Update("status", authStatus).Error
	if err != nil {
		return err
	}
	// 司机联动 service_status：禁用 → 3（停用）
	if req.Status == "disabled" || req.Status == "deleted" {
		_ = global.GVA_DB.WithContext(ctx).Model(&driverProfileRow{}).Where("id IN ?", req.IDs).Update("service_status", 3).Error
	}
	return nil
}

func (s *PersonService) BatchDeleteDrivers(ctx context.Context, req carpoolReq.PersonBatchDelete) error {
	if len(req.IDs) == 0 {
		return errors.New("ids is required")
	}
	var drivers int64
	if err := global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("id IN ? AND role = ?", req.IDs, "driver").Count(&drivers).Error; err != nil {
		return err
	}
	if drivers != int64(len(req.IDs)) {
		return errors.New("only driver can be batch deleted")
	}
	return global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("id IN ?", req.IDs).Update("status", 0).Error
}

// ---- 导入（创建真实用户） ----

func (s *PersonService) PreviewImport(ctx context.Context, payload carpoolReq.PersonImportPayload) (*PersonImportPreview, error) {
	return s.previewImport(ctx, payload, false)
}

func (s *PersonService) CommitImport(ctx context.Context, payload carpoolReq.PersonImportPayload) (*PersonImportPreview, error) {
	preview, err := s.previewImport(ctx, payload, true)
	if err != nil {
		return preview, err
	}
	if preview.ErrorCount > 0 {
		return preview, errors.New("import rows contain validation errors")
	}
	return preview, nil
}

func (s *PersonService) previewImport(ctx context.Context, payload carpoolReq.PersonImportPayload, commit bool) (*PersonImportPreview, error) {
	batchNo := fmt.Sprintf("PERSON-IMP-%d", time.Now().UnixNano())
	errorList := make([]carpoolModel.PersonImportError, 0)
	success := 0
	for i, row := range payload.Rows {
		rowNo := i + 1
		if err := validateAuthPerson(row); err != nil {
			errorList = append(errorList, carpoolModel.PersonImportError{BatchNo: batchNo, RowNo: rowNo, Field: "row", Message: err.Error()})
			continue
		}
		var count int64
		_ = global.GVA_DB.WithContext(ctx).Model(&authUserRow{}).Where("principal = ?", strings.TrimSpace(row.Phone)).Count(&count).Error
		if count > 0 {
			errorList = append(errorList, carpoolModel.PersonImportError{BatchNo: batchNo, RowNo: rowNo, Field: "phone", Message: "duplicate phone"})
			continue
		}
		if commit {
			if _, err := s.CreatePerson(ctx, row); err != nil {
				errorList = append(errorList, carpoolModel.PersonImportError{BatchNo: batchNo, RowNo: rowNo, Field: "row", Message: err.Error()})
				continue
			}
		}
		success++
	}
	return &PersonImportPreview{BatchNo: batchNo, Total: len(payload.Rows), SuccessCount: success, ErrorCount: len(errorList), Errors: errorList}, nil
}

func (s *PersonService) ExportPersons(ctx context.Context) string {
	return fmt.Sprintf("PERSON-EXP-%d", time.Now().UnixNano())
}

// ListImportErrors 查询导入错误（person_import_error 表，按批次过滤）。
func (s *PersonService) ListImportErrors(ctx context.Context, search carpoolReq.PersonImportErrorSearch) ([]carpoolModel.PersonImportError, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&carpoolModel.PersonImportError{})
	if search.BatchNo != "" {
		db = db.Where("batch_no = ?", search.BatchNo)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := personLimitOffset(search.Page, search.PageSize)
	var list []carpoolModel.PersonImportError
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// SeedPersonDefaults 已废弃：管理端不再写入 person_profile 假数据。
func SeedPersonDefaults(db *gorm.DB) error {
	return nil
}

func personLimitOffset(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return pageSize, (page - 1) * pageSize
}

func fallbackPersonStr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func maskIDCard(id string) string {
	if len(id) < 10 {
		return id
	}
	return id[:6] + "********" + id[len(id)-4:]
}