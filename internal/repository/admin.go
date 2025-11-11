package repository

import (
	"context"

	"github.com/frozenf1sh/gostudent/internal/model"

	"gorm.io/gorm"
)

// 接口：管理员仓库
type AdminRepository interface {
	// 创建管理员
	Create(ctx context.Context, admin *model.Admin) error
	// 根据用户名查找第一个
	FindByUsername(ctx context.Context, username string) (*model.Admin, error)
	// 根据ID查找
	FindByID(ctx context.Context, id uint) (*model.Admin, error)
}

// ----- 实现 -----

// 管理员仓库实现
type adminRepositoryImpl struct {
	db *gorm.DB
}

// 构造函数
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepositoryImpl{db: db}
}

// 方法实现
// Create 创建管理员
func (r *adminRepositoryImpl) Create(ctx context.Context, admin *model.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

// FindByUsername 通过用户名查找
func (r *adminRepositoryImpl) FindByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindByID 通过ID查找
func (r *adminRepositoryImpl) FindByID(ctx context.Context, id uint) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.WithContext(ctx).First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}
