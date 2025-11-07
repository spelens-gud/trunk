package logger

import (
	"path/filepath"
	"time"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置
type Config struct {
	// ServiceName 服务名称(用作日志前缀)
	ServiceName string
	// Environment 环境：dev, test, prod
	Environment string
	// Level 日志级别：debug, info, warn, error
	Level string
	// Console 是否输出到控制台
	Console bool
	// File 是否输出到文件
	File bool
	// FilePath 文件路径
	FilePath string
	// FileName 文件名
	FileName string
	// MaxSize 单个文件最大大小（MB）
	MaxSize int
	// MaxAge 保留旧文件的最大天数
	MaxAge int
	// MaxBackups 保留旧文件的最大个数
	MaxBackups int
	// Compress 是否压缩旧文件
	Compress bool
	// EnableCaller 是否启用调用者信息
	EnableCaller bool
	// EnableStacktrace 是否启用堆栈跟踪
	EnableStacktrace bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ServiceName:      "root",
		Environment:      "dev",
		Level:            "info",
		Console:          true,
		File:             false,
		FilePath:         "./logs",
		FileName:         "root.log",
		MaxSize:          100,
		MaxAge:           30,
		MaxBackups:       10,
		Compress:         true,
		EnableCaller:     true,
		EnableStacktrace: true,
	}
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// DefaultEncoderConfig 返回默认的编码器配置
func DefaultEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:     "time",
		LevelKey:    "level",
		NameKey:     "logger",
		CallerKey:   "caller",
		FunctionKey: zapcore.OmitKey,
		MessageKey:  "msg",
		// StacktraceKey: 堆栈跟踪
		StacktraceKey: "stacktrace",
		// LineEnding: 换行符为\n
		LineEnding: zapcore.DefaultLineEnding,
		// EncodeLevel: 自定义日志级别, 大写编码
		EncodeLevel: zapcore.CapitalLevelEncoder,
		// EncodeTime: 自定义时间格式
		EncodeTime: customTimeEncoder,
		// EncodeDuration: 默认为 zapcore.SecondsDurationEncoder, 浮点数显示为秒
		EncodeDuration: zapcore.SecondsDurationEncoder,
		// EncodeCaller: 默认为 zapcore.ShortCallerEncoder, 只保留最后一个文件调用
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
}

// customTimeEncoder 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func returnLoggerConfig(config *Config, fileName string) *lumberjack.Logger {
	if fileName == "" {
		fileName = config.FileName
	}

	return &lumberjack.Logger{
		Filename:   filepath.Join(config.FilePath, fileName),
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
		LocalTime:  true,
	}
}
