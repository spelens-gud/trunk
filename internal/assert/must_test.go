package assert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spelens-gud/trunk/internal/assert"
)

// TestMust 测试 Must 函数
func TestMust(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		msg       []any
		wantPanic bool
		panicMsg  string
	}{
		{
			name:      "条件为真不应panic",
			condition: true,
			msg:       []any{"错误消息"},
			wantPanic: false,
		},
		{
			name:      "条件为假无消息应panic默认消息",
			condition: false,
			msg:       []any{},
			wantPanic: true,
			panicMsg:  "断言失败",
		},
		{
			name:      "条件为假单个字符串消息",
			condition: false,
			msg:       []any{"自定义错误"},
			wantPanic: true,
			panicMsg:  "自定义错误",
		},
		{
			name:      "条件为假单个error消息",
			condition: false,
			msg:       []any{errors.New("错误对象")},
			wantPanic: true,
			panicMsg:  "错误对象",
		},
		{
			name:      "条件为假格式化消息",
			condition: false,
			msg:       []any{"错误: %s, 代码: %d", "测试", 500},
			wantPanic: true,
			panicMsg:  "错误: 测试, 代码: 500",
		},
		{
			name:      "条件为假多个非字符串参数",
			condition: false,
			msg:       []any{123, 456},
			wantPanic: true,
			panicMsg:  "123 456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic {
					if r == nil {
						t.Errorf("期望 panic 但没有发生")

						return
					}
					panicMsg := fmt.Sprint(r)
					if panicMsg != tt.panicMsg {
						t.Errorf("panic 消息 = %v, 期望 %v", panicMsg, tt.panicMsg)
					}
				} else {
					if r != nil {
						t.Errorf("不期望 panic 但发生了: %v", r)
					}
				}
			}()

			assert.Must(tt.condition, tt.msg...)
		})
	}
}

// TestMustValue 测试 MustValue 函数
func TestMustValue(t *testing.T) {
	t.Run("错误为nil应返回值", func(t *testing.T) {
		result := assert.MustValue(42, nil)
		if result != 42 {
			t.Errorf("MustValue() = %v, 期望 42", result)
		}
	})

	t.Run("错误不为nil应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustValue(42, errors.New("测试错误"))
	})

	t.Run("泛型支持字符串类型", func(t *testing.T) {
		result := assert.MustValue("hello", nil)
		if result != "hello" {
			t.Errorf("MustValue() = %v, 期望 hello", result)
		}
	})

	t.Run("泛型支持结构体类型", func(t *testing.T) {
		type User struct {
			Name string
		}
		user := User{Name: "张三"}
		result := assert.MustValue(user, nil)
		if result.Name != "张三" {
			t.Errorf("MustValue() = %v, 期望 张三", result.Name)
		}
	})
}

// TestMustFunc 测试 MustFunc 函数
func TestMustFunc(t *testing.T) {
	t.Run("函数返回nil不应panic", func(t *testing.T) {
		called := false
		assert.MustFunc(func() error {
			called = true
			return nil
		})
		if !called {
			t.Errorf("函数未被调用")
		}
	})

	t.Run("函数返回错误无消息应panic", func(t *testing.T) {
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

		assert.MustFunc(func() error {
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
			if fmt.Sprint(r) != "执行失败: 测试错误" {
				t.Errorf("panic 消息 = %v, 期望 执行失败: 测试错误", r)
			}
		}()

		assert.MustFunc(func() error {
			return errors.New("测试错误")
		}, "执行失败")
	})
}

// TestMustTrue 测试 MustTrue 函数
func TestMustTrue(t *testing.T) {
	t.Run("条件为真不应panic", func(t *testing.T) {
		assert.MustTrue(true, "错误消息")
	})

	t.Run("条件为假应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustTrue(false, "条件不满足")
	})
}

// TestMustFalse 测试 MustFalse 函数
func TestMustFalse(t *testing.T) {
	t.Run("条件为假不应panic", func(t *testing.T) {
		assert.MustFalse(false, "错误消息")
	})

	t.Run("条件为真应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("期望 panic 但没有发生")
			}
		}()

		assert.MustFalse(true, "条件应该为假")
	})
}

// BenchmarkMust 性能测试
func BenchmarkMust(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.Must(true, "错误消息")
	}
}

// BenchmarkMustValue 性能测试
func BenchmarkMustValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = assert.MustValue(42, nil)
	}
}
