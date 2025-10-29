package repository

import (
	"context"

	"github.com/frozenf1sh/gostudent/internal/model"

	"gorm.io/gorm"
)

// AdminRepository 定义管理员仓库接口
type AdminRepository interface {
	Create(ctx context.Context, admin *model.Admin) error
	FindByUsername(ctx context.Context, username string) (*model.Admin, error)
	FindByID(ctx context.Context, id uint) (*model.Admin, error)
}

// adminRepositoryImpl 实现了 AdminRepository 接口
type adminRepositoryImpl struct {
	db *gorm.DB
}

// NewAdminRepository 创建一个新的 adminRepositoryImpl
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepositoryImpl{db: db}
}

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
