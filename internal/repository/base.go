package repository

import "gorm.io/gorm"

// 仓库聚合设计模式
// Repository 接口定义了所有 Repository 层的聚合
type Repository interface {
	Admin() AdminRepository
	Activity() ActivityRepository
	Registration() RegistrationRepository
	// ... 在此添加其他 repo
}

// repositoryImpl 实现了 Repository 接口
type repositoryImpl struct {
	adminRepo        AdminRepository
	activityRepo     ActivityRepository
	registrationRepo RegistrationRepository
}

// NewRepository 创建 Repository 实例，并注入所有依赖
// 真正的依赖注入发生在这里
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{
		adminRepo:        NewAdminRepository(db),
		activityRepo:     NewActivityRepository(db),
		registrationRepo: NewRegistrationRepository(db),
	}
}

// Admin 返回 AdminRepository 实例
func (r *repositoryImpl) Admin() AdminRepository {
	return r.adminRepo
}

// Activity 返回 ActivityRepository 实例
func (r *repositoryImpl) Activity() ActivityRepository {
	return r.activityRepo
}

// Registration 返回 RegistrationRepository 实例
func (r *repositoryImpl) Registration() RegistrationRepository {
	return r.registrationRepo
}
