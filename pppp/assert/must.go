package assert

import (
	"fmt"

	"go.uber.org/zap"
)

// ILogger 日志接口定义（避免循环依赖）
type ILogger interface {
	Panic(msg string, fields ...zap.Field)
	Panicf(template string, args ...any)
	Error(msg string, fields ...zap.Field)
	Errorf(template string, args ...any)
}

var (
	// defaultLogger 默认日志记录器（可选）
	defaultLogger ILogger
)

// SetLogger 设置全局日志记录器
func SetLogger(logger ILogger) {
	defaultLogger = logger
}

// logError 记录错误到日志
func logPanic(err error, msg ...any) {
	if defaultLogger == nil {
		return
	}

	if len(msg) > 0 {
		if format, ok := msg[0].(string); ok && len(msg) > 1 {
			defaultLogger.Panicf(format+": %v", append(msg[1:], err)...)
		} else {
			defaultLogger.Panicf("%v: %v", fmt.Sprint(msg...), err)
		}
	} else {
		defaultLogger.Panic(err.Error(), zap.Error(err), zap.Stack("stack"))
	}
}

// mustNoError 断言错误必须为 nil，否则触发 panic
func mustNoError(err error, msg ...any) {
	if err == nil {
		return
	}

	logPanic(err, msg...)

	// 构造panic消息
	if len(msg) > 0 {
		if format, ok := msg[0].(string); ok && len(msg) > 1 {
			panic(fmt.Sprintf(format+": %v", append(msg[1:], err)...))
		} else {
			panic(fmt.Sprintf("%v: %v", fmt.Sprint(msg...), err))
		}
	} else {
		panic(err)
	}
}

// MustCall0E 执行无参数返回error的函数，错误不为nil时panic
func MustCall0E(f func() error, msg ...any) {
	err := f()
	mustNoError(err, msg...)
}

// MustCall0RE 执行无参数返回值和error的函数，错误不为nil时panic
func MustCall0RE[R any](f func() (R, error), msg ...any) R {
	value, err := f()
	mustNoError(err, msg...)
	return value
}

// MustCall1E 执行单参数返回error的函数，错误不为nil时panic
func MustCall1E[T any](f func(T) error, arg T, msg ...any) {
	err := f(arg)
	mustNoError(err, msg...)
}

// MustCall1RE 执行单参数返回值和error的函数，错误不为nil时panic
func MustCall1RE[T any, R any](f func(T) (R, error), arg T, msg ...any) R {
	value, err := f(arg)
	mustNoError(err, msg...)
	return value
}

// MustCall2E 执行双参数返回error的函数，错误不为nil时panic
func MustCall2E[T1, T2 any](f func(T1, T2) error, arg1 T1, arg2 T2, msg ...any) {
	err := f(arg1, arg2)
	mustNoError(err, msg...)
}

// MustCall2RE 执行双参数返回值和error的函数，错误不为nil时panic
func MustCall2RE[T1, T2, R any](f func(T1, T2) (R, error), arg1 T1, arg2 T2, msg ...any) R {
	value, err := f(arg1, arg2)
	mustNoError(err, msg...)
	return value
}

// MustCall3E 执行三参数返回error的函数，错误不为nil时panic
func MustCall3E[T1, T2, T3 any](f func(T1, T2, T3) error, arg1 T1, arg2 T2, arg3 T3, msg ...any) {
	err := f(arg1, arg2, arg3)
	mustNoError(err, msg...)
}

// MustCall3RE 执行三参数返回值和error的函数，错误不为nil时panic
func MustCall3RE[T1, T2, T3, R any](f func(T1, T2, T3) (R, error), arg1 T1, arg2 T2, arg3 T3, msg ...any) R {
	value, err := f(arg1, arg2, arg3)
	mustNoError(err, msg...)
	return value
}
