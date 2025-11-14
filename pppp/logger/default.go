package logger

import (
	"go.uber.org/zap"
)

// 全局默认 logger
var defaultLogger ILogger

// InitDefault 初始化默认 logger
func InitDefault(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	defaultLogger = logger

	return nil
}

// GetDefault 获取默认 logger
func GetDefault() ILogger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		logger, _ := NewLogger(DefaultConfig())
		defaultLogger = logger
	}

	return defaultLogger
}

func Debug(msg string, fields ...zap.Field) {
	GetDefault().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetDefault().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetDefault().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetDefault().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetDefault().Fatal(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	GetDefault().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	GetDefault().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	GetDefault().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	GetDefault().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	GetDefault().Fatalf(template, args...)
}
