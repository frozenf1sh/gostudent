package service

import (
	"context"
	"errors"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/internal/repository"
	"github.com/frozenf1sh/gostudent/pkg/utils" // 假设 utils 包中包含 JWT 和 Hash 函数
)

var (
	ErrAdminNotFound   = errors.New("admin not found")
	ErrInvalidPassword = errors.New("invalid username or password")
)

// 接口：管理员业务逻辑接口
type AdminService interface {
	// 登录，返回 JWT token
	Login(ctx context.Context, req *model.AdminLoginRequest) (string, error)
	// 通过 ID 获取
	GetByID(ctx context.Context, id uint) (*model.Admin, error)
	// CreateAdmin 用于初始化超级管理员 (通常只在 setup 阶段运行一次)
	CreateAdmin(ctx context.Context, username, password string) error
}

// 接口实现
type adminServiceImpl struct {
	adminRepo repository.AdminRepository
}

// NewAdminService 创建 AdminService 实例
func NewAdminService(repo repository.AdminRepository) AdminService {
	return &adminServiceImpl{adminRepo: repo}
}

// CreateAdmin 仅用于项目初始化，创建第一个管理员
func (s *adminServiceImpl) CreateAdmin(ctx context.Context, username, password string) error {
	// 1. 哈希密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	// 2. 构造模型
	admin := &model.Admin{
		Username:     username,
		PasswordHash: hashedPassword,
	}

	// 3. 存储到数据库
	return s.adminRepo.Create(ctx, admin)
}

// Login 处理管理员登录逻辑
func (s *adminServiceImpl) Login(ctx context.Context, req *model.AdminLoginRequest) (string, error) {
	// 1. 通过用户名查找管理员
	admin, err := s.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		// 统一返回错误，避免暴露用户是否存在的信息
		return "", ErrInvalidPassword
	}
	if admin == nil {
		return "", ErrInvalidPassword
	}

	// 2. 验证密码
	if !utils.CheckPasswordHash(req.Password, admin.PasswordHash) {
		return "", ErrInvalidPassword
	}

	// 3. 生成 JWT Token
	data := map[string]any{
		"type":     "admin_login", // 业务层定义的类型标签
		"admin_id": admin.ID,
	}
	token, err := utils.GenerateGenericJWT(data, config.GlobalConfig.JWT.AdminExpiresIn)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetByID 通过 ID 获取管理员信息
func (s *adminServiceImpl) GetByID(ctx context.Context, id uint) (*model.Admin, error) {
	return s.adminRepo.FindByID(ctx, id)
}
