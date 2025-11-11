// config/config.go

package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config 结构体：定义所有配置的映射
type Config struct {
	// 服务端设置
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`

	// 数据库设置
	Database struct {
		Driver   string `mapstructure:"driver"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		Password string `mapstructure:"password"`
	} `mapstructure:"database"`

	// JWT 配置
	JWT struct {
		Secret    string        `mapstructure:"secret"`     // JWT 密钥
		ExpiresIn time.Duration `mapstructure:"expires_in"` // Token 有效期
	} `mapstructure:"jwt"`

	LogFile string `mapstructure:"log_file"`
}

// GlobalConfig 是程序的全局配置实例
var GlobalConfig Config

// InitConfig 初始化配置，按照优先级：环境变量 > 配置文件 > 默认值
func InitConfig() {
	v := viper.New()

	// 配置文件设置
	v.SetConfigName("config") // 配置文件名 (不带扩展名)
	v.SetConfigType("yaml")   // 配置文件类型
	v.AddConfigPath(".")      // 查找配置文件路径（当前目录）

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalln("配置文件 config.yaml 不存在")
		} else {
			log.Fatalf("配置文件读取错误: %s \n", err)
		}
	} else {
		log.Println("配置文件 config.yaml 成功加载")
	}

	// 允许环境变量覆盖
	v.AutomaticEnv()
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("database.host", "DB_HOST")

	// 将 Viper 中的配置反序列化到全局结构体
	if err := v.Unmarshal(&GlobalConfig); err != nil {
		log.Fatalf("viper反序列化失败: %v", err)
	}

	log.Println("配置文件成功初始化")
}
