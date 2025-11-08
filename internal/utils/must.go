package utils

import "fmt"

// Must 断言条件必须为真，否则触发 panic
// 支持多种消息格式：
//   - Must(true, "错误")           // 单个消息
//   - Must(true, "错误: %v", err)  // 格式化消息
//   - Must(true, err)              // 直接传递 error
func Must(condition bool, msg ...any) {
	if condition {
		return
	}

	// 没有消息时使用默认消息
	if len(msg) == 0 {
		panic("断言失败")
	}

	// 单个参数直接 panic
	if len(msg) == 1 {
		panic(msg[0])
	}

	// 多个参数时尝试格式化
	if format, ok := msg[0].(string); ok {
		panic(fmt.Sprintf(format, msg[1:]...))
	}

	// 其他情况使用 Sprint
	panic(fmt.Sprint(msg...))
}

// MustNoError 断言错误必须为 nil，否则触发 panic
func MustNoError(err error, msg ...any) {
	if err == nil {
		return
	}

	if len(msg) == 0 {
		panic(err)
	}

	if len(msg) == 1 {
		if format, ok := msg[0].(string); ok {
			panic(fmt.Sprintf("%s: %v", format, err))
		}
		panic(fmt.Sprintf("%v: %v", msg[0], err))
	}

	if format, ok := msg[0].(string); ok {
		panic(fmt.Sprintf(format+": %v", append(msg[1:], err)...))
	}
	panic(fmt.Sprintf("%v: %v", fmt.Sprint(msg...), err))
}

// MustValue 返回值并断言错误必须为 nil
// 用法: value := MustValue(someFunc())
func MustValue[T any](value T, err error) T {
	MustNoError(err)
	return value
}

// MustFunc 执行函数并断言其返回的错误必须为 nil
func MustFunc(f func() error, msg ...any) {
	err := f()
	MustNoError(err, msg...)
}

func MustFuncValue[T any](f func() (T, error), msg ...any) T {
	value, err := f()
	MustNoError(err, msg)
	return value
}

// MustTrue 断言条件必须为真（Must 的别名，语义更清晰）
func MustTrue(condition bool, msg ...any) {
	Must(condition, msg...)
}

// MustFalse 断言条件必须为假
func MustFalse(condition bool, msg ...any) {
	Must(!condition, msg...)
}
