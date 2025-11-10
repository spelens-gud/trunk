package assert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spelens-gud/trunk/internal/assert"
)

// TestShould 测试 Should 函数
func TestShould(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		msg       []any
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "条件为真不应返回错误",
			condition: true,
			msg:       []any{"错误消息"},
			wantErr:   false,
		},
		{
			name:      "条件为假无消息应返回默认错误",
			condition: false,
			msg:       []any{},
			wantErr:   true,
			errMsg:    "断言失败",
		},
		{
			name:      "条件为假单个字符串消息",
			condition: false,
			msg:       []any{"自定义错误"},
			wantErr:   true,
			errMsg:    "自定义错误",
		},
		{
			name:      "条件为假单个error消息",
			condition: false,
			msg:       []any{errors.New("错误对象")},
			wantErr:   true,
			errMsg:    "错误对象",
		},
		{
			name:      "条件为假格式化消息",
			condition: false,
			msg:       []any{"错误: %s, 代码: %d", "测试", 500},
			wantErr:   true,
			errMsg:    "错误: 测试, 代码: 500",
		},
		{
			name:      "条件为假多个非字符串参数",
			condition: false,
			msg:       []any{123, 456},
			wantErr:   true,
			errMsg:    "123 456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assert.Should(tt.condition, tt.msg...)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望返回错误但没有")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("错误消息 = %v, 期望 %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误但返回了: %v", err)
				}
			}
		})
	}
}

// TestShouldValue 测试 ShouldValue 函数
func TestShouldValue(t *testing.T) {
	t.Run("错误为nil应返回值和nil错误", func(t *testing.T) {
		result, err := assert.ShouldValue(42, nil)
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
		if result != 42 {
			t.Errorf("ShouldValue() = %v, 期望 42", result)
		}
	})

	t.Run("错误不为nil应返回值和错误", func(t *testing.T) {
		testErr := errors.New("测试错误")
		result, err := assert.ShouldValue(42, testErr)
		if err == nil {
			t.Errorf("期望返回错误但没有")
		}
		if result != 42 {
			t.Errorf("ShouldValue() = %v, 期望 42", result)
		}
	})

	t.Run("错误不为nil带消息应返回包装错误", func(t *testing.T) {
		testErr := errors.New("测试错误")
		result, err := assert.ShouldValue(42, testErr, "操作失败")
		if err == nil {
			t.Errorf("期望返回错误但没有")
		}
		if err.Error() != "操作失败: 测试错误" {
			t.Errorf("错误消息 = %v, 期望 操作失败: 测试错误", err.Error())
		}
		if result != 42 {
			t.Errorf("ShouldValue() = %v, 期望 42", result)
		}
	})

	t.Run("泛型支持字符串类型", func(t *testing.T) {
		result, err := assert.ShouldValue("hello", nil)
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
		if result != "hello" {
			t.Errorf("ShouldValue() = %v, 期望 hello", result)
		}
	})

	t.Run("泛型支持结构体类型", func(t *testing.T) {
		type User struct {
			Name string
		}
		user := User{Name: "张三"}
		result, err := assert.ShouldValue(user, nil)
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
		if result.Name != "张三" {
			t.Errorf("ShouldValue() = %v, 期望 张三", result.Name)
		}
	})
}

// TestShouldFunc 测试 ShouldFunc 函数
func TestShouldFunc(t *testing.T) {
	t.Run("函数返回nil不应返回错误", func(t *testing.T) {
		called := false
		err := assert.ShouldFunc(func() error {
			called = true
			return nil
		})
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
		if !called {
			t.Errorf("函数未被调用")
		}
	})

	t.Run("函数返回错误无消息应返回错误", func(t *testing.T) {
		err := assert.ShouldFunc(func() error {
			return errors.New("测试错误")
		})
		if err == nil {
			t.Errorf("期望返回错误但没有")
			return
		}
		if err.Error() != "测试错误" {
			t.Errorf("错误消息 = %v, 期望 测试错误", err.Error())
		}
	})

	t.Run("函数返回错误带消息应返回包装错误", func(t *testing.T) {
		err := assert.ShouldFunc(func() error {
			return errors.New("测试错误")
		}, "执行失败")
		if err == nil {
			t.Errorf("期望返回错误但没有")
			return
		}
		if err.Error() != "执行失败: 测试错误" {
			t.Errorf("错误消息 = %v, 期望 执行失败: 测试错误", err.Error())
		}
	})
}

// TestShouldFuncValue 测试 ShouldFuncValue 函数
func TestShouldFuncValue(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result, err := assert.ShouldFuncValue(func() (int, error) {
			return 42, nil
		})
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
		if result != 42 {
			t.Errorf("ShouldFuncValue() = %v, 期望 42", result)
		}
	})

	t.Run("函数返回值和错误", func(t *testing.T) {
		result, err := assert.ShouldFuncValue(func() (int, error) {
			return 42, errors.New("测试错误")
		})
		if err == nil {
			t.Errorf("期望返回错误但没有")
		}
		if result != 42 {
			t.Errorf("ShouldFuncValue() = %v, 期望 42", result)
		}
	})

	t.Run("函数返回值和错误带消息", func(t *testing.T) {
		result, err := assert.ShouldFuncValue(func() (string, error) {
			return "hello", errors.New("测试错误")
		}, "操作失败")
		if err == nil {
			t.Errorf("期望返回错误但没有")
		}
		if err.Error() != "操作失败: 测试错误" {
			t.Errorf("错误消息 = %v, 期望 操作失败: 测试错误", err.Error())
		}
		if result != "hello" {
			t.Errorf("ShouldFuncValue() = %v, 期望 hello", result)
		}
	})
}

// TestShouldTrue 测试 ShouldTrue 函数
func TestShouldTrue(t *testing.T) {
	t.Run("条件为真不应返回错误", func(t *testing.T) {
		err := assert.ShouldTrue(true, "错误消息")
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
	})

	t.Run("条件为假应返回错误", func(t *testing.T) {
		err := assert.ShouldTrue(false, "条件不满足")
		if err == nil {
			t.Errorf("期望返回错误但没有")
		}
	})
}

// TestShouldFalse 测试 ShouldFalse 函数
func TestShouldFalse(t *testing.T) {
	t.Run("条件为假不应返回错误", func(t *testing.T) {
		err := assert.ShouldFalse(false, "错误消息")
		if err != nil {
			t.Errorf("不期望返回错误: %v", err)
		}
	})

	t.Run("条件为真应返回错误", func(t *testing.T) {
		err := assert.ShouldFalse(true, "条件应该为假")
		if err == nil {
			t.Errorf("期望返回错误但没有")
		}
	})
}

// 示例：演示 Should 的使用
func ExampleShould() {
	// 条件为真，不返回错误
	err := assert.Should(1 > 0, "数字应该大于0")
	fmt.Println(err == nil)

	// 条件为假，返回错误
	err = assert.Should(1 < 0, "数字应该大于0")
	fmt.Println(err != nil)

	// Output:
	// true
	// true
}

// BenchmarkShould 性能测试
func BenchmarkShould(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = assert.Should(true, "错误消息")
	}
}

// BenchmarkShouldValue 性能测试
func BenchmarkShouldValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = assert.ShouldValue(42, nil)
	}
}
