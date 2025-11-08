package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ILogger 日志接口定义
type ILogger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)

	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)
	Panicf(template string, args ...any)

	With(fields ...zap.Field) ILogger
	WithPrefix(prefix string) ILogger
	Sync() error
}

// Logger 日志实例
type Logger struct {
	// logger 极速吧日志句柄
	logger *zap.Logger
	// sugar 高级功能的logger句柄
	sugar *zap.SugaredLogger
	// config 日志配置
	config *Config
}

// NewLogger 创建新的日志实例
func NewLogger(config *Config) (ILogger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 解析日志级别
	level := parseLevel(config.Level)

	// 创建编码器配置
	encoderConfig := DefaultEncoderConfig()

	// 根据环境选择编码器
	encoder := returnZapCoreEncoder(config, encoderConfig)

	// 创建写入器
	var cores []zapcore.Core

	// 控制台输出
	if config.Console {
		consoleEncoder := encoderConfig
		if config.Environment != "prod" {
			consoleEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoder),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// 文件输出
	if config.File {
		// 确保日志目录存在
		if err := os.MkdirAll(config.FilePath, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}

		// 所有级别日志文件
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(returnLoggerConfig(config, "")),
			level,
		))

		// 错误日志单独文件
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(returnLoggerConfig(config, "error.log")),
			zapcore.ErrorLevel,
		))
	}

	// 合并所有 core
	core := zapcore.NewTee(cores...)

	// 创建 logger 选项
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if !config.EnableCaller {
		options = []zap.Option{}
	}

	if config.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 创建 logger，使用 Named 设置服务名称
	zapLogger := zap.New(core, options...).Named(strings.ToUpper(config.ServiceName) + ":" + strings.ToUpper(config.Environment))

	return &Logger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
		config: config,
	}, nil
}

func returnZapCoreEncoder(config *Config, encoderConfig zapcore.EncoderConfig) (encoder zapcore.Encoder) {
	if config.Environment == "prod" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	return
}

// Debug 输出 Debug 级别日志
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info 输出 Info 级别日志
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn 输出 Warn 级别日志
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error 输出 Error 级别日志
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 输出 Fatal 级别日志并退出程序
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// Panic 输出 Panic 级别日志并触发 panic
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.logger.Panic(msg, fields...)
}

// Debugf 格式化输出 Debug 级别日志
func (l *Logger) Debugf(template string, args ...any) {
	l.sugar.Debugf(template, args...)
}

// Infof 格式化输出 Info 级别日志
func (l *Logger) Infof(template string, args ...any) {
	l.sugar.Infof(template, args...)
}

// Warnf 格式化输出 Warn 级别日志
func (l *Logger) Warnf(template string, args ...any) {
	l.sugar.Warnf(template, args...)
}

// Errorf 格式化输出 Error 级别日志
func (l *Logger) Errorf(template string, args ...any) {
	l.sugar.Errorf(template, args...)
}

// Fatalf 格式化输出 Fatal 级别日志并退出程序
func (l *Logger) Fatalf(template string, args ...any) {
	l.sugar.Fatalf(template, args...)
}

// Panicf 格式化输出 Panic 级别日志并触发 panic
func (l *Logger) Panicf(template string, args ...any) {
	l.sugar.Panicf(template, args...)
}

// With 添加字段到日志上下文
func (l *Logger) With(fields ...zap.Field) ILogger {
	newLogger := l.logger.With(fields...)

	return &Logger{
		logger: newLogger,
		sugar:  newLogger.Sugar(),
		config: l.config,
	}
}

// WithPrefix 添加前缀到日志
func (l *Logger) WithPrefix(prefix string) ILogger {
	return l.With(zap.String("prefix", prefix))
}

// Sync 刷新缓冲区
func (l *Logger) Sync() error {
	return l.logger.Sync()
}
