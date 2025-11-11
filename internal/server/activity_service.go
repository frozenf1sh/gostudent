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
	ErrActivityNotFound         = errors.New("activity not found")
	ErrActivityAlreadyPublished = errors.New("activity is already published or ended")
	ErrActivityRegistrationOver = errors.New("registration deadline has passed")
	ErrActivityIsRunning        = errors.New("activity is already running or finished")
)

// ActivityService 定义活动业务逻辑接口
type ActivityService interface {
	CreateActivity(ctx context.Context, adminID uint, req *model.CreateActivityRequest) (*model.Activity, error)
	GetActivityByID(ctx context.Context, id uint) (*model.Activity, error)
	ListActivities(ctx context.Context, params *model.ListActivitiesParams) ([]*model.Activity, int64, error)
	UpdateActivity(ctx context.Context, id uint, req *model.UpdateActivityRequest) error
	DeleteActivity(ctx context.Context, id uint) error
	PublishActivity(ctx context.Context, id uint) error // 发布活动 (核心功能之一)
}

type activityServiceImpl struct {
	db           *gorm.DB // 用于事务
	activityRepo repository.ActivityRepository
}

// NewActivityService 创建 ActivityService 实例
func NewActivityService(db *gorm.DB, repo repository.ActivityRepository) ActivityService {
	return &activityServiceImpl{
		db:           db,
		activityRepo: repo,
	}
}

// CreateActivity 创建活动
func (s *activityServiceImpl) CreateActivity(ctx context.Context, adminID uint, req *model.CreateActivityRequest) (*model.Activity, error) {
	// 1. 基本校验：报名截止时间不能晚于活动开始时间
	if req.RegistrationDeadline.After(req.StartTime) {
		return nil, errors.New("registration deadline must be before activity start time")
	}

	// 2. DTO -> Model 转换
	activity := &model.Activity{
		AdminID:              adminID,
		Title:                req.Title,
		Type:                 req.Type,
		Description:          req.Description,
		StartTime:            req.StartTime,
		Location:             req.Location,
		RegistrationDeadline: req.RegistrationDeadline,
		MaxParticipants:      req.MaxParticipants,
		LiveURL:              req.LiveURL,
		AttachmentURL:        req.AttachmentURL,
		// 状态默认为 DRAFT
		Status: model.ActivityStatusDraft,
	}

	// 3. 调用 Repository 存储
	if err := s.activityRepo.Create(ctx, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

// PublishActivity 发布活动，将状态从 DRAFT 变为 PUBLISHED
func (s *activityServiceImpl) PublishActivity(ctx context.Context, id uint) error {
	// 1. 获取活动
	activity, err := s.activityRepo.FindByID(ctx, id)
	if err != nil {
		return ErrActivityNotFound
	}

	// 2. 状态校验
	if activity.Status != model.ActivityStatusDraft {
		return ErrActivityAlreadyPublished
	}

	// 3. 时间校验：活动开始时间不能已过
	if activity.StartTime.Before(time.Now()) {
		return ErrActivityIsRunning
	}
	// 4. 报名截止时间不能已过
	if activity.RegistrationDeadline.Before(time.Now()) {
		return ErrActivityRegistrationOver
	}

	// 5. 更新状态
	activity.Status = model.ActivityStatusPublished

	// 6. 调用 Repository 更新
	return s.activityRepo.Update(ctx, activity)

	// Note: 实际项目中，这里还可以调用 utils/qrcode.go 来生成二维码并保存链接
}

// GetActivityByID 获取单个活动详情
func (s *activityServiceImpl) GetActivityByID(ctx context.Context, id uint) (*model.Activity, error) {
	activity, err := s.activityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrActivityNotFound
	}
	return activity, nil
}

// ListActivities 列出活动列表
func (s *activityServiceImpl) ListActivities(ctx context.Context, params *model.ListActivitiesParams) ([]*model.Activity, int64, error) {
	return s.activityRepo.List(ctx, params)
}

// UpdateActivity 更新活动（此处省略删除和更新的详细实现，原理类似 GetByID + Update）
func (s *activityServiceImpl) UpdateActivity(ctx context.Context, id uint, req *model.UpdateActivityRequest) error {
	// 1. 查找活动
	activity, err := s.activityRepo.FindByID(ctx, id)
	if err != nil {
		return ErrActivityNotFound
	}

	// 2. 检查活动是否在允许修改的状态 (例如，PUBLISHED 状态后只允许修改部分字段)
	if activity.Status != model.ActivityStatusDraft && activity.Status != model.ActivityStatusPublished {
		return errors.New("cannot update activity in current status")
	}

	// 3. DTO -> Model 赋值 (只更新非空字段)
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.MaxParticipants != nil {
		activity.MaxParticipants = *req.MaxParticipants
	}
	// ... 其他字段的更新

	// 4. 调用 Repository 更新
	return s.activityRepo.Update(ctx, activity)
}

// DeleteActivity 删除活动
func (s *activityServiceImpl) DeleteActivity(ctx context.Context, id uint) error {
	// TODO: 考虑删除活动的连锁反应（报名记录）。如果使用 Gorm 外键约束 ON DELETE CASCADE，则会自动删除。
	return s.activityRepo.Delete(ctx, id)
}
