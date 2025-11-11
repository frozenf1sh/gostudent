package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/internal/service"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AdminHandler 接口定义管理员操作的 API 方法
type AdminHandler interface {
	Login(c *gin.Context)
	// GetAdminInfo(c *gin.Context) // 例如: 获取当前登录管理员信息
}

type adminHandlerImpl struct {
	svc service.AdminService
}

// NewAdminHandler 创建 AdminHandler 实例
func NewAdminHandler(svc service.AdminService) AdminHandler {
	return &adminHandlerImpl{svc: svc}
}

// Login godoc
// @Summary 管理员登录
// @Description 使用用户名和密码获取 JWT Token
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body model.AdminLoginRequest true "登录请求"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} gin.H "请求参数错误"
// @Failure 401 {object} gin.H "用户名或密码错误"
// @Router /admin/login [post]
func (h *adminHandlerImpl) Login(c *gin.Context) {
	var req model.AdminLoginRequest
	// 1. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("登录参数绑定错误", "error", err)
		utils.Error(c, http.StatusBadRequest, "请求参数不完整或格式错误")
		return
	}

	// 2. 调用 Service 层业务逻辑
	token, err := h.svc.Login(c, &req)

	// 3. 处理业务逻辑错误
	if err != nil {
		slog.Error("Admin login failed", "username", req.Username, "error", err)
		if errors.Is(err, service.ErrInvalidPassword) { // 假设 Service 定义了该错误
			utils.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		} else {
			utils.Error(c, http.StatusInternalServerError, "登录失败: "+err.Error())
		}
		return
	}

	// 4. 返回成功响应
	utils.Success(c, model.AdminLoginResponse{Token: token, Username: req.Username})
}
