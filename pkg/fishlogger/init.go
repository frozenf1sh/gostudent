package fishlogger

import (
	"github.com/frozenf1sh/gostudent/internal/config"
	"log/slog"
)

// 全局日志对象
var AppLogger *slog.Logger

func LogInit() {
	logConfigs := []LogChannelConfig{
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
	AppLogger = NewMultiChannelLogger(
		logConfigs,
		// slog.String("app_name", "FishApp"), // 添加默认属性
	)
	SetDefaultLogger(AppLogger)

	slog.Info("日志器已成功初始化")
	slog.Debug("可以在stdout中输出调试信息")
}
