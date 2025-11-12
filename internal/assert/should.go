package assert

import (
	"fmt"

	"go.uber.org/zap"
)

// logError 记录错误到日志
func logError(err error, msg ...any) {
	if defaultLogger == nil {
		return
	}

	if len(msg) > 0 {
		if format, ok := msg[0].(string); ok && len(msg) > 1 {
			defaultLogger.Errorf(format+": %v", append(msg[1:], err)...)
		} else {
			defaultLogger.Errorf("%v: %v", fmt.Sprint(msg...), err)
		}
	} else {
		defaultLogger.Error(err.Error(), zap.Error(err), zap.Stack("stack"))
	}
}

// shouldNoError 断言错误应该为 nil，否则返回错误并记录日志
func shouldNoError(err error, msg ...any) {
	if err == nil {
		return
	}

	logError(err, msg...)
}

// ShouldCall0E 执行无参数返回error的函数
func ShouldCall0E(f func() error, msg ...any) {
	err := f()
	shouldNoError(err, msg...)
}

// ShouldCall0RE 执行无参数返回值和error的函数
func ShouldCall0RE[R any](f func() (R, error), msg ...any) R {
	value, err := f()
	shouldNoError(err, msg...)
	return value
}

// ShouldCall1E 执行单参数返回error的函数
func ShouldCall1E[T any](f func(T) error, arg T, msg ...any) {
	err := f(arg)
	shouldNoError(err, msg...)
}

// ShouldCall1RE 执行单参数返回值和error的函数
func ShouldCall1RE[T any, R any](f func(T) (R, error), arg T, msg ...any) R {
	value, err := f(arg)
	shouldNoError(err, msg...)
	return value
}

// ShouldCall2E 执行双参数返回error的函数
func ShouldCall2E[T1, T2 any](f func(T1, T2) error, arg1 T1, arg2 T2, msg ...any) {
	err := f(arg1, arg2)
	shouldNoError(err, msg...)
}

// ShouldCall2RE 执行双参数返回值和error的函数
func ShouldCall2RE[T1, T2, R any](f func(T1, T2) (R, error), arg1 T1, arg2 T2, msg ...any) R {
	value, err := f(arg1, arg2)
	shouldNoError(err, msg...)
	return value
}

// ShouldCall3E 执行三参数返回error的函数
func ShouldCall3E[T1, T2, T3 any](f func(T1, T2, T3) error, arg1 T1, arg2 T2, arg3 T3, msg ...any) {
	err := f(arg1, arg2, arg3)
	shouldNoError(err, msg...)
}

// ShouldCall3RE 执行三参数返回值和error的函数
func ShouldCall3RE[T1, T2, T3, R any](f func(T1, T2, T3) (R, error), arg1 T1, arg2 T2, arg3 T3, msg ...any) R {
	value, err := f(arg1, arg2, arg3)
	shouldNoError(err, msg...)
	return value
}
