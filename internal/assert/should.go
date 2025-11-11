package assert

import "fmt"

// Should 断言条件应该为真，否则静默失败
// 支持多种消息格式：
//   - Should(true, "错误")           // 单个消息
//   - Should(true, "错误: %v", err)  // 格式化消息
//   - Should(true, err)              // 直接传递 error
func Should(condition bool, msg ...any) {
	if condition {
		return
	}

	// 没有消息时使用默认消息
	if len(msg) == 0 {
		_ = fmt.Errorf("断言失败")
		return
	}

	// 单个参数处理
	if len(msg) == 1 {
		if err, ok := msg[0].(error); ok {
			_ = fmt.Errorf("%w", err)
			return
		}
		if format, ok := msg[0].(string); ok {
			_ = fmt.Errorf("%s", format)
			return
		}

		_ = fmt.Errorf("%v", msg[0])
		return
	}

	// 多个参数时尝试格式化
	if format, ok := msg[0].(string); ok {
		_ = fmt.Errorf(format, msg[1:]...)
		return
	}

	// 其他情况使用 Sprint
	_ = fmt.Errorf("%s", fmt.Sprint(msg...))
}

// shouldNoError 断言错误应该为 nil，否则静默失败
func shouldNoError(err error, msg ...any) {
	if err == nil {
		return
	}

	if len(msg) == 0 {
		_ = fmt.Errorf("%w", err)
		return
	}

	if len(msg) == 1 {
		if format, ok := msg[0].(string); ok {
			_ = fmt.Errorf("%s: %w", format, err)
			return
		}

		_ = fmt.Errorf("%v: %w", msg[0], err)
		return
	}

	if format, ok := msg[0].(string); ok {
		_ = fmt.Errorf(format+": %w", append(msg[1:], err)...)
		return
	}

	_ = fmt.Errorf("%v: %w", fmt.Sprint(msg...), err)
}

// ShouldValue 返回值和错误，如果错误不为 nil 则包装错误信息
// 用法: value, err := ShouldValue(someFunc())
func ShouldValue[T any](value T, err error) T {
	shouldNoError(err)

	return value
}

// ShouldFunc 执行函数并返回其错误（可选包装）
func ShouldFunc(f func() error, msg ...any) {
	err := f()

	shouldNoError(err, msg...)
}

// ShouldFuncValue 执行函数并返回值和错误（可选包装）
func ShouldFuncValue[T any](f func() (T, error), msg ...any) T {
	value, err := f()

	shouldNoError(err, msg...)

	return value
}

// ShouldTrue 断言条件应该为真（Should 的别名，语义更清晰）
func ShouldTrue(condition bool, msg ...any) {
	Should(condition, msg...)
}

// ShouldFalse 断言条件应该为假
func ShouldFalse(condition bool, msg ...any) {
	Should(!condition, msg...)
}
