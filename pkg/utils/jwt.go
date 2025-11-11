package utils

import (
	"errors"
	"time"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// jwtSecretKey 密钥
	jwtSecretKey []byte
	// jwtExpiresIn Token 有效期
	jwtExpiresIn time.Duration // 72 小时
)

func InitJWT() {
	jwtSecretKey = []byte(config.GlobalConfig.JWT.Secret)
	jwtExpiresIn = config.GlobalConfig.JWT.ExpiresIn
}

// CustomClaims 自定义 JWT claims
type CustomClaims struct {
	AdminID uint `json:"admin_id"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT Token
func GenerateJWT(adminID uint) (string, error) {
	// 创建 claims
	claims := CustomClaims{
		AdminID: adminID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xdu-activity-system",
		},
	}

	// 使用 HS256 签名算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWT 解析 JWT Token
func ParseJWT(tokenString string) (*CustomClaims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		// 确保签名方法是 HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 token 并获取 claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
