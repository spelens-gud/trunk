package assert

import (
	"fmt"

	"github.com/spelens-gud/trunk/internal/logger"
)

// Should 断言条件应该为真，否则静默失败
// 支持多种消息格式：
//   - Should(true, "错误")           // 单个消息
//   - Should(true, "错误: %v", err)  // 格式化消息
//   - Should(true, err)              // 直接传递 error
func Should(log logger.ILogger, condition bool, msg ...any) {
	if condition {
		return
	}

	// 没有消息时使用默认消息
	if len(msg) == 0 {
		log.Errorf("断言失败")
		return
	}

	// 单个参数处理
	if len(msg) == 1 {
		if err, ok := msg[0].(error); ok {
			log.Errorf(err.Error())
			return
		}
		if format, ok := msg[0].(string); ok {
			log.Errorf("%s", format)
			return
		}

		log.Errorf("%v", msg[0])
		return
	}

	// 多个参数时尝试格式化
	if format, ok := msg[0].(string); ok {
		log.Errorf(format, msg[1:]...)
		return
	}

	// 其他情况使用 Sprint
	log.Errorf("%s", fmt.Sprint(msg...))
}

// shouldNoError 断言错误应该为 nil，否则静默失败
func shouldNoError(log logger.ILogger, err error, msg ...any) {
	if err == nil {
		return
	}

	if len(msg) == 0 {
		log.Errorf("错误: %v", err)
	}

	if len(msg) == 1 {
		if format, ok := msg[0].(string); ok {
			log.Errorf("%s: %v", format, err)
		}

		log.Errorf("%v: %v", msg[0], err)
	}

	if format, ok := msg[0].(string); ok {
		log.Errorf(format+": %w", append(msg[1:], err)...)
	}

	log.Errorf("%s: %v", fmt.Sprint(msg...), err)
}

// ShouldValue 返回值和错误，如果错误不为 nil 则包装错误信息
// 用法: value, err := ShouldValue(someFunc())
func ShouldValue[T any](log logger.ILogger, value T, err error) T {
	shouldNoError(log, err)

	return value
}

// ShouldFunc 执行函数并返回其错误（可选包装）
func ShouldFunc(log logger.ILogger, f func() error, msg ...any) {
	err := f()

	shouldNoError(log, err, msg...)
}

// ShouldFuncValue 执行函数并返回值和错误（可选包装）
func ShouldFuncValue[T any](log logger.ILogger, f func() (T, error), msg ...any) T {
	value, err := f()

	shouldNoError(log, err, msg...)

	return value
}

// ShouldTrue 断言条件应该为真（Should 的别名，语义更清晰）
func ShouldTrue(log logger.ILogger, condition bool, msg ...any) {
	Should(log, condition, msg...)
}

// ShouldFalse 断言条件应该为假
func ShouldFalse(log logger.ILogger, condition bool, msg ...any) {
	Should(log, !condition, msg...)
}
