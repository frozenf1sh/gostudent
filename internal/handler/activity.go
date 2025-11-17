package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time" // 引入 time 以便使用 ActivityResponse 结构体

	// 注意：这里需要替换为你项目的实际导入路径
	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/internal/service"
	"github.com/frozenf1sh/gostudent/pkg/redis"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ActivityHandler 接口定义活动相关的 API 方法
type ActivityHandler interface {
	CreateActivity(c *gin.Context)
	ListActivities(c *gin.Context)
	GetActivityByID(c *gin.Context)
	UpdateActivity(c *gin.Context)
	DeleteActivity(c *gin.Context)
	PublishActivity(c *gin.Context)
	GetSignInToken(c *gin.Context)
}

type activityHandlerImpl struct {
	svc service.ActivityService
}

// NewActivityHandler 创建 ActivityHandler 实例
func NewActivityHandler(svc service.ActivityService) ActivityHandler {
	return &activityHandlerImpl{svc: svc}
}

// getAdminIDFromContext 从 Gin Context 中获取 AdminID (假设 JWT 中间件已将其注入)
func getAdminIDFromContext(c *gin.Context) (uint, error) {
	// JWT 中间件应将解析出的 AdminID 存储到 c.Set("admin_id", ID)
	adminID, exists := c.Get("admin_id")
	if !exists {
		return 0, errors.New("admin_id not found in context")
	}
	// 确保类型是 uint
	id, ok := adminID.(uint)
	if !ok {
		return 0, errors.New("admin_id format incorrect in context")
	}
	return id, nil
}

// --- DTO 转换辅助函数 ---

// toActivityResponse 将 model.Activity 转换为 model.ActivityResponse DTO
func toActivityResponse(activity *model.Activity) model.ActivityResponse {
	// 确保 model.ActivityResponse 包含 model.Activity 的所有公共字段
	return model.ActivityResponse{
		ID:                   activity.ID,
		AdminID:              activity.AdminID,
		Title:                activity.Title,
		Type:                 activity.Type,
		Description:          activity.Description,
		StartTime:            activity.StartTime,
		EndTime:              activity.EndTime,
		Location:             activity.Location,
		RegistrationDeadline: activity.RegistrationDeadline,
		MaxParticipants:      activity.MaxParticipants,
		RegisteredCount:      activity.RegisteredCount,
		Status:               activity.Status,
		LiveURL:              activity.LiveURL,
		AttachmentURL:        activity.AttachmentURL,
		CreatedAt:            activity.CreatedAt,
	}
}

// toActivityResponseList 将 []*model.Activity 转换为 []model.ActivityResponse DTO 列表
func toActivityResponseList(activities []*model.Activity) []model.ActivityResponse {
	list := make([]model.ActivityResponse, len(activities))
	for i, activity := range activities {
		list[i] = toActivityResponse(activity)
	}
	return list
}

// CreateActivity 创建活动 (Admin 接口)
// @Summary 创建新活动
// @Tags Activity
// @Accept json
// @Produce json
// @Param request body model.CreateActivityRequest true "活动创建请求"
// @Success 200 {object} model.ActivityResponse "创建的活动详情" // 修正 Swagger
// @Router /admin/activities [post]
func (h *activityHandlerImpl) CreateActivity(c *gin.Context) {
	adminID, err := getAdminIDFromContext(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "未登录或Token无效")
		return
	}

	var req model.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	activity, err := h.svc.CreateActivity(c, adminID, &req)
	if err != nil {
		slog.Error("Failed to create activity", "admin_id", adminID, "error", err)
		utils.Error(c, http.StatusInternalServerError, "创建活动失败: "+err.Error())
		return
	}

	// DTO 转换
	utils.Success(c, toActivityResponse(activity))
}

// ListActivities 获取活动列表 (Admin 或 Public 接口)
// @Summary 获取活动列表
// @Tags Activity
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(10)
// @Param status query string false "活动状态过滤"
// @Success 200 {object} gin.H{list=[]model.ActivityResponse,total=int} "活动列表和总数" // 修正 Swagger
// @Router /activities [get]
func (h *activityHandlerImpl) ListActivities(c *gin.Context) {
	var params model.ListActivitiesParams

	// 使用 ShouldBindQuery 绑定所有查询参数
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.Error(c, http.StatusBadRequest, "查询参数格式错误: "+err.Error())
		return
	}

	// 1. 调用 Service 层列表查询逻辑
	list, total, err := h.svc.ListActivities(c, &params)
	if err != nil {
		slog.Error("Failed to list activities", "error", err, "params", params)
		utils.Error(c, http.StatusInternalServerError, "查询活动列表失败: "+err.Error())
		return
	}

	// 2. DTO 列表转换
	responseList := toActivityResponseList(list)

	// 3. 返回结果
	utils.Success(c, gin.H{
		"list":  responseList,
		"total": total,
		"page":  params.Page,
	})
}

