package assert_test

import (
	"errors"
	"testing"

	"github.com/spelens-gud/trunk/internal/assert"
)

// TestMay 测试 May 函数
func TestMay(t *testing.T) {
	t.Run("条件为真应执行onTrue回调", func(t *testing.T) {
		trueCalled := false
		falseCalled := false

		assert.May(true,
			func() { trueCalled = true },
			func() { falseCalled = true },
		)

		if !trueCalled {
			t.Errorf("onTrue 回调未被调用")
		}
		if falseCalled {
			t.Errorf("onFalse 回调不应被调用")
		}
	})

	t.Run("条件为假应执行onFalse回调", func(t *testing.T) {
		trueCalled := false
		falseCalled := false

		assert.May(false,
			func() { trueCalled = true },
			func() { falseCalled = true },
		)

		if trueCalled {
			t.Errorf("onTrue 回调不应被调用")
		}
		if !falseCalled {
			t.Errorf("onFalse 回调未被调用")
		}
	})

	t.Run("onTrue为nil不应panic", func(t *testing.T) {
		assert.May(true, nil, func() {})
	})

	t.Run("onFalse为nil不应panic", func(t *testing.T) {
		assert.May(false, func() {}, nil)
	})

	t.Run("两个回调都为nil不应panic", func(t *testing.T) {
		assert.May(true, nil, nil)
		assert.May(false, nil, nil)
	})
}

// TestMayTrue 测试 MayTrue 函数
func TestMayTrue(t *testing.T) {
	t.Run("条件为真应执行回调", func(t *testing.T) {
		called := false
		assert.MayTrue(true, func() { called = true })
		if !called {
			t.Errorf("回调未被调用")
		}
	})

	t.Run("条件为假不应执行回调", func(t *testing.T) {
		called := false
		assert.MayTrue(false, func() { called = true })
		if called {
			t.Errorf("回调不应被调用")
		}
	})

	t.Run("回调为nil不应panic", func(t *testing.T) {
		assert.MayTrue(true, nil)
		assert.MayTrue(false, nil)
	})
}

// TestMayFalse 测试 MayFalse 函数
func TestMayFalse(t *testing.T) {
	t.Run("条件为假应执行回调", func(t *testing.T) {
		called := false
		assert.MayFalse(false, func() { called = true })
		if !called {
			t.Errorf("回调未被调用")
		}
	})

	t.Run("条件为真不应执行回调", func(t *testing.T) {
		called := false
		assert.MayFalse(true, func() { called = true })
		if called {
			t.Errorf("回调不应被调用")
		}
	})

	t.Run("回调为nil不应panic", func(t *testing.T) {
		assert.MayFalse(true, nil)
		assert.MayFalse(false, nil)
	})
}

// TestMayValue 测试 MayValue 函数
func TestMayValue(t *testing.T) {
	t.Run("错误为nil应执行onSuccess回调并返回值", func(t *testing.T) {
		var receivedValue int
		result := assert.MayValue(42, nil,
			func(v int) { receivedValue = v },
			func(err error) {},
		)

		if result != 42 {
			t.Errorf("返回值 = %v, 期望 42", result)
		}
		if receivedValue != 42 {
			t.Errorf("onSuccess 接收到的值 = %v, 期望 42", receivedValue)
		}
	})

	t.Run("错误不为nil应执行onError回调并返回零值", func(t *testing.T) {
		var receivedErr error
		testErr := errors.New("测试错误")

		result := assert.MayValue(42, testErr,
			func(v int) {},
			func(err error) { receivedErr = err },
		)

		if result != 0 {
			t.Errorf("返回值 = %v, 期望 0", result)
		}
		if receivedErr != testErr {
			t.Errorf("onError 接收到的错误 = %v, 期望 %v", receivedErr, testErr)
		}
	})

	t.Run("泛型支持字符串类型", func(t *testing.T) {
		result := assert.MayValue("hello", nil, nil, nil)
		if result != "hello" {
			t.Errorf("返回值 = %v, 期望 hello", result)
		}

		result = assert.MayValue("hello", errors.New("错误"), nil, nil)
		if result != "" {
			t.Errorf("返回值 = %v, 期望空字符串", result)
		}
	})

	t.Run("泛型支持结构体类型", func(t *testing.T) {
		type User struct {
			Name string
		}
		user := User{Name: "张三"}

		result := assert.MayValue(user, nil, nil, nil)
		if result.Name != "张三" {
			t.Errorf("返回值 = %v, 期望 张三", result.Name)
		}

		result = assert.MayValue(user, errors.New("错误"), nil, nil)
		if result.Name != "" {
			t.Errorf("返回值应该是零值")
		}
	})

	t.Run("回调为nil不应panic", func(t *testing.T) {
		assert.MayValue(42, nil, nil, nil)
		assert.MayValue(42, errors.New("错误"), nil, nil)
	})
}

