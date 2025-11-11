package assert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
	"go.uber.org/zap"
)

// mockLogger 是一个用于测试的 mock logger
type mockLogger struct {
	logs []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{logs: make([]string, 0)}
}

func (m *mockLogger) Debug(msg string, fields ...zap.Field) {
	m.logs = append(m.logs, msg)
}
func (m *mockLogger) Info(msg string, fields ...zap.Field) {
	m.logs = append(m.logs, msg)
}
func (m *mockLogger) Warn(msg string, fields ...zap.Field) {
	m.logs = append(m.logs, msg)
}
func (m *mockLogger) Error(msg string, fields ...zap.Field) {
	m.logs = append(m.logs, msg)
}
func (m *mockLogger) Fatal(msg string, fields ...zap.Field) {
	m.logs = append(m.logs, msg)
}
func (m *mockLogger) Panic(msg string, fields ...zap.Field) {
	m.logs = append(m.logs, msg)
}
func (m *mockLogger) Debugf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}
func (m *mockLogger) Infof(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}
func (m *mockLogger) Warnf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}
func (m *mockLogger) Errorf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}
func (m *mockLogger) Fatalf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}
func (m *mockLogger) Panicf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}
func (m *mockLogger) With(fields ...zap.Field) logger.ILogger {
	return m
}
func (m *mockLogger) WithPrefix(prefix string) logger.ILogger {
	return m
}
func (m *mockLogger) Sync() error {
	return nil
}

// TestShould 测试 Should 函数
func TestShould(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		msg       []any
	}{
		{
			name:      "条件为真不应panic",
			condition: true,
			msg:       []any{"错误消息"},
		},
		{
			name:      "条件为假单个字符串消息",
			condition: false,
			msg:       []any{"自定义错误"},
		},
		{
			name:      "条件为假单个error消息",
			condition: false,
			msg:       []any{errors.New("错误对象")},
		},
		{
			name:      "条件为假格式化消息",
			condition: false,
			msg:       []any{"错误: %s, 代码: %d", "测试", 500},
		},
		{
			name:      "条件为假多个非字符串参数",
			condition: false,
			msg:       []any{123, 456},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := newMockLogger()
			assert.Should(log, tt.condition, tt.msg...)
		})
	}
}

// TestShouldValue 测试 ShouldValue 函数
func TestShouldValue(t *testing.T) {
	t.Run("错误为nil应返回值", func(t *testing.T) {
		log := newMockLogger()
		result := assert.ShouldValue(log, 42, nil)
		if result != 42 {
			t.Errorf("ShouldValue() = %v, 期望 42", result)
		}
	})

	t.Run("泛型支持字符串类型", func(t *testing.T) {
		log := newMockLogger()
		result := assert.ShouldValue(log, "hello", nil)
		if result != "hello" {
			t.Errorf("ShouldValue() = %v, 期望 hello", result)
		}
	})

	t.Run("泛型支持结构体类型", func(t *testing.T) {
		log := newMockLogger()
		type User struct {
			Name string
		}
		user := User{Name: "张三"}
		result := assert.ShouldValue(log, user, nil)
		if result.Name != "张三" {
			t.Errorf("ShouldValue() = %v, 期望 张三", result.Name)
		}
	})
}

// TestShouldFunc 测试 ShouldFunc 函数
func TestShouldFunc(t *testing.T) {
	t.Run("函数返回nil", func(t *testing.T) {
		log := newMockLogger()
		called := false
		assert.ShouldFunc(log, func() error {
			called = true
			return nil
		})
		if !called {
			t.Errorf("函数未被调用")
		}
	})

	t.Run("函数返回错误带消息", func(t *testing.T) {
		log := newMockLogger()
		assert.ShouldFunc(log, func() error {
			return errors.New("测试错误")
		}, "执行失败")
	})
}

// TestShouldFuncValue 测试 ShouldFuncValue 函数
func TestShouldFuncValue(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		log := newMockLogger()
		result := assert.ShouldFuncValue(log, func() (int, error) {
			return 42, nil
		})
		if result != 42 {
			t.Errorf("ShouldFuncValue() = %v, 期望 42", result)
		}
	})

	t.Run("函数返回值和错误带消息", func(t *testing.T) {
		log := newMockLogger()
		result := assert.ShouldFuncValue(log, func() (string, error) {
			return "hello", errors.New("测试错误")
		}, "操作失败")
		if result != "hello" {
			t.Errorf("ShouldFuncValue() = %v, 期望 hello", result)
		}
	})
}

// TestShouldTrue 测试 ShouldTrue 函数
func TestShouldTrue(t *testing.T) {
	t.Run("条件为真", func(t *testing.T) {
		log := newMockLogger()
		assert.ShouldTrue(log, true, "错误消息")
	})

	t.Run("条件为假", func(t *testing.T) {
		log := newMockLogger()
		assert.ShouldTrue(log, false, "条件不满足")
	})
}

// TestShouldFalse 测试 ShouldFalse 函数
func TestShouldFalse(t *testing.T) {
	t.Run("条件为假", func(t *testing.T) {
		log := newMockLogger()
		assert.ShouldFalse(log, false, "错误消息")
	})

	t.Run("条件为真", func(t *testing.T) {
		log := newMockLogger()
		assert.ShouldFalse(log, true, "条件应该为假")
	})
}

// 示例：演示 Should 的使用
func ExampleShould() {
	log := newMockLogger()

	// 条件为真，不执行任何操作
	assert.Should(log, 1 > 0, "数字应该大于0")
	fmt.Println("条件为真")

	// 条件为假，执行错误处理
	assert.Should(log, 1 < 0, "数字应该大于0")
	fmt.Println("条件为假")

	// Output:
	// 条件为真
	// 条件为假
}

// BenchmarkShould 性能测试
func BenchmarkShould(b *testing.B) {
	log := newMockLogger()
	for i := 0; i < b.N; i++ {
		assert.Should(log, true, "错误消息")
	}
}

// BenchmarkShouldValue 性能测试
func BenchmarkShouldValue(b *testing.B) {
	log := newMockLogger()
	for i := 0; i < b.N; i++ {
		_ = assert.ShouldValue(log, 42, nil)
	}
}
