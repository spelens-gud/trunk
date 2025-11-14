package assert_test

import (
	"errors"
	"fmt"
	"testing"

	`github.com/spelens-gud/assert`
	"go.uber.org/zap"
)

// TestMustNoError 测试 mustNoError 的行为（通过 MustCall 函数间接测试）
func TestMustNoError(t *testing.T) {
	t.Run("错误为nil不应panic", func(t *testing.T) {
		assert.MustCall0E(func() error {
			return nil
		})
	})

	t.Run("错误不为nil应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall0E(func() error {
			return errors.New("测试错误")
		})
	})

	t.Run("带自定义消息的错误应panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall0E(func() error {
			return errors.New("测试错误")
		}, "操作失败")
	})
}

// ============ 泛型函数调用测试 ============

// TestMustCall0E 测试无参数返回error的函数
func TestMustCall0E(t *testing.T) {
	t.Run("函数返回nil不应panic", func(t *testing.T) {
		called := false
		assert.MustCall0E(func() error {
			called = true
			return nil
		})
		if !called {
			t.Errorf("函数未被调用")
		}
	})

	t.Run("函数返回错误应panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("期望 panic 但没有发生")
				return
			}
			if fmt.Sprint(r) != "测试错误" {
				t.Errorf("panic 消息 = %v, 期望 测试错误", r)
			}
		}()

		assert.MustCall0E(func() error {
			return errors.New("测试错误")
		})
	})

	t.Run("函数返回错误带消息应panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("期望 panic 但没有发生")
				return
			}
			if fmt.Sprint(r) != "操作失败: 测试错误" {
				t.Errorf("panic 消息 = %v, 期望 操作失败: 测试错误", r)
			}
		}()

		assert.MustCall0E(func() error {
			return errors.New("测试错误")
		}, "操作失败")
	})
}

// TestMustCall0RE 测试无参数返回值和error的函数
func TestMustCall0RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.MustCall0RE(func() (int, error) {
			return 42, nil
		})
		if result != 42 {
			t.Errorf("MustCall0RE() = %v, 期望 42", result)
		}
	})

	t.Run("函数返回值和错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall0RE(func() (string, error) {
			return "hello", errors.New("测试错误")
		}, "获取失败")
	})
}

// TestMustCall1E 测试单参数返回error的函数
func TestMustCall1E(t *testing.T) {
	t.Run("函数返回nil不应panic", func(t *testing.T) {
		var receivedArg int
		assert.MustCall1E(func(x int) error {
			receivedArg = x
			return nil
		}, 42)
		if receivedArg != 42 {
			t.Errorf("参数传递错误，收到 %v, 期望 42", receivedArg)
		}
	})

	t.Run("函数返回错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall1E(func(x int) error {
			return fmt.Errorf("处理 %d 失败", x)
		}, 42, "处理失败")
	})
}

// TestMustCall1RE 测试单参数返回值和error的函数
func TestMustCall1RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.MustCall1RE(func(x int) (string, error) {
			return fmt.Sprintf("值: %d", x), nil
		}, 42)
		if result != "值: 42" {
			t.Errorf("MustCall1RE() = %v, 期望 '值: 42'", result)
		}
	})

	t.Run("函数返回值和错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall1RE(func(x int) (string, error) {
			return fmt.Sprintf("值: %d", x), errors.New("转换失败")
		}, 42, "转换失败")
	})
}

// TestMustCall2E 测试双参数返回error的函数
func TestMustCall2E(t *testing.T) {
	t.Run("函数返回nil不应panic", func(t *testing.T) {
		var sum int
		assert.MustCall2E(func(x, y int) error {
			sum = x + y
			return nil
		}, 10, 20)
		if sum != 30 {
			t.Errorf("计算错误，收到 %v, 期望 30", sum)
		}
	})

	t.Run("函数返回错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall2E(func(x, y int) error {
			return fmt.Errorf("处理 %d + %d 失败", x, y)
		}, 10, 20, "计算失败")
	})
}

