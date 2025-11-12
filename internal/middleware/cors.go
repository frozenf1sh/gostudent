package middleware

import (
	"log/slog"
	"time"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetCors() gin.HandlerFunc {
	var config = cors.Config{
		// **AllowOrigins**: 允许访问的源列表（必填，不能和 AllowAllOrigins = true 同时使用）
		// 允许来自 http://127.0.0.1:5173 的请求
		AllowOrigins: config.GlobalConfig.Cors.AllowOrigins,

		// **AllowMethods**: 允许的 HTTP 方法
		AllowMethods: config.GlobalConfig.Cors.AllowMethods,

		// **AllowHeaders**: 允许的请求头
		// 通常需要允许 Content-Type、Authorization（用于 token/JWT）等
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "Accept"},

		// **ExposeHeaders**: 允许浏览器访问的响应头（非必需）
		ExposeHeaders: []string{"Content-Length"},

		// **AllowCredentials**: 是否允许携带 Cookie 或认证信息
		// 如果设置为 true，AllowOrigins 中就不能使用通配符 "*"
		AllowCredentials: true,

		// **MaxAge**: 预检请求（Preflight Request，OPTIONS 方法）的缓存时间
		// 在此时间内，浏览器不需要再次发送 OPTIONS 请求
		MaxAge: 12 * time.Hour,

		// **AllowOriginFunc**: 可以使用函数来动态判断是否允许该源
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "http://some-dynamic-domain.com"
		// },
	}
	slog.Info("Cors允许的源", "AllowOrigins", config.AllowOrigins)
	slog.Info("Cors允许的方法", "AllowMethods", config.AllowMethods)
	return cors.New(config)
}
