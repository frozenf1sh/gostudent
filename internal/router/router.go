package router

import (
	"net/http"

	"github.com/frozenf1sh/gostudent/internal/handler"
	"github.com/frozenf1sh/gostudent/internal/middleware"
	"github.com/frozenf1sh/gostudent/pkg/fishlogger"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 Gin 路由器并设置所有 API 路由
func InitRouter(
	adminH handler.AdminHandler,
	activityH handler.ActivityHandler,
	registrationH handler.RegistrationHandler,
	dashboardH handler.DashboardHandler, // 新增参数
) *gin.Engine {
	// 创建 Gin 实例
	r := gin.New()

	// 1. 设置全局中间件
	// 添加Logger
	ginAdapter := fishlogger.NewGinSlogAdapter(fishlogger.AppLogger)
	r.Use(gin.LoggerWithWriter(ginAdapter))
	// 恢复器
	r.Use(gin.Recovery())
	// 跨域处理 (CORS)
	// r.Use(middleware.GetCors())

	// Ping 接口，用于健康检查
	r.GET("/ping", func(c *gin.Context) {
		utils.Success(c, "pong")
	})

	// =========================================================
	// 2. Public Group (公开接口: 无需认证)
	// =========================================================
	publicGroup := r.Group("/api/v1")
	{
		// P1 & P2: 活动查询
		publicGroup.GET("/activities", activityH.ListActivities)
		// 修正: 将 :id 统一为 :activity_id 以匹配 Handler 中的 c.Param("activity_id")
		publicGroup.GET("/activities/:activity_id", activityH.GetActivityByID)

		// P3 & P4: 活动报名与签到 (路径已规范)
		publicGroup.POST("/activities/:activity_id/register", registrationH.Register)
		publicGroup.POST("/activities/:activity_id/signin", registrationH.SignIn)

		// 获取签到Token
		publicGroup.GET("/activities/:activity_id/signin-token", activityH.GetSignInToken)

		// A1: 管理员登录 (唯一一个在 Public Group 中的 Admin 接口)
		publicGroup.POST("/admin/login", adminH.Login)
	}

	// =========================================================
	// 3. Admin Group (管理接口: 需要 JWT 认证)
	// =========================================================
	// 假设 middleware.JWTAuth 是我们用于校验 Token 的中间件
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.JWTAuthAdmin()) // 应用 JWT 认证中间件
	{
		// A2 - A6: 活动管理 (CRUD + 发布)
		adminGroup.POST("/activities", activityH.CreateActivity)
		adminGroup.GET("/activities", activityH.ListActivities) // A5: 管理员查询所有活动（包含草稿等状态）

		// 修正: 将 :id 统一为 :activity_id 以匹配 Handler 中的 c.Param("activity_id")
		adminGroup.GET("/activities/:activity_id", activityH.GetActivityByID)
		adminGroup.PUT("/activities/:activity_id", activityH.UpdateActivity)
		adminGroup.DELETE("/activities/:activity_id", activityH.DeleteActivity)

		// 修正: 将 :id/publish 统一为 :activity_id/publish
		adminGroup.POST("/activities/:activity_id/publish", activityH.PublishActivity)

		// 移除adminGroup中的签到Token端点

		// A7 & A8: 报名记录管理 (路径已规范)
		adminGroup.GET("/activities/:activity_id/registrations", registrationH.ListRegistrations)
		adminGroup.GET("/registrations/:registration_id", registrationH.GetRegistrationByID) // A8

		// A9: Admin面板统计信息
		adminGroup.GET("/dashboard", dashboardH.GetDashboardData) // 仪表盘统计接口
	}

	// 4. 处理 404 错误
	r.NoRoute(func(c *gin.Context) {
		utils.Error(c, http.StatusNotFound, "找不到该路由")
	})

	return r
}