// GetActivityByID 获取活动详情 (Admin & Public)
// @Summary 获取活动详情
// @Tags Activity
// @Produce json
// @Param activity_id path int true "活动ID"
// @Success 200 {object} model.ActivityResponse "活动详情" // 修正 Swagger
// @Failure 404 {object} gin.H "活动不存在"
// @Router /activities/{activity_id} [get]
func (h *activityHandlerImpl) GetActivityByID(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	activity, err := h.svc.GetActivityByID(c, uint(activityID))
	if err != nil {
		slog.Error("Activity not found", "id", activityID, "error", err)
		if errors.Is(err, service.ErrActivityNotFound) {
			utils.Error(c, http.StatusNotFound, "活动不存在")
		} else {
			utils.Error(c, http.StatusInternalServerError, "查询活动失败: "+err.Error())
		}
		return
	}

	// DTO 转换
	utils.Success(c, toActivityResponse(activity))
}

// UpdateActivity 编辑活动 (Admin 接口)
// @Summary 更新活动信息
// @Tags Activity
// @Accept json
// @Produce json
// @Param activity_id path int true "活动ID"
// @Param request body model.UpdateActivityRequest true "活动更新请求"
// @Success 200 {object} model.ActivityResponse "更新后的活动详情" // 修正 Swagger
// @Router /admin/activities/{activity_id} [put]
func (h *activityHandlerImpl) UpdateActivity(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	var req model.UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 调用 Service 更新逻辑
	err = h.svc.UpdateActivity(c, uint(activityID), &req)
	if err != nil {
		slog.Error("Failed to update activity", "id", activityID, "error", err)
		// 检查特定的业务错误
		if errors.Is(err, service.ErrActivityNotFound) {
			utils.Error(c, http.StatusNotFound, "活动不存在")
		} else {
			utils.Error(c, http.StatusInternalServerError, "更新活动失败: "+err.Error())
		}
		return
	}

	// 更新成功后，为了返回最新的数据，再次获取活动详情
	updatedActivity, err := h.svc.GetActivityByID(c, uint(activityID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新成功但查询最新数据失败: "+err.Error())
		return
	}

	// DTO 转换
	utils.Success(c, toActivityResponse(updatedActivity))
}

// DeleteActivity 删除活动 (Admin 接口)
// @Summary 删除活动
// @Tags Activity
// @Produce json
// @Param activity_id path int true "活动ID"
// @Success 200 {object} gin.H "删除成功"
// @Router /admin/activities/{activity_id} [delete]
func (h *activityHandlerImpl) DeleteActivity(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	// 调用 Service 删除逻辑
	if err := h.svc.DeleteActivity(c, uint(activityID)); err != nil {
		slog.Error("Failed to delete activity", "id", activityID, "error", err)
		if errors.Is(err, service.ErrActivityNotFound) {
			utils.Error(c, http.StatusNotFound, "活动不存在或已被删除")
		} else {
			utils.Error(c, http.StatusInternalServerError, "删除活动失败: "+err.Error())
		}
		return
	}

	utils.Success(c, gin.H{"message": "活动删除成功"})
}

// PublishActivity 发布活动 (Admin 接口)
// @Summary 发布活动
// @Tags Activity
// @Produce json
// @Param activity_id path int true "活动ID"
// @Success 200 {object} model.ActivityResponse "发布成功后的活动详情" // 修正 Swagger
// @Router /admin/activities/{activity_id}/publish [post]
func (h *activityHandlerImpl) PublishActivity(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	// 调用 Service 发布逻辑
	err = h.svc.PublishActivity(c, uint(activityID))
	if err != nil {
		slog.Error("Failed to publish activity", "id", activityID, "error", err)
		utils.Error(c, http.StatusInternalServerError, "发布活动失败: "+err.Error())
		return
	}

	// 发布成功后，为了返回 LiveURL/QR等信息，再次获取活动详情
	publishedActivity, err := h.svc.GetActivityByID(c, uint(activityID))
	if err != nil {
		// 尽管发布成功，但查询失败，返回一个提示
		utils.Success(c, gin.H{"message": "活动发布成功，但查询最新详情失败"})
		return
	}

	// DTO 转换
	utils.Success(c, toActivityResponse(publishedActivity))
}

// GetSignInToken godoc
// @Summary 生成活动签到Token
// @Description 为特定活动生成临时签到Token（有效时间30秒，活动期间可获取）
// @Tags Activity
// @Produce json
// @Param activity_id path int true "活动ID"
// @Success 200 {object} gin.H "生成成功，返回签到Token"
// @Failure 400 {object} gin.H "活动ID格式错误"
// @Failure 404 {object} gin.H "活动不存在"
// @Failure 403 {object} gin.H "不在活动时间范围内，无法获取签到Token"
// @Failure 500 {object} gin.H "生成Token失败"
// @Router /api/v1/activities/{activity_id}/signin-token [get]
func (h *activityHandlerImpl) GetSignInToken(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	// 1. 获取活动信息
	activity, err := h.svc.GetActivityByID(c, uint(activityID))
	if err != nil {
		if errors.Is(err, service.ErrActivityNotFound) {
			utils.Error(c, http.StatusNotFound, "活动不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "获取活动信息失败: "+err.Error())
		return
	}

	// 2. 检查是否在活动时间范围内
	now := time.Now()
	if now.Before(activity.StartTime) || now.After(activity.EndTime) {
		utils.Error(c, http.StatusForbidden, "当前时间不在活动时间范围内，无法获取签到Token")
		return
	}

	// 3. 生成并存储签到Token到Redis
	token, err := redis.GenerateAndStoreToken(
		c.Request.Context(),
		uint(activityID),
		config.GlobalConfig.JWT.SignInExpiresIn,
	)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "生成签到Token失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"token": token})
}