// TestMayFunc 测试 MayFunc 函数
func TestMayFunc(t *testing.T) {
	t.Run("函数返回nil应执行onSuccess回调", func(t *testing.T) {
		funcCalled := false
		successCalled := false

		assert.MayFunc(
			func() error {
				funcCalled = true
				return nil
			},
			func() { successCalled = true },
			func(err error) {},
		)

		if !funcCalled {
			t.Errorf("函数未被调用")
		}
		if !successCalled {
			t.Errorf("onSuccess 回调未被调用")
		}
	})

	t.Run("函数返回错误应执行onError回调", func(t *testing.T) {
		testErr := errors.New("测试错误")
		var receivedErr error

		assert.MayFunc(
			func() error { return testErr },
			func() {},
			func(err error) { receivedErr = err },
		)

		if receivedErr != testErr {
			t.Errorf("onError 接收到的错误 = %v, 期望 %v", receivedErr, testErr)
		}
	})
}

// TestMayFuncValue 测试 MayFuncValue 函数
func TestMayFuncValue(t *testing.T) {
	t.Run("函数返回值和nil错误应执行onSuccess回调", func(t *testing.T) {
		var receivedValue int
		result := assert.MayFuncValue(
			func() (int, error) { return 42, nil },
			func(v int) { receivedValue = v },
			func(err error) {},
		)

		if result != 42 {
			t.Errorf("返回值 = %v, 期望 42", result)
		}
		if receivedValue != 42 {
			t.Errorf("onSuccess 接收到的值 = %v, 期望 42", receivedValue)
		}
	})

	t.Run("函数返回值和错误应执行onError回调", func(t *testing.T) {
		testErr := errors.New("测试错误")
		var receivedErr error

		result := assert.MayFuncValue(
			func() (int, error) { return 42, testErr },
			func(v int) {},
			func(err error) { receivedErr = err },
		)

		if result != 0 {
			t.Errorf("返回值 = %v, 期望 0", result)
		}
		if receivedErr != testErr {
			t.Errorf("onError 接收到的错误 = %v, 期望 %v", receivedErr, testErr)
		}
	})
}

// TestThen 测试链式调用
func TestThen(t *testing.T) {
	t.Run("条件为真应执行Do回调", func(t *testing.T) {
		called := false
		assert.Then(true).Do(func() { called = true })
		if !called {
			t.Errorf("Do 回调未被调用")
		}
	})

	t.Run("条件为假应执行Else回调", func(t *testing.T) {
		called := false
		assert.Then(false).Else(func() { called = true })
		if !called {
			t.Errorf("Else 回调未被调用")
		}
	})

	t.Run("条件为真不应执行Else回调", func(t *testing.T) {
		called := false
		assert.Then(true).Do(func() {}).Else(func() { called = true })
		if called {
			t.Errorf("Else 回调不应被调用")
		}
	})

	t.Run("条件为假不应执行Do回调", func(t *testing.T) {
		called := false
		assert.Then(false).Do(func() { called = true }).Else(func() {})
		if called {
			t.Errorf("Do 回调不应被调用")
		}
	})

	t.Run("链式调用只执行一次", func(t *testing.T) {
		count := 0
		assert.Then(true).
			Do(func() { count++ }).
			Do(func() { count++ }).
			Else(func() { count++ })

		if count != 1 {
			t.Errorf("回调执行次数 = %v, 期望 1", count)
		}
	})

	t.Run("回调为nil不应panic", func(t *testing.T) {
		assert.Then(true).Do(nil).Else(nil)
		assert.Then(false).Do(nil).Else(nil)
	})

	t.Run("复杂链式调用", func(t *testing.T) {
		result := ""
		assert.Then(false).
			Do(func() { result = "do1" }).
			Do(func() { result = "do2" }).
			Else(func() { result = "else1" }).
			Else(func() { result = "else2" })

		if result != "else1" {
			t.Errorf("result = %v, 期望 else1", result)
		}
	})
}

// 示例：演示 May 的使用
func ExampleMay() {
	value := 10

	assert.May(value > 5,
		func() {
			// 条件为真时执行
			println("值大于5")
		},
		func() {
			// 条件为假时执行
			println("值不大于5")
		},
	)
}

// 示例：演示 MayTrue 的使用
func ExampleMayTrue() {
	value := 10

	assert.MayTrue(value > 5, func() {
		println("值大于5")
	})
}

// 示例：演示 Then 链式调用的使用
func ExampleThen() {
	value := 10

	assert.Then(value > 5).
		Do(func() {
			println("值大于5")
		}).
		Else(func() {
			println("值不大于5")
		})
}

// BenchmarkMay 性能测试
func BenchmarkMay(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.May(true, func() {}, func() {})
	}
}

// BenchmarkMayTrue 性能测试
func BenchmarkMayTrue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.MayTrue(true, func() {})
	}
}

// BenchmarkMayValue 性能测试
func BenchmarkMayValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = assert.MayValue(42, nil, func(v int) {}, func(err error) {})
	}
}

// BenchmarkThen 性能测试
func BenchmarkThen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.Then(true).Do(func() {}).Else(func() {})
	}
}
