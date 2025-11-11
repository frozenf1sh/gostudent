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
)

// RegistrationService 定义报名业务逻辑接口
type RegistrationService interface {
	Register(ctx context.Context, activityID uint, req *model.CreateRegistrationRequest) (*model.Registration, error) // 核心事务逻辑
	ListRegistrationsByActivityID(ctx context.Context, activityID uint, page, pageSize int) ([]*model.Registration, int64, error)
	// (可选) SignIn 签到逻辑
	// SignIn(ctx context.Context, activityID uint, phone string, token string) error
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
	// 使用事务确保报名和人数更新的原子性
	var newRegistration *model.Registration

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
