package logger

import (
	"time"

	"go.uber.org/zap"
)

// Middleware 日志中间件接口
type Middleware func(next func()) func()

// WithRequestID 添加请求 ID 中间件
func WithRequestID(logger ILogger, requestID string) ILogger {
	return logger.With(zap.String("request_id", requestID))
}

// WithUserID 添加用户 ID 中间件
func WithUserID(logger ILogger, userID string) ILogger {
	return logger.With(zap.String("user_id", userID))
}

// WithTraceID 添加追踪 ID 中间件(用于分布式追踪)
func WithTraceID(logger ILogger, traceID string) ILogger {
	return logger.With(zap.String("trace_id", traceID))
}

// WithSpanID 添加 Span ID 中间件
func WithSpanID(logger ILogger, spanID string) ILogger {
	return logger.With(zap.String("span_id", spanID))
}

// WithDuration 记录执行时间的中间件
func WithDuration(logger ILogger, operation string, fn func()) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logger.Info("操作完成",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}()
	fn()
}

// WithRecover 捕获 panic 的中间件
func WithRecover(logger ILogger, fn func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("捕获到 panic",
				zap.Any("error", err),
				zap.Stack("stacktrace"),
			)
		}
	}()

	fn()
}

// ContextFields 上下文字段构建器
type ContextFields struct {
	fields []zap.Field
}

// NewContextFields 创建新的上下文字段构建器
func NewContextFields() *ContextFields {
	return &ContextFields{
		fields: make([]zap.Field, 0),
	}
}

// Add 添加字段
func (c *ContextFields) Add(field zap.Field) *ContextFields {
	c.fields = append(c.fields, field)

	return c
}

// AddString 添加字符串字段
func (c *ContextFields) AddString(key, value string) *ContextFields {
	c.fields = append(c.fields, zap.String(key, value))

	return c
}

// AddInt 添加整数字段
func (c *ContextFields) AddInt(key string, value int) *ContextFields {
	c.fields = append(c.fields, zap.Int(key, value))

	return c
}

// AddBool 添加布尔字段
func (c *ContextFields) AddBool(key string, value bool) *ContextFields {
	c.fields = append(c.fields, zap.Bool(key, value))

	return c
}

// AddFloat64 添加浮点字段
func (c *ContextFields) AddFloat64(key string, value float64) *ContextFields {
	c.fields = append(c.fields, zap.Float64(key, value))

	return c
}

// AddAny 添加任意类型字段
func (c *ContextFields) AddAny(key string, value any) *ContextFields {
	c.fields = append(c.fields, zap.Any(key, value))

	return c
}

// Build 构建字段数组
func (c *ContextFields) Build() []zap.Field {
	return c.fields
}

// ApplyToLogger 应用到 logger
func (c *ContextFields) ApplyToLogger(logger ILogger) ILogger {
	return logger.With(c.fields...)
}
