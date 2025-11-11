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
	}{
		{
			name:      "条件为真不应返回错误",
			condition: true,
			msg:       []any{"错误消息"},
			wantErr:   false,
		},
		{
			name:      "条件为假单个字符串消息",
			condition: false,
			msg:       []any{"自定义错误"},
			wantErr:   true,
		},
		{
			name:      "条件为假单个error消息",
			condition: false,
			msg:       []any{errors.New("错误对象")},
			wantErr:   true,
		},
		{
			name:      "条件为假格式化消息",
			condition: false,
			msg:       []any{"错误: %s, 代码: %d", "测试", 500},
			wantErr:   true,
		},
		{
			name:      "条件为假多个非字符串参数",
			condition: false,
			msg:       []any{123, 456},
			wantErr:   true,
		},
		{
			name:      "条件为假无消息",
			condition: false,
			msg:       []any{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assert.Should(tt.condition, tt.msg...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Should() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestShouldValue 测试 ShouldValue 函数
func TestShouldValue(t *testing.T) {
	t.Run("错误为nil应返回值", func(t *testing.T) {
		result, err := assert.ShouldValue(42, nil)
		if err != nil {
			t.Errorf("ShouldValue() error = %v, 期望 nil", err)
		}
		if result != 42 {
			t.Errorf("ShouldValue() = %v, 期望 42", result)
		}
	})

	t.Run("错误不为nil应返回错误", func(t *testing.T) {
		testErr := errors.New("测试错误")
		result, err := assert.ShouldValue(42, testErr)
		if err == nil {
			t.Errorf("ShouldValue() error = nil, 期望错误")
		}
		if result != 42 {
			t.Errorf("ShouldValue() = %v, 期望 42", result)
		}
	})

	t.Run("泛型支持字符串类型", func(t *testing.T) {
		result, err := assert.ShouldValue("hello", nil)
		if err != nil {
			t.Errorf("ShouldValue() error = %v, 期望 nil", err)
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
			t.Errorf("ShouldValue() error = %v, 期望 nil", err)
		}
		if result.Name != "张三" {
			t.Errorf("ShouldValue() = %v, 期望 张三", result.Name)
		}
	})
}

// TestShouldFunc 测试 ShouldFunc 函数
func TestShouldFunc(t *testing.T) {
	t.Run("函数返回nil", func(t *testing.T) {
		called := false
		err := assert.ShouldFunc(func() error {
			called = true
			return nil
		})
		if !called {
			t.Errorf("函数未被调用")
		}
		if err != nil {
			t.Errorf("ShouldFunc() error = %v, 期望 nil", err)
		}
	})

	t.Run("函数返回错误", func(t *testing.T) {
		err := assert.ShouldFunc(func() error {
			return errors.New("测试错误")
		})
		if err == nil {
			t.Errorf("ShouldFunc() error = nil, 期望错误")
		}
	})

	t.Run("函数返回错误带消息", func(t *testing.T) {
		err := assert.ShouldFunc(func() error {
			return errors.New("测试错误")
		}, "执行失败")
		if err == nil {
			t.Errorf("ShouldFunc() error = nil, 期望错误")
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
			t.Errorf("ShouldFuncValue() error = %v, 期望 nil", err)
		}
		if result != 42 {
			t.Errorf("ShouldFuncValue() = %v, 期望 42", result)
		}
	})

	t.Run("函数返回值和错误", func(t *testing.T) {
		result, err := assert.ShouldFuncValue(func() (string, error) {
			return "hello", errors.New("测试错误")
		})
		if err == nil {
			t.Errorf("ShouldFuncValue() error = nil, 期望错误")
		}
		if result != "hello" {
			t.Errorf("ShouldFuncValue() = %v, 期望 hello", result)
		}
	})

	t.Run("函数返回值和错误带消息", func(t *testing.T) {
		result, err := assert.ShouldFuncValue(func() (string, error) {
			return "hello", errors.New("测试错误")
		}, "操作失败")
		if err == nil {
			t.Errorf("ShouldFuncValue() error = nil, 期望错误")
		}
		if result != "hello" {
			t.Errorf("ShouldFuncValue() = %v, 期望 hello", result)
		}
	})
}

// TestShouldTrue 测试 ShouldTrue 函数
func TestShouldTrue(t *testing.T) {
	t.Run("条件为真", func(t *testing.T) {
		err := assert.ShouldTrue(true, "错误消息")
		if err != nil {
			t.Errorf("ShouldTrue() error = %v, 期望 nil", err)
		}
	})

	t.Run("条件为假", func(t *testing.T) {
		err := assert.ShouldTrue(false, "条件不满足")
		if err == nil {
			t.Errorf("ShouldTrue() error = nil, 期望错误")
		}
	})
}

// TestShouldFalse 测试 ShouldFalse 函数
func TestShouldFalse(t *testing.T) {
	t.Run("条件为假", func(t *testing.T) {
		err := assert.ShouldFalse(false, "错误消息")
		if err != nil {
			t.Errorf("ShouldFalse() error = %v, 期望 nil", err)
		}
	})

	t.Run("条件为真", func(t *testing.T) {
		err := assert.ShouldFalse(true, "条件应该为假")
		if err == nil {
			t.Errorf("ShouldFalse() error = nil, 期望错误")
		}
	})
}

// 示例：演示 Should 的使用
func ExampleShould() {
	// 条件为真，不返回错误
	err := assert.Should(1 > 0, "数字应该大于0")
	if err == nil {
		fmt.Println("条件为真")
	}

	// 条件为假，返回错误（错误信息会输出到 stderr）
	err = assert.Should(1 < 0, "数字应该大于0")
	if err != nil {
		fmt.Println("条件为假")
		fmt.Printf("错误: %v\n", err)
	}

	// Output:
	// 条件为真
	// 条件为假
	// 错误: 数字应该大于0
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