// TestMustCall2RE 测试双参数返回值和error的函数
func TestMustCall2RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.MustCall2RE(func(x, y int) (int, error) {
			return x + y, nil
		}, 10, 20)
		if result != 30 {
			t.Errorf("MustCall2RE() = %v, 期望 30", result)
		}
	})

	t.Run("函数返回值和错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall2RE(func(x, y int) (int, error) {
			return x + y, errors.New("计算失败")
		}, 10, 20, "加法失败")
	})
}

// TestMustCall3E 测试三参数返回error的函数
func TestMustCall3E(t *testing.T) {
	t.Run("函数返回nil不应panic", func(t *testing.T) {
		var result int
		assert.MustCall3E(func(x, y, z int) error {
			result = x + y + z
			return nil
		}, 10, 20, 30)
		if result != 60 {
			t.Errorf("计算错误，收到 %v, 期望 60", result)
		}
	})

	t.Run("函数返回错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall3E(func(x, y, z int) error {
			return fmt.Errorf("处理 %d + %d + %d 失败", x, y, z)
		}, 10, 20, 30, "计算失败")
	})
}

// TestMustCall3RE 测试三参数返回值和error的函数
func TestMustCall3RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.MustCall3RE(func(x, y, z int) (int, error) {
			return x + y + z, nil
		}, 10, 20, 30)
		if result != 60 {
			t.Errorf("MustCall3RE() = %v, 期望 60", result)
		}
	})

	t.Run("函数返回值和错误应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustCall3RE(func(x, y, z int) (int, error) {
			return x + y + z, errors.New("计算失败")
		}, 10, 20, 30, "计算失败")
	})
}

// mockPanicLogger 测试用的模拟panic日志记录器
type mockLogger struct {
	lastMsg   string
	lastError string
	callCount int
}

func (l *mockLogger) Panic(msg string, fields ...zap.Field) {
	l.callCount++
	l.lastMsg = msg
}

func (l *mockLogger) Panicf(template string, args ...any) {
	l.callCount++
	l.lastMsg = fmt.Sprintf(template, args...)
}
func (l *mockLogger) Error(msg string, fields ...zap.Field) {
	l.callCount++
	l.lastMsg = msg
}

func (l *mockLogger) Errorf(template string, args ...any) {
	l.callCount++
	l.lastMsg = fmt.Sprintf(template, args...)
	if len(args) > 0 {
		l.lastError = fmt.Sprint(args[len(args)-1])
	}
}

// TestSetPanicLogger 测试日志记录器设置
func TestSetPanicLogger(t *testing.T) {
	// 创建一个模拟的日志记录器
	mock := &mockLogger{}

	// 设置日志记录器
	assert.SetLogger(mock)

	// 测试panic时的日志记录
	defer func() {
		if r := recover(); r != nil {
			// 验证日志记录器被调用
			if mock.callCount > 0 {
				t.Logf("日志记录器被调用 %d 次，消息: %s", mock.callCount, mock.lastMsg)
			}
		}
		// 清理：重置日志记录器
		assert.SetLogger(nil)
	}()

	assert.MustCall0E(func() error {
		return errors.New("测试错误")
	}, "操作失败")
}

// 示例：演示泛型函数调用的使用
func ExampleMustCall0RE() {
	// 无参数返回值和error
	value := assert.MustCall0RE(func() (int, error) {
		return 42, nil
	}, "获取值失败")

	fmt.Printf("获取到值: %d\n", value)

	// Output:
	// 获取到值: 42
}

func ExampleMustCall1RE() {
	// 单参数返回值和error
	result := assert.MustCall1RE(func(x int) (string, error) {
		return fmt.Sprintf("数字是: %d", x), nil
	}, 42, "转换失败")

	fmt.Println(result)

	// Output:
	// 数字是: 42
}

func ExampleMustCall0E() {
	// 无参数返回error
	assert.MustCall0E(func() error {
		fmt.Println("执行成功")
		return nil
	}, "操作失败")

	// Output:
	// 执行成功
}

// BenchmarkMustCall0E 性能测试
func BenchmarkMustCall0E(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.MustCall0E(func() error {
			return nil
		})
	}
}

// BenchmarkMustCall0RE 性能测试
func BenchmarkMustCall0RE(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = assert.MustCall0RE(func() (int, error) {
			return 42, nil
		})
	}
}
