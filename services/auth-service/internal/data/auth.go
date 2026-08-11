package data

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/auth-service/internal/biz"
)

type accountModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Principal   string    `gorm:"column:principal;type:varchar(64);not null;index:idx_auth_user_principal_role,unique"`
	Role        string    `gorm:"column:role;type:varchar(32);not null;index:idx_auth_user_principal_role,unique"`
	Status      int       `gorm:"column:status;type:tinyint;not null;default:1"`
	LastLoginAt time.Time `gorm:"column:last_login_at"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (accountModel) TableName() string {
	return "auth_user"
}

type AccountRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewAccountRepo(db *gorm.DB, log *zap.Logger) *AccountRepo {
	return &AccountRepo{db: db, log: log}
}

func (r *AccountRepo) GetByPrincipalAndRole(ctx context.Context, principal string, role biz.Role) (*biz.Account, error) {
	var m accountModel
	if err := r.db.WithContext(ctx).Where("principal = ? AND role = ?", principal, string(role)).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *AccountRepo) GetByID(ctx context.Context, id int64) (*biz.Account, error) {
	var m accountModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *AccountRepo) Create(ctx context.Context, account *biz.Account) error {
	if account.LastLoginAt.IsZero() {
		account.LastLoginAt = time.Now()
	}
	record := toRecord(account)
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return err
	}
	account.ID = record.ID
	account.CreatedAt = record.CreatedAt
	account.UpdatedAt = record.UpdatedAt
	return nil
}

func (r *AccountRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&accountModel{}).Where("id = ?", id).Updates(map[string]any{"last_login_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrAccountNotFound
	}
	return nil
}

func (r *AccountRepo) EnsureUserRole(ctx context.Context, userID int64, role biz.Role) error {
	roleRecord := roleModel{
		Code:   string(role),
		Name:   string(role),
		Status: 1,
	}
	if err := r.db.WithContext(ctx).Where("code = ?", roleRecord.Code).FirstOrCreate(&roleRecord, roleModel{Code: roleRecord.Code}).Error; err != nil {
		return err
	}
	if roleRecord.Name == "" || roleRecord.Status == 0 {
		roleRecord.Name = string(role)
		roleRecord.Status = 1
		if err := r.db.WithContext(ctx).Save(&roleRecord).Error; err != nil {
			return err
		}
	}
	relation := userRoleModel{UserID: userID, RoleID: roleRecord.ID}
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleRecord.ID).FirstOrCreate(&relation).Error
}

func (r *AccountRepo) UserHasRole(ctx context.Context, userID int64, role biz.Role) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table(userRoleModel{}.TableName()+" AS ur").
		Joins("JOIN "+roleModel{}.TableName()+" AS r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.code = ?", userID, string(role)).
		Count(&count).Error
	return count > 0, err
}

func toDomain(m *accountModel) *biz.Account {
	return &biz.Account{
		ID:          m.ID,
		Principal:   m.Principal,
		Role:        biz.Role(m.Role),
		Status:      m.Status,
		LastLoginAt: m.LastLoginAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toRecord(e *biz.Account) *accountModel {
	return &accountModel{
		ID:          e.ID,
		Principal:   e.Principal,
		Role:        string(e.Role),
		Status:      e.Status,
		LastLoginAt: e.LastLoginAt,
	}
}
