package assert_test

import (
	"errors"
	"testing"

	`github.com/spelens-gud/assert`
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

// TestMayWithError 测试 May 函数处理错误场景
func TestMayWithError(t *testing.T) {
	t.Run("使用May处理错误情况", func(t *testing.T) {
		err := errors.New("测试错误")
		errorHandled := false

		assert.May(err == nil,
			func() {
				// 成功时执行
			},
			func() {
				// 错误时执行
				errorHandled = true
			},
		)

		if !errorHandled {
			t.Errorf("错误未被处理")
		}
	})

	t.Run("使用May处理成功情况", func(t *testing.T) {
		var err error = nil
		successHandled := false

		assert.May(err == nil,
			func() {
				// 成功时执行
				successHandled = true
			},
			func() {
				// 错误时执行
			},
		)

		if !successHandled {
			t.Errorf("成功情况未被处理")
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

// BenchmarkMayFalse 性能测试
func BenchmarkMayFalse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.MayFalse(false, func() {})
	}
}

// BenchmarkThen 性能测试
func BenchmarkThen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.Then(true).Do(func() {}).Else(func() {})
	}
}
