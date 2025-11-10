package assert

// May 根据条件是否成立执行相应的回调函数
// 如果条件为真，执行 onTrue 回调（如果提供）
// 如果条件为假，执行 onFalse 回调（如果提供）
func May(condition bool, onTrue func(), onFalse func()) {
	if condition {
		if onTrue != nil {
			onTrue()
		}
	} else {
		if onFalse != nil {
			onFalse()
		}
	}
}

// MayTrue 当条件为真时执行回调
func MayTrue(condition bool, callback func()) {
	if condition && callback != nil {
		callback()
	}
}

// MayFalse 当条件为假时执行回调
func MayFalse(condition bool, callback func()) {
	if !condition && callback != nil {
		callback()
	}
}

// mayNoError 当错误为 nil 时执行 onSuccess 回调，否则执行 onError 回调
func mayNoError(err error, onSuccess func(), onError func(error)) {
	if err == nil {
		if onSuccess != nil {
			onSuccess()
		}
	} else {
		if onError != nil {
			onError(err)
		}
	}
}

// MayValue 根据错误状态执行回调并返回值
// 如果错误为 nil，执行 onSuccess 回调（如果提供）并返回值
// 如果错误不为 nil，执行 onError 回调（如果提供）并返回零值
func MayValue[T any](value T, err error, onSuccess func(T), onError func(error)) T {
	if err == nil {
		if onSuccess != nil {
			onSuccess(value)
		}
		return value
	}

	if onError != nil {
		onError(err)
	}
	var zero T

	return zero
}

// MayFunc 执行函数，根据返回的错误状态执行相应回调
func MayFunc(f func() error, onSuccess func(), onError func(error)) {
	err := f()
	mayNoError(err, onSuccess, onError)
}

// MayFuncValue 执行函数，根据返回的错误状态执行相应回调并返回值
func MayFuncValue[T any](f func() (T, error), onSuccess func(T), onError func(error)) T {
	value, err := f()

	return MayValue(value, err, onSuccess, onError)
}

// MayElse 提供链式调用的条件执行
type MayElse struct {
	condition bool
	executed  bool
}

// Then 创建一个可链式调用的条件执行器
func Then(condition bool) *MayElse {
	return &MayElse{condition: condition, executed: false}
}

// Do 当条件为真且尚未执行时执行回调
func (m *MayElse) Do(callback func()) *MayElse {
	if m.condition && !m.executed && callback != nil {
		callback()
		m.executed = true
	}

	return m
}

// Else 当条件为假且尚未执行时执行回调
func (m *MayElse) Else(callback func()) *MayElse {
	if !m.condition && !m.executed && callback != nil {
		callback()
		m.executed = true
	}

	return m
}
