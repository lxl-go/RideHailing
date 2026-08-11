package data

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	RolePassenger = "passenger"
	RoleDriver    = "driver"
)

type roleModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Code      string    `gorm:"column:code;type:varchar(64);not null;uniqueIndex"`
	Name      string    `gorm:"column:name;type:varchar(64);not null;default:''"`
	Status    int       `gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (roleModel) TableName() string {
	return "auth_role"
}

type userRoleModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:idx_auth_user_role_user_role"`
	RoleID    int64     `gorm:"column:role_id;not null;uniqueIndex:idx_auth_user_role_user_role"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (userRoleModel) TableName() string {
	return "auth_user_role"
}

type permissionModel struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Code      string    `gorm:"column:code;type:varchar(128);not null;uniqueIndex"`
	Name      string    `gorm:"column:name;type:varchar(128);not null;default:''"`
	Resource  string    `gorm:"column:resource;type:varchar(128);not null;default:''"`
	Action    string    `gorm:"column:action;type:varchar(64);not null;default:''"`
	Status    int       `gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (permissionModel) TableName() string {
	return "auth_permission"
}

type rolePermissionModel struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	RoleID       int64     `gorm:"column:role_id;not null;uniqueIndex:idx_auth_role_permission_role_permission"`
	PermissionID int64     `gorm:"column:permission_id;not null;uniqueIndex:idx_auth_role_permission_role_permission"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (rolePermissionModel) TableName() string {
	return "auth_role_permission"
}

type builtinPermission struct {
	Code     string
	Name     string
	Resource string
	Action   string
	Roles    []string
}

var builtinPermissions = []builtinPermission{
	{Code: "trip:search", Name: "Search trips", Resource: "trip", Action: "search", Roles: []string{RolePassenger, RoleDriver}},
	{Code: "trip:view_detail", Name: "View trip detail", Resource: "trip", Action: "view_detail", Roles: []string{RolePassenger, RoleDriver}},
	{Code: "trip:publish", Name: "Publish trip", Resource: "trip", Action: "publish", Roles: []string{RoleDriver}},
	{Code: "trip:list_driver_self", Name: "List own driver trips", Resource: "trip", Action: "list_driver_self", Roles: []string{RoleDriver}},
	{Code: "trip:update_status_self", Name: "Update own trip status", Resource: "trip", Action: "update_status_self", Roles: []string{RoleDriver}},
	{Code: "order:create", Name: "Create order", Resource: "order", Action: "create", Roles: []string{RolePassenger}},
	{Code: "order:cancel_self", Name: "Cancel own order", Resource: "order", Action: "cancel_self", Roles: []string{RolePassenger}},
	{Code: "order:list_passenger_self", Name: "List own passenger orders", Resource: "order", Action: "list_passenger_self", Roles: []string{RolePassenger}},
	{Code: "order:view_passenger_self", Name: "View own passenger order", Resource: "order", Action: "view_passenger_self", Roles: []string{RolePassenger}},
	{Code: "order:list_driver_pending", Name: "List pending driver orders", Resource: "order", Action: "list_driver_pending", Roles: []string{RoleDriver}},
	{Code: "order:accept_driver_self", Name: "Accept own driver order", Resource: "order", Action: "accept_driver_self", Roles: []string{RoleDriver}},
	{Code: "order:reject_driver_self", Name: "Reject own driver order", Resource: "order", Action: "reject_driver_self", Roles: []string{RoleDriver}},
	{Code: "review:submit", Name: "Submit review", Resource: "review", Action: "submit", Roles: []string{RolePassenger, RoleDriver}},
	{Code: "passenger:profile:view_self", Name: "View passenger profile", Resource: "passenger_profile", Action: "view_self", Roles: []string{RolePassenger}},
	{Code: "passenger:profile:update_self", Name: "Update passenger profile", Resource: "passenger_profile", Action: "update_self", Roles: []string{RolePassenger}},
	{Code: "driver:profile:view_self", Name: "View driver profile", Resource: "driver_profile", Action: "view_self", Roles: []string{RoleDriver}},
	{Code: "driver:profile:update_self", Name: "Update driver profile", Resource: "driver_profile", Action: "update_self", Roles: []string{RoleDriver}},
	{Code: "driver:certification:submit_self", Name: "Submit driver certification", Resource: "driver_certification", Action: "submit_self", Roles: []string{RoleDriver}},
	{Code: "driver:certification:view_self", Name: "View driver certification", Resource: "driver_certification", Action: "view_self", Roles: []string{RoleDriver}},
	{Code: "driver:vehicle:manage_self", Name: "Manage driver vehicle", Resource: "driver_vehicle", Action: "manage_self", Roles: []string{RoleDriver}},
	{Code: "driver:vehicle:list_self", Name: "List driver vehicles", Resource: "driver_vehicle", Action: "list_self", Roles: []string{RoleDriver}},
}

func SeedBuiltinRBAC(ctx context.Context, db *gorm.DB) error {
	roles := []roleModel{
		{Code: RolePassenger, Name: "Passenger", Status: 1},
		{Code: RoleDriver, Name: "Driver", Status: 1},
	}
	for _, role := range roles {
		if err := db.WithContext(ctx).Where("code = ?", role.Code).FirstOrCreate(&role).Error; err != nil {
			return err
		}
		if role.Name == "" || role.Status == 0 {
			role.Name = role.Code
			role.Status = 1
			if err := db.WithContext(ctx).Save(&role).Error; err != nil {
				return err
			}
		}
	}

	roleByCode := map[string]roleModel{}
	var existingRoles []roleModel
	if err := db.WithContext(ctx).Where("code IN ?", []string{RolePassenger, RoleDriver}).Find(&existingRoles).Error; err != nil {
		return err
	}
	for _, role := range existingRoles {
		roleByCode[role.Code] = role
	}

	for _, item := range builtinPermissions {
		permission := permissionModel{
			Code:     item.Code,
			Name:     item.Name,
			Resource: item.Resource,
			Action:   item.Action,
			Status:   1,
		}
		if err := db.WithContext(ctx).Where("code = ?", permission.Code).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
		if permission.Name == "" || permission.Resource == "" || permission.Action == "" || permission.Status == 0 {
			permission.Name = item.Name
			permission.Resource = item.Resource
			permission.Action = item.Action
			permission.Status = 1
			if err := db.WithContext(ctx).Save(&permission).Error; err != nil {
				return err
			}
		}
		for _, roleCode := range item.Roles {
			role, ok := roleByCode[roleCode]
			if !ok {
				continue
			}
			binding := rolePermissionModel{RoleID: role.ID, PermissionID: permission.ID}
			if err := db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).FirstOrCreate(&binding).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type RBACRepo struct {
	db *gorm.DB
}

func NewRBACRepo(db *gorm.DB) *RBACRepo {
	return &RBACRepo{db: db}
}

func (r *RBACRepo) CheckUserPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table(accountModel{}.TableName()+" AS u").
		Joins("JOIN "+userRoleModel{}.TableName()+" AS ur ON ur.user_id = u.id").
		Joins("JOIN "+roleModel{}.TableName()+" AS r ON r.id = ur.role_id").
		Joins("JOIN "+rolePermissionModel{}.TableName()+" AS rp ON rp.role_id = r.id").
		Joins("JOIN "+permissionModel{}.TableName()+" AS p ON p.id = rp.permission_id").
		Where("u.id = ? AND u.status = ? AND r.status = ? AND p.status = ? AND p.code = ?", userID, 1, 1, 1, permissionCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
