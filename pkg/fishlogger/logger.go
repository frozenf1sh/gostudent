package fishlogger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

// --- MultiHandler 实现 slog.Handler 接口 (保持不变) ---

// MultiHandler 包含一个子 Handler 列表，负责将日志记录分发给所有启用的子 Handler。
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler 创建 MultiHandler 实例。
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// Enabled 只要其中一个子 Handler 启用该级别，就返回 true。
func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range m.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle 处理 Record。它遍历所有子 Handler，如果子 Handler 被启用，则调用其 Handle 方法。
func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range m.handlers {
		if handler.Enabled(ctx, r.Level) {
			// 忽略子 Handle 产生的错误，遵循 slog 的标准行为
			_ = handler.Handle(ctx, r)
		}
	}
	return nil
}

// WithAttrs 返回一个新的 MultiHandler，新 Handler 的所有子 Handler 都带上了新的属性。
func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, handler := range m.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(newHandlers...)
}

// WithGroup 返回一个新的 MultiHandler，新 Handler 的所有子 Handler 都带上了新的组名。
func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, handler := range m.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return NewMultiHandler(newHandlers...)
}

// --- 新增：可配置的初始化函数和结构 ---

// LogChannelConfig 定义单个日志通道（Handler）的配置。
type LogChannelConfig struct {
	// DestinationPath: 输出目标。如果为空，则使用 os.Stdout。
	// 如果是文件路径，则创建并输出到文件。
	DestinationPath string

	// MinLevel: 该通道接受的最低日志级别。
	MinLevel slog.Level

	// Format: 输出格式，可以是 "text" 或 "json"。
	Format string

	// MaxLevel: 可选，该通道接受的最高日志级别。
	// 如果设置为 MaxLevel，则使用自定义过滤逻辑。
	MaxLevel slog.Level
}

// NewMultiChannelLogger 根据配置创建并返回一个 Multi-channel Logger。
// 它可以设置为全局默认 Logger，或者用于特定场景。
func NewMultiChannelLogger(
	configs []LogChannelConfig,
	defaultAttrs ...slog.Attr,
) *slog.Logger {
	handlers := make([]slog.Handler, 0, len(configs))

	for _, cfg := range configs {
		var writer io.Writer
		if cfg.DestinationPath == "" || cfg.DestinationPath == "stdout" {
			writer = os.Stdout
		} else if cfg.DestinationPath == "stderr" {
			writer = os.Stderr
		} else {
			// 文件输出
			logFile, err := os.OpenFile(
				cfg.DestinationPath,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0666,
			)
			if err != nil {
				// 如果文件创建失败，严重警告并跳过该通道
				log.Printf("Warning: Failed to open log file %s: %v. Channel skipped.", cfg.DestinationPath, err)
				continue
			}
			writer = logFile
		}

		options := &slog.HandlerOptions{
			// MinLevel 设置为该通道的最低级别
			Level: cfg.MinLevel,
			// AddSource: true,
		}

		// --- 新增：专门用于Text Handler的简洁格式化 ---
		var formatReplaceAttr func([]string, slog.Attr) slog.Attr
		if cfg.Format == "text" {
			// 定义Text格式的ReplaceAttr函数
			formatReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
				// 1. 格式化时间戳 (TimeKey)
				if a.Key == slog.TimeKey {
					t := a.Value.Any().(time.Time)
					a.Value = slog.StringValue(t.Format("2006/01/02 15:04:05"))
					return a
				}

				// 2. 格式化级别 (LevelKey)
				if a.Key == slog.LevelKey {
					// 将级别值转换为 [LEVEL] 格式，例如: [INFO]
					a.Value = slog.StringValue(fmt.Sprintf("[%s]", a.Value.String()))
					return a
				}

				// 3. SourceKey
				if a.Key == slog.SourceKey {
					a.Value = slog.StringValue(fmt.Sprintf("[%s]", a.Value.String()))
					return a
				}

				// 4. 丢弃 MessageKey 和所有非核心属性
				if a.Key != slog.MessageKey && len(groups) == 0 {
					// 丢弃所有顶层附加属性（如 app_name, sql, latency_ms, component 等）
					// 确保 Text 输出最简洁，只留下 time, level, msg
					return slog.Attr{}
				}

				return a
			}
		}
		// --- 结束：新增简洁格式化 ---

		// 实现 MaxLevel 过滤逻辑
		if cfg.MaxLevel != 0 && cfg.MaxLevel >= cfg.MinLevel {
			// 在这里，我们需要组合 formatReplaceAttr 和 MaxLevel 过滤
			maxLevel := cfg.MaxLevel

			// 如果同时有格式化和MaxLevel过滤，需要组合它们
			originalReplaceAttr := formatReplaceAttr // 如果是Text格式，originalReplaceAttr 就是格式化函数
			if cfg.Format != "text" {
				// 如果不是Text格式，但有MaxLevel过滤，则没有格式化函数
				originalReplaceAttr = nil
			}

			options.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
				// 1. 如果有原有的 ReplaceAttr (比如Text格式化函数)，先执行它
				if originalReplaceAttr != nil {
					a = originalReplaceAttr(groups, a)
				}
				// 如果上一步返回了空属性 (格式化时被丢弃)，则直接返回
				if a.Key == "" && a.Value.Kind() == 0 {
					return a
				}

				// 2. MaxLevel 过滤逻辑
				if a.Key == slog.LevelKey {
					levelVal := a.Value.Any().(slog.Level)
					if levelVal > maxLevel {
						return slog.Attr{} // 丢弃该日志记录
					}
				}
				return a
			}
		} else if cfg.Format == "text" {
			// 只有Text格式化，没有MaxLevel过滤的情况
			options.ReplaceAttr = formatReplaceAttr
		}

		var handler slog.Handler
		if cfg.Format == "json" {
			handler = slog.NewJSONHandler(writer, options)
		} else {
			// 默认使用 Text 格式
			handler = slog.NewTextHandler(writer, options)
		}

		handlers = append(handlers, handler)
	}

	if len(handlers) == 0 {
		fmt.Println("Warning: No valid log handlers configured. Using default STDOUT Text handler.")
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	combinedHandler := NewMultiHandler(handlers...)

	// 创建 Logger 并添加默认属性
	return slog.New(combinedHandler.WithAttrs(defaultAttrs))
}

// SetDefaultLogger 是一个辅助函数，用于将新创建的 Logger 设置为全局默认 Logger。
func SetDefaultLogger(logger *slog.Logger) {
	slog.SetDefault(logger)
}
