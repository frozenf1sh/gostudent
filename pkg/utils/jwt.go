package utils

import (
	"errors"
	"time"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// jwtSecretKey 密钥 (所有令牌共享，但可以根据需要配置多个)
	jwtSecretKey []byte
	// jwtIssuer 签发者
	jwtIssuer = "xdu-activity-system"
)

// InitJWT 初始化 JWT 配置
func InitJWT() {
	// 假设配置中有一个通用的 JWT Secret
	jwtSecretKey = []byte(config.GlobalConfig.JWT.Secret)
	// 注意：这里不再初始化过期时间，而是让业务层传入
}

// MapClaims 通用的 Claims 结构体，用于封装业务数据
type MapClaims struct {
	Data map[string]any `json:"data,omitempty"` // 用于存储 AdminID, ActivityID 等业务数据
	jwt.RegisteredClaims
}

// GenerateGenericJWT 通用生成 JWT Token
// 参数:
//
//	data map[string]any: 包含业务数据的映射 (例如 {"admin_id": 1})
//	expiresIn time.Duration: Token 的有效时长
func GenerateGenericJWT(data map[string]any, expiresIn time.Duration) (string, error) {
	if len(jwtSecretKey) == 0 {
		return "", errors.New("JWT secret key is not initialized")
	}
	if expiresIn <= 0 {
		return "", errors.New("expiresIn must be a positive duration")
	}

	// 1. 创建 Claims
	claims := MapClaims{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
		},
	}

	// 2. 签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseGenericJWT 通用解析 JWT Token
// 返回: 包含业务数据的 MapClaims 结构体
func ParseGenericJWT(tokenString string) (*MapClaims, error) {
	// 1. 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &MapClaims{}, func(token *jwt.Token) (any, error) {
		// 确保签名方法是 HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 2. 验证 token 并获取 claims
	if claims, ok := token.Claims.(*MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
