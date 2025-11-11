package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/frozenf1sh/gostudent/pkg/utils"
	"github.com/gin-gonic/gin"
)

// --- 常量定义 (简化起见，定义在 middleware 包内) ---

// ContextKeyAdminID 用于存储 AdminID 的 Context Key
const ContextKeyAdminID = "admin_id"

// Claims Data Key: 令牌类型
const ClaimsDataKeyType = "type"

// Claims Data Key: Admin ID
const ClaimsDataKeyAdminID = "admin_id"

// Claim Data Value: 管理员令牌类型的值
const ClaimTypeAdmin = "admin_login"

// JWTAuthAdmin 是一个用于校验 Admin JWT Token 的 Gin 中间件
// 职责: 校验签名、过期时间、令牌类型，并提取 AdminID
func JWTAuthAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Header 中获取 Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "请求头中缺少 Authorization 信息")
			c.Abort()
			return
		}

		// 检查格式是否为 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, http.StatusUnauthorized, "Token 格式错误，应为 Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 2. 通用解析和校验 Token
		claims, err := utils.ParseGenericJWT(tokenString)
		if err != nil {
			slog.Warn("JWT Token 解析失败", "error", err)
			utils.Error(c, http.StatusUnauthorized, "无效或过期 Token")
			c.Abort()
			return
		}

		// 3. 校验 Token 类型 (解耦后的业务逻辑：确保是管理员令牌)
		claimType, ok := claims.Data[ClaimsDataKeyType].(string)
		if !ok || claimType != ClaimTypeAdmin {
			slog.Warn("JWT Token 类型错误或缺失", "type", claimType)
			utils.Error(c, http.StatusForbidden, "令牌类型错误，非管理员令牌")
			c.Abort()
			return
		}

		// 4. 提取 AdminID 并处理类型转换
		// **注意：由于 MapClaims 使用 map[string]any，JSON 解析器会将数字解析为 float64**
		adminIDFloat, ok := claims.Data[ClaimsDataKeyAdminID].(float64)
		if !ok {
			slog.Error("AdminID 字段缺失或格式错误", "data", claims.Data)
			utils.Error(c, http.StatusForbidden, "令牌数据结构错误 (AdminID 字段缺失或非数字)")
			c.Abort()
			return
		}

		// 转换为业务所需的 uint 类型
		adminID := uint(adminIDFloat)

		// 5. 将 AdminID 存储在 Context 中，供后续 Handler 使用
		c.Set(ContextKeyAdminID, adminID)

		// 可选：将完整的 claims 存储在 Context 中，供需要原始数据的业务层使用
		// c.Set("jwt_claims_map", claims.Data)

		// Token 校验成功，继续处理请求
		c.Next()
	}
}
