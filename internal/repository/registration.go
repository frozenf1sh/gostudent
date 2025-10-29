package repository

import (
	"context"
	"time"

	"github.com/frozenf1sh/gostudent/internal/model"

	"gorm.io/gorm"
)

// RegistrationRepository 定义报名仓库接口
type RegistrationRepository interface {
	// WithTx 返回一个使用事务的仓库实例
	WithTx(tx *gorm.DB) RegistrationRepository

	Create(ctx context.Context, registration *model.Registration) error
	// FindByActivityAndPhone 检查是否重复报名
	FindByActivityAndPhone(ctx context.Context, activityID uint, phone string) (*model.Registration, error)
	ListByActivityID(ctx context.Context, activityID uint, page, pageSize int) ([]*model.Registration, int64, error)
	FindByID(ctx context.Context, id uint) (*model.Registration, error)
	UpdateSignInStatus(ctx context.Context, registrationID uint, signedIn bool, signedInAt time.Time) error
}

// registrationRepositoryImpl 实现了 RegistrationRepository 接口
type registrationRepositoryImpl struct {
	db *gorm.DB
}

// NewRegistrationRepository 创建一个新的 registrationRepositoryImpl
func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &registrationRepositoryImpl{db: db}
}

// WithTx 允许仓库在事务中运行
func (r *registrationRepositoryImpl) WithTx(tx *gorm.DB) RegistrationRepository {
	return &registrationRepositoryImpl{db: tx}
}

// Create 创建报名记录
// 必须在事务中调用
func (r *registrationRepositoryImpl) Create(ctx context.Context, registration *model.Registration) error {
	return r.db.WithContext(ctx).Create(registration).Error
}

// FindByActivityAndPhone 检查重复报名
// 可以在事务中调用 (使用 WithTx)
func (r *registrationRepositoryImpl) FindByActivityAndPhone(ctx context.Context, activityID uint, phone string) (*model.Registration, error) {
	var reg model.Registration
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND participant_phone = ?", activityID, phone).
		First(&reg).Error

	// Gorm 的 ErrRecordNotFound 是正常情况，表示未找到
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err // 真正的数据库错误
	}
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 未找到，不算错误
	}
	return &reg, nil // 找到了
}

// ListByActivityID 列出某个活动的所有报名者 (带分页)
func (r *registrationRepositoryImpl) ListByActivityID(ctx context.Context, activityID uint, page, pageSize int) ([]*model.Registration, int64, error) {
	var registrations []*model.Registration
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Registration{}).Where("activity_id = ?", activityID)

	// 1. 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. 应用分页
	offset := (page - 1) * pageSize
	if err := query.Order("registered_at ASC").Limit(pageSize).Offset(offset).Find(&registrations).Error; err != nil {
		return nil, 0, err
	}

	return registrations, total, nil
}

// FindByID 通过ID查找报名记录
func (r *registrationRepositoryImpl) FindByID(ctx context.Context, id uint) (*model.Registration, error) {
	var reg model.Registration
	if err := r.db.WithContext(ctx).First(&reg, id).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}

// UpdateSignInStatus 更新签到状态 (可选功能)
func (r *registrationRepositoryImpl) UpdateSignInStatus(ctx context.Context, registrationID uint, signedIn bool, signedInAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Registration{}).Where("id = ?", registrationID).Updates(map[string]interface{}{
		"is_signed_in": signedIn,
		"signed_in_at": signedInAt,
	}).Error
}
