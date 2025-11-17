package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/frozenf1sh/gostudent/internal/config"
	"github.com/redis/go-redis/v9"
)

// Client Redis客户端
var Client *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() error {
	// 创建Redis客户端配置
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.GlobalConfig.Redis.Host, config.GlobalConfig.Redis.Port),
		Password: config.GlobalConfig.Redis.Password,
		DB:       config.GlobalConfig.Redis.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}

	Client = rdb
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// GenerateAndStoreToken 生成并存储临时Token到Redis
// 参数:
//   activityID: 活动ID
//   duration: Token有效期
// 返回:
//   token: 生成的Token
//   err: 错误信息
func GenerateAndStoreToken(ctx context.Context, activityID uint, duration time.Duration) (string, error) {
	// 检查Redis中是否已有该活动的有效Token
	existingToken, err := Client.Get(ctx, getTokenKey(activityID)).Result()
	if err == nil {
		// Token存在且有效，直接返回
		return existingToken, nil
	} else if err != redis.Nil {
		// 发生了非"key不存在"的错误
		return "", err
	}

	// 生成新的随机Token
	token := generateRandomToken()

	// 存储到Redis
	err = Client.Set(ctx, getTokenKey(activityID), token, duration).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyToken 验证Token是否有效
// 参数:
//   activityID: 活动ID
//   token: 要验证的Token
// 返回:
//   bool: Token是否有效
//   err: 错误信息
func VerifyToken(ctx context.Context, activityID uint, token string) (bool, error) {
	// 从Redis获取Token
	storedToken, err := Client.Get(ctx, getTokenKey(activityID)).Result()
	if err != nil {
		if err == redis.Nil {
			// Token不存在或已过期
			return false, nil
		}
		// 发生了其他错误
		return false, err
	}

	// 比较Token
	return storedToken == token, nil
}

// getTokenKey 生成Redis中存储Token的Key
func getTokenKey(activityID uint) string {
	return fmt.Sprintf("signin:token:%d", activityID)
}

// generateRandomToken 生成随机Token
func generateRandomToken() string {
	// 使用crypto/rand生成安全的随机字符串
	// 生成16字节的随机数据，转换为32位的十六进制字符串
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 发生错误时返回一个默认值（在实际应用中应该处理此错误）
		return "default-token"
	}
	return hex.EncodeToString(b)
}