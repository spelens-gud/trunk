package assert

import (
	"errors"
	"fmt"
	"os"
)

// Should 断言条件应该为真，否则返回错误
// 支持多种消息格式：
//   - Should(true, "错误")           // 单个消息
//   - Should(true, "错误: %v", err)  // 格式化消息
//   - Should(true, err)              // 直接传递 error
func Should(condition bool, msg ...any) error {
	if condition {
		return nil
	}

	var err error
	// 没有消息时使用默认消息
	if len(msg) == 0 {
		err = errors.New("断言失败")
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	// 单个参数处理
	if len(msg) == 1 {
		if e, ok := msg[0].(error); ok {
			fmt.Fprintf(os.Stderr, "%s\n", e.Error())
			return e
		}
		if format, ok := msg[0].(string); ok {
			err = errors.New(format)
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return err
		}

		err = fmt.Errorf("%v", msg[0])
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	// 多个参数时尝试格式化
	if format, ok := msg[0].(string); ok {
		err = fmt.Errorf(format, msg[1:]...)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	// 其他情况使用 Sprint
	err = errors.New(fmt.Sprint(msg...))
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	return err
}

// shouldNoError 断言错误应该为 nil，否则返回错误
func shouldNoError(err error, msg ...any) error {
	if err == nil {
		return nil
	}

	var resultErr error
	if len(msg) == 0 {
		resultErr = fmt.Errorf("错误: %w", err)
		fmt.Fprintf(os.Stderr, "%s\n", resultErr.Error())
		return resultErr
	}

	if len(msg) == 1 {
		if format, ok := msg[0].(string); ok {
			resultErr = fmt.Errorf("%s: %w", format, err)
			fmt.Fprintf(os.Stderr, "%s\n", resultErr.Error())
			return resultErr
		}

		resultErr = fmt.Errorf("%v: %w", msg[0], err)
		fmt.Fprintf(os.Stderr, "%s\n", resultErr.Error())
		return resultErr
	}

	if format, ok := msg[0].(string); ok {
		resultErr = fmt.Errorf(format+": %w", append(msg[1:], err)...)
		fmt.Fprintf(os.Stderr, "%s\n", resultErr.Error())
		return resultErr
	}

	resultErr = fmt.Errorf("%s: %w", fmt.Sprint(msg...), err)
	fmt.Fprintf(os.Stderr, "%s\n", resultErr.Error())
	return resultErr
}

// ShouldValue 返回值和错误，如果错误不为 nil 则包装错误信息
// 用法: value, err := ShouldValue(someFunc())
func ShouldValue[T any](value T, err error) (T, error) {
	return value, shouldNoError(err)
}

// ShouldFunc 执行函数并返回其错误（可选包装）
func ShouldFunc(f func() error, msg ...any) error {
	err := f()
	return shouldNoError(err, msg...)
}

// ShouldFuncValue 执行函数并返回值和错误（可选包装）
func ShouldFuncValue[T any](f func() (T, error), msg ...any) (T, error) {
	value, err := f()
	return value, shouldNoError(err, msg...)
}

// ShouldTrue 断言条件应该为真（Should 的别名，语义更清晰）
func ShouldTrue(condition bool, msg ...any) error {
	return Should(condition, msg...)
}

// ShouldFalse 断言条件应该为假
func ShouldFalse(condition bool, msg ...any) error {
	return Should(!condition, msg...)
}
