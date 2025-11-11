package fishlogger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm/logger" // 导入 logger.Interface 等日志类型
)

// --- GIN 日志适配器 (保持不变) ---

// GinSlogAdapter 实现了 io.Writer 接口，用于将 Gin 的访问日志桥接到 slog。
type GinSlogAdapter struct {
	Logger *slog.Logger
}

// NewGinSlogAdapter 创建 Gin 适配器实例。
func NewGinSlogAdapter(logger *slog.Logger) *GinSlogAdapter {
	return &GinSlogAdapter{Logger: logger}
}

// Write 实现了 io.Writer 接口。
func (w *GinSlogAdapter) Write(p []byte) (n int, err error) {
	// 移除 Gin 日志末尾的换行符和空格
	msg := string(bytes.TrimSpace(p))

	// 默认使用 Info 级别记录访问日志，并添加来源字段
	w.Logger.Info(
		msg,
		slog.String("source", "gin"),
	)
	return len(p), nil
}

// --- GORM 日志适配器 ---

// GormSlogAdapter 实现了 logger.Interface 接口。
type GormSlogAdapter struct {
	Logger *slog.Logger
	Level  logger.LogLevel
}

// NewGormSlogAdapter 创建 GORM 适配器实例。
func NewGormSlogAdapter(slogger *slog.Logger) *GormSlogAdapter {
	return &GormSlogAdapter{
		Logger: slogger.With("source", "gorm"),
		Level:  logger.Info, // 默认 Info 级别
	}
}

// LogMode 实现了 gorm.Logger 接口的 LogMode 方法。
func (l *GormSlogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.Level = level
	return &newLogger
}

// Trace 实现了 logger.Interface 接口的 Trace 方法。
// 修正：fc 的签名改为 func() (string, int64)
func (l *GormSlogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Level <= logger.Silent {
		return
	}

	// 调用 fc() 获取 SQL 语句和影响的行数 (修正: 不再传递参数)
	sql, rows := fc()
	elapsed := time.Since(begin)

	logLevel := slog.LevelInfo

	// 日志级别判断逻辑：
	if err != nil && l.Level >= logger.Error {
		logLevel = slog.LevelError
	} else if elapsed > time.Duration(100)*time.Millisecond && l.Level >= logger.Warn { // 示例慢查询阈值 100ms
		logLevel = slog.LevelWarn
	} else if l.Level == logger.Info {
		logLevel = slog.LevelInfo
	} else {
		return
	}

	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Duration("latency_ms", elapsed),
		slog.Int64("rows_affected", rows),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	// 修正：使用 ...any 展开属性切片
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	// 使用 slog.Log 记录结构化数据
	l.Logger.Log(ctx, logLevel, "SQL 语句已执行", args...)
}

// Info, Warn, Error 方法的简化实现
func (l *GormSlogAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Level >= logger.Info {
		l.Logger.InfoContext(ctx, fmt.Sprintf(msg, data...), slog.String("gorm_msg_type", "info"))
	}
}
func (l *GormSlogAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Level >= logger.Warn {
		l.Logger.WarnContext(ctx, fmt.Sprintf(msg, data...), slog.String("gorm_msg_type", "warn"))
	}
}
func (l *GormSlogAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Level >= logger.Error {
		l.Logger.ErrorContext(ctx, fmt.Sprintf(msg, data...), slog.String("gorm_msg_type", "error"))
	}
}
