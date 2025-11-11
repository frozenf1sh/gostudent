package service

import (
	"gorm.io/gorm"

	"github.com/frozenf1sh/gostudent/internal/repository"
)

// Service 接口定义了所有 Service 层的聚合
type Service interface {
	Admin() AdminService
	Activity() ActivityService
	Registration() RegistrationService
	// Add other services here...
}

// serviceImpl 实现了 Service 接口
type serviceImpl struct {
	adminService        AdminService
	activityService     ActivityService
	registrationService RegistrationService
}

// Option 用于 Service 结构体的可选配置
type Option func(*serviceImpl)

// NewService 创建 Service 实例，并注入所有依赖
// db 实例用于需要事务的 Service
func NewService(db *gorm.DB, repo repository.Repository) Service {
	// 初始化所有具体的 Service
	s := &serviceImpl{
		adminService:        NewAdminService(repo.Admin()),
		activityService:     NewActivityService(db, repo.Activity()),
		registrationService: NewRegistrationService(db, repo.Activity(), repo.Registration()),
	}

	return s
}

// Admin 返回 AdminService 实例
func (s *serviceImpl) Admin() AdminService {
	return s.adminService
}

// Activity 返回 ActivityService 实例
func (s *serviceImpl) Activity() ActivityService {
	return s.activityService
}

// Registration 返回 RegistrationService 实例
func (s *serviceImpl) Registration() RegistrationService {
	return s.registrationService
}

// AdminService, ActivityService, RegistrationService 接口定义将分别放在各自的文件中
