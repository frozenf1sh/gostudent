package service

import (
	"context"
	"errors"
	"time"

	"log/slog"

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
	StartActivityStatusUpdater(ctx context.Context, interval time.Duration)
}

type activityServiceImpl struct {
	db           *gorm.DB // 用于事务
	activityRepo repository.ActivityRepository
}

// StartActivityStatusUpdater 启动活动状态自动更新定时任务（建议在 main.go 初始化时调用）
func (s *activityServiceImpl) StartActivityStatusUpdater(ctx context.Context, interval time.Duration) {
	slog.Info("活动状态自动更新任务已启动")
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("活动状态自动更新任务已停止")
				return
			case t := <-ticker.C:
				closed, finished, err := s.activityRepo.UpdateStatusByDeadline(ctx)
				if err != nil {
					slog.Error("活动状态自动更新失败", "err", err)
				} else {
					slog.Debug("活动状态自动更新", "time", t, "closed_count", closed, "finished_count", finished)
				}
			}
		}
	}()
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
		return nil, errors.New("报名截止时间不能晚于活动开始时间")
	} else if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("结束时间不能晚于开始时间")
	}

	// 2. DTO -> Model 转换
	activity := &model.Activity{
		AdminID:              adminID,
		Title:                req.Title,
		Type:                 req.Type,
		Description:          req.Description,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
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

// UpdateActivity 完整更新活动逻辑
func (s *activityServiceImpl) UpdateActivity(ctx context.Context, id uint, req *model.UpdateActivityRequest) error {
	// 1. 查找活动
	activity, err := s.activityRepo.FindByID(ctx, id)
	if err != nil {
		return ErrActivityNotFound
	}

	// 2. 检查活动是否在允许修改的状态
	if activity.Status == model.ActivityStatusFinished {
		return errors.New("无法修改已结束的活动")
	}

	// 3. DTO -> Model 赋值 (只更新非空字段)

	// A. 字符串类型更新
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Type != nil { // 修复：Type
		activity.Type = *req.Type
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.Location != nil { // 修复：Location
		activity.Location = *req.Location
	}
	if req.LiveURL != nil { // 修复：LiveURL
		activity.LiveURL = *req.LiveURL
	}
	if req.AttachmentURL != nil { // 修复：AttachmentURL
		activity.AttachmentURL = *req.AttachmentURL
	}

	// B. 数值类型更新
	if req.MaxParticipants != nil {
		activity.MaxParticipants = *req.MaxParticipants
	}

	// C. 时间类型更新 (使用临时变量来执行时间校验)
	newStartTime := activity.StartTime
	if req.StartTime != nil {
		newStartTime = *req.StartTime // 修复：StartTime
	}
	newEndTime := activity.EndTime
	if req.EndTime != nil {
		newEndTime = *req.EndTime // 修复：StartTime
	}

	newDeadline := activity.RegistrationDeadline
	if req.RegistrationDeadline != nil {
		newDeadline = *req.RegistrationDeadline // 修复：RegistrationDeadline
	}

	// D. 业务逻辑校验：报名截止时间不能晚于活动开始时间
	if newDeadline.After(newStartTime) {
		return errors.New("报名截止时间不能晚于活动开始时间")
	} else if newEndTime.Before(newStartTime) {
		return errors.New("结束时间不能晚于开始时间")
	}

	// 如果校验通过，才赋值回 activity model
	activity.StartTime = newStartTime
	activity.EndTime = newEndTime
	activity.RegistrationDeadline = newDeadline

	// E. 状态更新
	if req.Status != nil {
		// 修复：Status 字段赋值和类型转换
		newStatus := model.ActivityStatus(*req.Status)

		// 简单的状态值校验
		if newStatus != model.ActivityStatusDraft &&
			newStatus != model.ActivityStatusPublished &&
			newStatus != model.ActivityStatusClosed &&
			newStatus != model.ActivityStatusFinished {
			return errors.New("invalid status value")
		}
		activity.Status = newStatus
	}

	// 4. 调用 Repository 更新
	return s.activityRepo.Update(ctx, activity)
}

// DeleteActivity 删除活动
func (s *activityServiceImpl) DeleteActivity(ctx context.Context, id uint) error {
	// 考虑删除活动的连锁反应（报名记录）。如果使用 Gorm 外键约束 ON DELETE CASCADE，则会自动删除。
	return s.activityRepo.Delete(ctx, id)
}
