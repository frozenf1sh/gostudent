package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/frozenf1sh/gostudent/internal/model"
	"github.com/frozenf1sh/gostudent/pkg/fishlogger"
	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// 错误处理
	err error

	logConfigs []fishlogger.LogChannelConfig
	appLogger  *slog.Logger

	// Gorm数据库
	db *gorm.DB

	// Gin Engine
	router *gin.Engine
)

func main() {
	// 读取服务器配置
	config.InitConfig()
	// JWT令牌加载
	utils.InitJWT()
	// 双通道Log初始化
	logInit()

	// 数据库
	gormInit()

	// Web服务
	gin.SetMode(gin.ReleaseMode)
	ginInit()
}

func logInit() {
	logConfigs = []fishlogger.LogChannelConfig{
		{
			// 标准输出（Stdout）通道
			DestinationPath: "stdout",
			MinLevel:        slog.LevelDebug, // 调试级别以上都输出
			Format:          "text",          // 文本格式
		},
		{
			// 文件通道
			DestinationPath: config.GlobalConfig.LogFile,
			MinLevel:        slog.LevelInfo, // 信息级别以上才写入文件
			Format:          "json",         // JSON 格式
			// MaxLevel:        slog.LevelWarn, // 仅写入 Info 和 Warn 级别的日志
		},
	}

	// 初始化logger并设为默认
	appLogger = fishlogger.NewMultiChannelLogger(
		logConfigs,
		// slog.String("app_name", "FishApp"), // 添加默认属性
	)
	fishlogger.SetDefaultLogger(appLogger)

	slog.Info("日志器已成功初始化")
	slog.Debug("可以在stdout中输出调试信息")
}

func gormInit() {
	// 拼接dsn
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%v)/xdu_activity?charset=utf8mb4&parseTime=True&loc=Local",
		config.GlobalConfig.Database.Name,
		config.GlobalConfig.Database.Password,
		config.GlobalConfig.Database.Host,
		config.GlobalConfig.Database.Port)

	// 连接数据库，并设置gorm日志
	gormAdapter := fishlogger.NewGormSlogAdapter(appLogger)
	db, err = gorm.Open(mysql.Open(DSN), &gorm.Config{
		Logger:      gormAdapter.LogMode(logger.Info),
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	//设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("无法获取底层DB对象", "reason", err)
	}
	sqlDB.SetMaxIdleConns(25)                 // 最大允许的空闲连接
	sqlDB.SetMaxOpenConns(100)                // 最大连接数
	sqlDB.SetConnMaxLifetime(20 * time.Hour)  // 示例：连接可复用 1 小时}
	sqlDB.SetConnMaxIdleTime(4 * time.Minute) // 示例：连接可复用 1 小时}

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		slog.Error("连接数据库失败", "reason", err)
		panic("连接数据库失败")
	}

	slog.Info("已连接到数据库")

	err = db.AutoMigrate(&model.Admin{})
	if err != nil {
		slog.Error("数据库自动迁移失败", "reason", err)
		os.Exit(1)
	}
}

func ginInit() {
	// 创建空Engine
	router = gin.New()

	// 添加Logger
	ginAdapter := fishlogger.NewGinSlogAdapter(appLogger)
	router.Use(gin.LoggerWithWriter(ginAdapter))

	// 恢复器
	router.Use(gin.Recovery())

	// 注册自定义验证器
	setupValidator()

	// === API 路由组 ===
	api := router.Group("/api")
	{
		// GET 请求：/api/hello
		api.GET("/hello", func(c *gin.Context) {
			// 模拟一个轻微的网络延迟，让加载状态更明显
			time.Sleep(time.Millisecond * 500)

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "👋 恭喜！这是来自 Gin 后端的数据！",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
		})

		api.POST("/register", RegisterHandler)
	}

	// 监听host和端口
	var (
		serverHost = config.GlobalConfig.Server.Host
		serverPort = config.GlobalConfig.Server.Port
	)
	if err = router.Run(serverHost + ":" + strconv.Itoa(serverPort)); err != nil {
		slog.Error("Gin 启动失败", "reason", err.Error())
		panic("Gin 启动失败")
	}
}

// --- 1. 自定义验证函数 ---
func MobileValidator(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	// 简单的 11 位数字正则匹配
	pattern := `^\d{11}$`
	match, _ := regexp.MatchString(pattern, mobile)
	return match
}

// --- 2. 注册验证器 ---
func setupValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证规则，名称为 "mobile"
		v.RegisterValidation("mobile", MobileValidator)
	}
}

// --- 3. 结构体定义 ---
type RegisterForm struct {
	Username string `json:"username" binding:"required,min=5,max=15"`
	Mobile   string `json:"mobile" binding:"required,mobile"`
	Password string `json:"password" binding:"required"`
}

// --- 4. 错误处理辅助函数 ---
func getErrorMsg(fe validator.FieldError) string {
	// 这是一个简化版的错误信息翻译，生产环境通常会使用 i18n
	switch fe.Tag() {
	case "required":
		return fe.Field() + " 字段不能为空"
	case "min":
		return fe.Field() + " 字段长度/值小于最小值要求 (>=5)"
	case "max":
		return fe.Field() + " 字段长度/值大于最大值要求 (<=15)"
	case "mobile":
		return fe.Field() + " 格式不符合手机号码规范 (11位数字)"
	default:
		return fe.Field() + " 字段校验失败 (标签：" + fe.Tag() + ")"
	}
}

// --- 5. Handler 函数 ---
func RegisterHandler(c *gin.Context) {
	var form RegisterForm

	if err := c.ShouldBindJSON(&form); err != nil {
		// 类型断言：判断错误是否为 validator.ValidationErrors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMap := make(map[string]string)

			for _, fieldError := range validationErrors {
				// fieldError 包含了验证失败的字段、规则、期望值等信息
				fieldName := fieldError.Field()
				errorMap[fieldName] = getErrorMsg(fieldError)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "参数校验失败",
				"data": errorMap,
			})
			return
		}

		// 绑定失败，但不是 ValidationErrors (如 JSON 解析错误)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求数据格式或类型错误",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "注册成功", "data": form})
}
