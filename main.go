package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/frozenf1sh/gostudent/internal/handler"
	"github.com/frozenf1sh/gostudent/internal/repository"
	"github.com/frozenf1sh/gostudent/internal/router"
	"github.com/frozenf1sh/gostudent/internal/service"
	"github.com/frozenf1sh/gostudent/pkg/fishlogger"
	"github.com/frozenf1sh/gostudent/pkg/redis"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	// 错误处理
	err error

	// 全局数据库对象
	db *gorm.DB

	// Gin Engine
	r *gin.Engine
)

func main() {
	// 读取服务器配置
	config.InitConfig()

	// 双通道Logger初始化
	fishlogger.LogInit()

	// 数据库
	db = repository.GormInit()

	// JWT令牌加载
	utils.InitJWT()

	// Redis初始化
	err = redis.InitRedis()
	if err != nil {
		slog.Error("Redis初始化失败", "reason", err.Error())
		panic("Redis初始化失败")
	}
	slog.Info("Redis初始化成功")

	// 4. 依赖注入 (DI): 从底层到顶层依次组装

	// --- 4.1. 注入 Repositories (数据访问层) ---
	adminRepo := repository.NewAdminRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)

	// --- 4.2. 注入 Services (业务逻辑层) ---
	adminSvc := service.NewAdminService(adminRepo)
	activitySvc := service.NewActivityService(db, activityRepo)                           // ActivityService 需要 db 来处理事务
	registrationSvc := service.NewRegistrationService(db, activityRepo, registrationRepo) // RegistrationService 涉及活动和报名两个 Repo

	// --- 4.3. 注入 Handlers (API 接口层) ---
	adminH := handler.NewAdminHandler(adminSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	registrationH := handler.NewRegistrationHandler(registrationSvc)
	dashboardH := handler.NewDashboardHandler(db, activityRepo, registrationRepo)

	// 初始化超级管理员
	initSuperAdmin(adminSvc)

	// 启动活动状态自动更新任务
	activitySvc.StartActivityStatusUpdater(context.Background(), config.GlobalConfig.ActivityStatusUpdateInterval)

	// Web服务
	gin.SetMode(gin.ReleaseMode)
	// 创建路由
	r = router.InitRouter(adminH, activityH, registrationH, dashboardH)

	// 监听host和端口
	var (
		serverHost = config.GlobalConfig.Server.Host
		serverPort = config.GlobalConfig.Server.Port
	)
	if err = r.Run(serverHost + ":" + strconv.Itoa(serverPort)); err != nil {
		slog.Error("Gin 启动失败", "reason", err.Error())
		panic("Gin 启动失败")
	}
}

func initSuperAdmin(adminSvc service.AdminService) {
	defaultUsername := config.GlobalConfig.Admin.Username // 假设配置中有这个字段
	defaultPassword := config.GlobalConfig.Admin.Password
	admin, _ := adminSvc.GetByID(context.Background(), 1)
	if admin == nil {
		slog.Info("未找到任何管理员，开始创建默认超级管理员...")

		// 2. 如果不存在，则创建
		if defaultUsername == "" || defaultPassword == "" {
			slog.Error("admin初始化配置缺失", "reason", " 必须提供默认用户名和密码")
			os.Exit(1)
		}

		err := adminSvc.CreateAdmin(context.Background(), defaultUsername, defaultPassword)
		if err != nil {
			slog.Error("创建默认超级管理员失败", "reason", err)
			os.Exit(1)
		}
		slog.Info("默认超级管理员创建成功", "username", defaultUsername)
	} else {
		slog.Info("超级管理员已存在")
	}
}
