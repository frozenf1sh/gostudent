package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/internal/service"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RegistrationHandler 接口定义活动报名和签到相关的 API 方法
type RegistrationHandler interface {
	Register(c *gin.Context)
	ListRegistrations(c *gin.Context)
	SignIn(c *gin.Context) // 签到功能 (目前禁用)
	GetRegistrationByID(c *gin.Context)
}

type registrationHandlerImpl struct {
	svc service.RegistrationService
}

// NewRegistrationHandler 创建 RegistrationHandler 实例
func NewRegistrationHandler(svc service.RegistrationService) RegistrationHandler {
	return &registrationHandlerImpl{svc: svc}
}

// Register godoc
// @Summary 参与者报名活动
// @Description 用户通过活动ID和个人信息进行报名
// @Tags Registration
// @Accept json
// @Produce json
// @Param activity_id path int true "活动ID"
// @Param request body model.CreateRegistrationRequest true "报名请求"
// @Success 200 {object} model.RegistrationResponse "报名成功，返回报名记录"
// @Failure 400 {object} gin.H "请求参数错误或活动ID格式错误"
// @Failure 409 {object} gin.H "重复报名或人数已满"
// @Failure 500 {object} gin.H "内部系统错误"
// @Router /activities/{activity_id}/register [post]
func (h *registrationHandlerImpl) Register(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	var req model.CreateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 调用 Service 层报名业务逻辑
	registration, err := h.svc.Register(c, uint(activityID), &req)
	if err != nil {
		slog.Error("Failed to register for activity", "activity_id", activityID, "error", err)

		// 检查 Service 层定义的业务错误
		if errors.Is(err, service.ErrRegistrationDuplicate) {
			utils.Error(c, http.StatusConflict, "您已报名该活动，请勿重复操作")
			return
		}
		if errors.Is(err, service.ErrRegistrationMaxed) {
			utils.Error(c, http.StatusConflict, "抱歉，该活动报名人数已满")
			return
		}
		if errors.Is(err, service.ErrRegistrationNotOpen) {
			utils.Error(c, http.StatusForbidden, "该活动报名未开始或已截止")
			return
		}
		// 其他错误（如活动不存在、数据库错误）
		utils.Error(c, http.StatusInternalServerError, "报名失败: "+err.Error())
		return
	}

	// 报名成功，返回报名记录
	utils.Success(c, registration)
}

// ListRegistrations godoc
// @Summary 管理员获取活动报名列表
// @Description 管理员根据活动ID获取该活动的所有报名记录
// @Tags Registration
// @Security ApiKeyAuth
// @Produce json
// @Param activity_id path int true "活动ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(10)
// @Success 200 {object} gin.H{list=[]model.RegistrationResponse,total=int} "报名记录列表和总数"
// @Failure 400 {object} gin.H "请求参数错误"
// @Failure 500 {object} gin.H "内部系统错误"
// @Router /admin/activities/{activity_id}/registrations [get]
func (h *registrationHandlerImpl) ListRegistrations(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 调用 Service 层查询逻辑
	list, total, err := h.svc.ListRegistrationsByActivityID(c, uint(activityID), page, pageSize)
	if err != nil {
		slog.Error("Failed to list registrations", "activity_id", activityID, "error", err)
		utils.Error(c, http.StatusInternalServerError, "查询报名列表失败: "+err.Error())
		return
	}

	// 返回结果
	utils.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
	})
}

// GetRegistrationByID godoc
// @Summary 获取单条报名记录详情
// @Description 管理员根据报名记录ID获取详情
// @Tags Registration
// @Security ApiKeyAuth
// @Produce json
// @Param registration_id path int true "报名记录ID"
// @Success 200 {object} model.RegistrationResponse "报名记录详情"
// @Failure 404 {object} gin.H "报名记录不存在"
// @Router /admin/registrations/{registration_id} [get]
func (h *registrationHandlerImpl) GetRegistrationByID(c *gin.Context) {
	registrationIDStr := c.Param("registration_id")
	registrationID, err := strconv.ParseUint(registrationIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "报名ID格式错误")
		return
	}

	registration, err := h.svc.GetRegistrationByID(c, uint(registrationID))
	if err != nil {
		if errors.Is(err, service.ErrRegistrationNotFound) {
			utils.Error(c, http.StatusNotFound, "报名记录不存在")
			return
		}
		slog.Error("Failed to get registration", "id", registrationID, "error", err)
		utils.Error(c, http.StatusInternalServerError, "查询报名记录失败")
		return
	}

	// 转换 model.Registration 为 model.RegistrationResponse DTO
	response := model.RegistrationResponse{
		ID:                 registration.ID,
		ActivityID:         registration.ActivityID,
		ParticipantName:    registration.ParticipantName,
		ParticipantPhone:   registration.ParticipantPhone,
		ParticipantCollege: registration.ParticipantCollege,
		RegisteredAt:       registration.RegisteredAt,
		IsSignedIn:         registration.IsSignedIn,
	}

	utils.Success(c, response)
}

// SignIn godoc
// @Summary 参与者签到
// @Description 参与者通过活动ID和手机号进行签到（此功能已禁用）
// @Tags Registration
// @Accept json
// @Produce json
// @Param activity_id path int true "活动ID"
// @Param request body model.SignInRequest true "签到请求，包含手机号"
// @Success 200 {object} gin.H "签到成功"
// @Router /activities/{activity_id}/signin [post]
func (h *registrationHandlerImpl) SignIn(c *gin.Context) {
	// 根据要求，签到部分暂时设置为无效
	utils.Error(c, http.StatusNotImplemented, "签到功能暂未开放")

	/* 完整的调用逻辑应为:
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "活动ID格式错误")
		return
	}

	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 假设 Service.SignIn 需要 token 或其他验证机制，此处简单传递空字符串
	err = h.svc.SignIn(c, uint(activityID), req.Phone, "")
	if err != nil {
		slog.Error("Failed to sign in", "activity_id", activityID, "phone", req.Phone, "error", err)
		utils.Error(c, http.StatusInternalServerError, "签到失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "签到成功"})
	*/
}
