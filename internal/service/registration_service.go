package service

import (
	"context"
	"errors"
	"time"

	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrRegistrationDuplicate = errors.New("you have already registered for this activity")
	ErrRegistrationMaxed     = errors.New("registration count has reached the maximum limit")
	ErrRegistrationNotOpen   = errors.New("registration is not currently open")
	ErrRegistrationNotFound  = errors.New("registration record not found") // 新增错误：报名记录未找到
)

// 接口：报名业务逻辑接口
type RegistrationService interface {
	// 报名业务
	Register(ctx context.Context, activityID uint, req *model.CreateRegistrationRequest) (*model.Registration, error) // 核心事务逻辑
	// 列出活动报名者
	ListRegistrationsByActivityID(ctx context.Context, activityID uint, page, pageSize int) ([]*model.Registration, int64, error)
	// 多条件查询报名记录
	ListRegistrations(ctx context.Context, params *model.ListRegistrationsParams) ([]*model.Registration, int64, error)
	// SignIn 签到逻辑
	SignIn(ctx context.Context, activityID uint, phone string, token string) error
	// 获取单条报名记录详情 (新增)
	GetRegistrationByID(ctx context.Context, registrationID uint) (*model.Registration, error)
	// UpdateSignInStatusByAdmin 管理员更新签到状态 (新增)
	UpdateSignInStatusByAdmin(ctx context.Context, registrationID uint, isSignedIn bool) error
}

type registrationServiceImpl struct {
	db               *gorm.DB // 用于启动事务
	activityRepo     repository.ActivityRepository
	registrationRepo repository.RegistrationRepository
}

// NewRegistrationService 创建 RegistrationService 实例
func NewRegistrationService(db *gorm.DB, aRepo repository.ActivityRepository, rRepo repository.RegistrationRepository) RegistrationService {
	return &registrationServiceImpl{
		db:               db,
		activityRepo:     aRepo,
		registrationRepo: rRepo,
	}
}

// Register 处理参与者报名活动的核心事务逻辑
func (s *registrationServiceImpl) Register(ctx context.Context, activityID uint, req *model.CreateRegistrationRequest) (*model.Registration, error) {
	// 用于返回报名响应
	var newRegistration *model.Registration
	// 使用自动事务确保报名和人数更新的原子性
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 获取活动信息 (使用事务锁)
		activity, err := s.activityRepo.WithTx(tx).FindByIDForUpdate(ctx, activityID)
		if err != nil {
			// 如果活动不存在或数据库错误，事务回滚
			return ErrActivityNotFound
		}

		// 2. 状态校验：检查活动是否处于报名中状态
		if activity.Status != model.ActivityStatusPublished {
			return ErrRegistrationNotOpen
		}

		// 3. 时间校验：检查是否过了报名截止时间
		if time.Now().After(activity.RegistrationDeadline) {
			return ErrActivityRegistrationOver
		}

		// 4. 重复报名校验：检查该手机号是否已报名
		existingReg, err := s.registrationRepo.WithTx(tx).FindByActivityAndPhone(ctx, activityID, req.ParticipantPhone)
		if err != nil {
			return err
		}
		if existingReg != nil {
			return ErrRegistrationDuplicate
		}

		// 5. 人数上限校验
		if activity.MaxParticipants > 0 && activity.RegisteredCount >= activity.MaxParticipants {
			return ErrRegistrationMaxed
		}

		// 6. 创建报名记录
		registration := &model.Registration{
			ActivityID:         activityID,
			ParticipantName:    req.ParticipantName,
			ParticipantPhone:   req.ParticipantPhone,
			ParticipantCollege: req.ParticipantCollege,
		}
		if err := s.registrationRepo.WithTx(tx).Create(ctx, registration); err != nil {
			return err
		}
		newRegistration = registration // 记录新创建的报名对象以便返回

		// 7. 更新活动已报名人数 (核心更新)
		activity.RegisteredCount += 1
		if err := s.activityRepo.WithTx(tx).Update(ctx, activity); err != nil {
			return err
		}

		// 事务提交
		return nil
	})

	if err != nil {
		return nil, err
	}

	return newRegistration, nil
}

// ListRegistrationsByActivityID 列出某个活动的报名记录
func (s *registrationServiceImpl) ListRegistrationsByActivityID(ctx context.Context, activityID uint, page, pageSize int) ([]*model.Registration, int64, error) {
	// 校验 activityID 权限和存在性（通常在 handler/service.activity.GetByID 中完成）
	return s.registrationRepo.ListByActivityID(ctx, activityID, page, pageSize)
}

// ListRegistrations 多条件查询报名记录
func (s *registrationServiceImpl) ListRegistrations(ctx context.Context, params *model.ListRegistrationsParams) ([]*model.Registration, int64, error) {
	// 权限校验可以根据实际业务需求添加
	return s.registrationRepo.List(ctx, params)
}

// GetRegistrationByID 根据ID获取单条报名记录详情 (新增实现)
func (s *registrationServiceImpl) GetRegistrationByID(ctx context.Context, registrationID uint) (*model.Registration, error) {
	// 假设 registrationRepo 接口包含 FindByID 方法
	reg, err := s.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		// 如果是 GORM 记录未找到错误，则返回 Service 层的 ErrRegistrationNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegistrationNotFound
		}
		// 其他数据库错误
		return nil, err
	}
	return reg, nil
}

// UpdateSignInStatusByAdmin 管理员更新签到状态实现
func (s *registrationServiceImpl) UpdateSignInStatusByAdmin(ctx context.Context, registrationID uint, isSignedIn bool) error {
	// 1. 检查报名记录是否存在
	reg, err := s.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegistrationNotFound
		}
		return err
	}

	// 2. 如果已经是目标状态，无需更新
	if reg.IsSignedIn == isSignedIn {
		return nil
	}

	// 3. 更新签到状态
	now := time.Now()
	return s.registrationRepo.UpdateSignInStatus(ctx, registrationID, isSignedIn, now)
}

func (s *registrationServiceImpl) SignIn(ctx context.Context, activityID uint, phone string, token string) error {
	// 1. 检查活动是否正在进行中
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("活动不存在")
		}
		return errors.New("查询活动失败: " + err.Error())
	}

	now := time.Now()
	if now.Before(activity.StartTime) || now.After(activity.EndTime) {
		return errors.New("当前时间不在活动时间范围内，无法签到")
	}

	// 2. 查找报名记录
	reg, err := s.registrationRepo.FindByActivityAndPhone(ctx, activityID, phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("报名记录未找到或手机号错误")
		}
		return errors.New("查询报名记录失败: " + err.Error())
	}
	if reg == nil {
		return errors.New("报名记录未找到或手机号错误")
	}

	// 3. 检查是否已签到
	if reg.IsSignedIn {
		return errors.New("您已签到，请勿重复操作")
	}

	// 4. 更新签到状态和时间
	err = s.registrationRepo.UpdateSignInStatus(ctx, reg.ID, true, now)
	if err != nil {
		return errors.New("更新签到状态失败: " + err.Error())
	}
	return nil
}
