package assert_test

import (
	"errors"
	"fmt"
	"testing"

	`github.com/spelens-gud/assert`
)

// ============ 泛型函数调用测试 ============

// TestShouldCall0E 测试无参数返回error的函数
func TestShouldCall0E(t *testing.T) {
	t.Run("函数返回nil", func(t *testing.T) {
		called := false
		assert.ShouldCall0E(func() error {
			called = true
			return nil
		})
		if !called {
			t.Errorf("函数未被调用")
		}
	})

	t.Run("函数返回错误", func(t *testing.T) {
		// ShouldCall0E 不返回错误，只记录日志
		assert.ShouldCall0E(func() error {
			return errors.New("测试错误")
		}, "操作失败")
		// 函数应该被调用，但不会panic
	})
}

// TestShouldCall0RE 测试无参数返回值和error的函数
func TestShouldCall0RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.ShouldCall0RE(func() (int, error) {
			return 42, nil
		})
		if result != 42 {
			t.Errorf("ShouldCall0RE() = %v, 期望 42", result)
		}
	})

	t.Run("函数返回值和错误", func(t *testing.T) {
		result := assert.ShouldCall0RE(func() (string, error) {
			return "hello", errors.New("测试错误")
		}, "获取失败")
		// 即使有错误，也会返回值（错误只记录日志）
		if result != "hello" {
			t.Errorf("ShouldCall0RE() = %v, 期望 hello", result)
		}
	})
}

// TestShouldCall1E 测试单参数返回error的函数
func TestShouldCall1E(t *testing.T) {
	t.Run("函数返回nil", func(t *testing.T) {
		var receivedArg int
		assert.ShouldCall1E(func(x int) error {
			receivedArg = x
			return nil
		}, 42)
		if receivedArg != 42 {
			t.Errorf("参数传递错误，收到 %v, 期望 42", receivedArg)
		}
	})

	t.Run("函数返回错误", func(t *testing.T) {
		assert.ShouldCall1E(func(x int) error {
			return fmt.Errorf("处理 %d 失败", x)
		}, 42, "处理失败")
		// 函数应该被调用，错误只记录日志
	})
}

// TestShouldCall1RE 测试单参数返回值和error的函数
func TestShouldCall1RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.ShouldCall1RE(func(x int) (string, error) {
			return fmt.Sprintf("值: %d", x), nil
		}, 42)
		if result != "值: 42" {
			t.Errorf("ShouldCall1RE() = %v, 期望 '值: 42'", result)
		}
	})

	t.Run("函数返回值和错误", func(t *testing.T) {
		result := assert.ShouldCall1RE(func(x int) (string, error) {
			return fmt.Sprintf("值: %d", x), errors.New("转换失败")
		}, 42, "转换失败")
		// 即使有错误，也会返回值（错误只记录日志）
		if result != "值: 42" {
			t.Errorf("ShouldCall1RE() = %v, 期望 '值: 42'", result)
		}
	})
}

// TestShouldCall2E 测试双参数返回error的函数
func TestShouldCall2E(t *testing.T) {
	t.Run("函数返回nil", func(t *testing.T) {
		var sum int
		assert.ShouldCall2E(func(x, y int) error {
			sum = x + y
			return nil
		}, 10, 20)
		if sum != 30 {
			t.Errorf("计算错误，收到 %v, 期望 30", sum)
		}
	})

	t.Run("函数返回错误", func(t *testing.T) {
		assert.ShouldCall2E(func(x, y int) error {
			return fmt.Errorf("处理 %d + %d 失败", x, y)
		}, 10, 20, "计算失败")
		// 函数应该被调用，错误只记录日志
	})
}

// TestShouldCall2RE 测试双参数返回值和error的函数
func TestShouldCall2RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.ShouldCall2RE(func(x, y int) (int, error) {
			return x + y, nil
		}, 10, 20)
		if result != 30 {
			t.Errorf("ShouldCall2RE() = %v, 期望 30", result)
		}
	})

	t.Run("函数返回值和错误", func(t *testing.T) {
		result := assert.ShouldCall2RE(func(x, y int) (int, error) {
			return x + y, errors.New("计算失败")
		}, 10, 20, "加法失败")
		// 即使有错误，也会返回值（错误只记录日志）
		if result != 30 {
			t.Errorf("ShouldCall2RE() = %v, 期望 30", result)
		}
	})
}

// TestShouldCall3E 测试三参数返回error的函数
func TestShouldCall3E(t *testing.T) {
	t.Run("函数返回nil", func(t *testing.T) {
		var result int
		assert.ShouldCall3E(func(x, y, z int) error {
			result = x + y + z
			return nil
		}, 10, 20, 30)
		if result != 60 {
			t.Errorf("计算错误，收到 %v, 期望 60", result)
		}
	})

	t.Run("函数返回错误", func(t *testing.T) {
		assert.ShouldCall3E(func(x, y, z int) error {
			return fmt.Errorf("处理 %d + %d + %d 失败", x, y, z)
		}, 10, 20, 30, "计算失败")
		// 函数应该被调用，错误只记录日志
	})
}

// TestShouldCall3RE 测试三参数返回值和error的函数
func TestShouldCall3RE(t *testing.T) {
	t.Run("函数返回值和nil错误", func(t *testing.T) {
		result := assert.ShouldCall3RE(func(x, y, z int) (int, error) {
			return x + y + z, nil
		}, 10, 20, 30)
		if result != 60 {
			t.Errorf("ShouldCall3RE() = %v, 期望 60", result)
		}
	})

	t.Run("函数返回值和错误", func(t *testing.T) {
		result := assert.ShouldCall3RE(func(x, y, z int) (int, error) {
			return x + y + z, errors.New("计算失败")
		}, 10, 20, 30, "计算失败")
		// 即使有错误，也会返回值（错误只记录日志）
		if result != 60 {
			t.Errorf("ShouldCall3RE() = %v, 期望 60", result)
		}
	})
}

// 示例：演示泛型函数调用的使用
func ExampleShouldCall0RE() {
	// 无参数返回值和error
	value := assert.ShouldCall0RE(func() (int, error) {
		return 42, nil
	}, "获取值失败")

	fmt.Printf("获取到值: %d\n", value)

	// Output:
	// 获取到值: 42
}

func ExampleShouldCall1RE() {
	// 单参数返回值和error
	result := assert.ShouldCall1RE(func(x int) (string, error) {
		return fmt.Sprintf("数字是: %d", x), nil
	}, 42, "转换失败")

	fmt.Println(result)

	// Output:
	// 数字是: 42
}

func ExampleShouldCall0E() {
	// 无参数返回error
	assert.ShouldCall0E(func() error {
		fmt.Println("执行成功")
		return nil
	}, "操作失败")

	// Output:
	// 执行成功
}
