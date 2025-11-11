package repository

import (
	"context"

	"github.com/frozenf1sh/gostudent/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 接口：活动仓库接口
type ActivityRepository interface {
	// 返回一个使用事务的仓库实例
	WithTx(tx *gorm.DB) ActivityRepository

	// CUD
	Create(ctx context.Context, activity *model.Activity) error
	Update(ctx context.Context, activity *model.Activity) error
	Delete(ctx context.Context, id uint) error

	// 通过唯一id查找
	FindByID(ctx context.Context, id uint) (*model.Activity, error)
	// 通过ID查找并锁定行，用于事务
	FindByIDForUpdate(ctx context.Context, id uint) (*model.Activity, error)
	// List 列出活动 (带过滤和分页)
	List(ctx context.Context, params *model.ListActivitiesParams) ([]*model.Activity, int64, error)
}

// ----- 实现 -----
// 实现了 ActivityRepository 接口
type activityRepositoryImpl struct {
	// 可以是事务
	db *gorm.DB
}

// 构造函数
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepositoryImpl{db: db}
}

// 接受一个事务，返回一个基于该事务的实例
// service 层调用: txRepo := activityRepo.WithTx(tx)
func (r *activityRepositoryImpl) WithTx(tx *gorm.DB) ActivityRepository {
	// 返回一个新的实例，它持有事务 *gorm.DB
	return &activityRepositoryImpl{db: tx}
}

// Create 创建活动
func (r *activityRepositoryImpl) Create(ctx context.Context, activity *model.Activity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// Update 更新活动
func (r *activityRepositoryImpl) Update(ctx context.Context, activity *model.Activity) error {
	// 使用 Gorm 的 Save 方法来更新所有字段
	// 如果使用 Update，需要用 map[string]interface{} 来更新，或者用 Select()
	return r.db.WithContext(ctx).Save(activity).Error
}

// Delete 删除活动
func (r *activityRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Activity{}, id).Error
}

// FindByID 通过ID查找
func (r *activityRepositoryImpl) FindByID(ctx context.Context, id uint) (*model.Activity, error) {
	var activity model.Activity
	if err := r.db.WithContext(ctx).Preload("Admin").First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// FindByIDForUpdate 通过ID查找并使用 "FOR UPDATE" 锁
// 悲观锁查找，在查询时阻止其他并发事务修改这条记录，直到当前事务提交或回滚
// 必须在事务(WithTx)中调用
func (r *activityRepositoryImpl) FindByIDForUpdate(ctx context.Context, id uint) (*model.Activity, error) {
	var activity model.Activity
	// GORM V2 使用 clause.Locking{Strength: "UPDATE"}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// List 列出活动 (带过滤和分页)
func (r *activityRepositoryImpl) List(ctx context.Context, params *model.ListActivitiesParams) ([]*model.Activity, int64, error) {
	var activities []*model.Activity
	var total int64

	// 创建两个独立查询构建器，一个计数，一个分页查找
	query := r.db.WithContext(ctx).Model(&model.Activity{})
	countQuery := r.db.WithContext(ctx).Model(&model.Activity{})

	// 应用过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
		countQuery = countQuery.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
		countQuery = countQuery.Where("status = ?", params.Status)
	}
	if !params.DateFrom.IsZero() {
		query = query.Where("start_time >= ?", params.DateFrom)
		countQuery = countQuery.Where("start_time >= ?", params.DateFrom)
	}
	if !params.DateTo.IsZero() {
		query = query.Where("start_time <= ?", params.DateTo)
		countQuery = countQuery.Where("start_time <= ?", params.DateTo)
	}

	// 1. 获取总数 (在应用分页前)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. 应用分页和排序
	offset := (params.Page - 1) * params.PageSize
	query = query.Order("created_at DESC").Limit(params.PageSize).Offset(offset)

	// 3. 执行查询
	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}
