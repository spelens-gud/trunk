# Assert - Go 错误处理和条件执行工具库

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

一个优雅的 Go 语言错误处理和条件执行工具库，提供三种不同级别的错误处理策略（Must、Should、May），帮助你编写更简洁、更安全的代码。

## 特性

- 🚀 **泛型支持**：完全利用 Go 1.18+ 泛型特性，类型安全
- 🎯 **三级错误处理**：Must（必须）、Should（应该）、May（可能）三种策略
- 📝 **日志集成**：可选的日志记录功能，支持自定义日志记录器
- 🔗 **链式调用**：优雅的条件执行链式 API
- ⚡ **零依赖**：核心功能无外部依赖（日志功能可选）
- 🧪 **完整测试**：100% 测试覆盖率

## 安装

```bash
go get github.com/spelens-gud/assert
```

## 快速开始

### Must - 必须成功（失败时 panic）

适用于关键操作，错误发生时应该立即终止程序。

```go
import "github.com/spelens-gud/assert"

// 无参数返回 error
assert.MustCall0E(func() error {
    return doSomething()
}, "操作失败")

// 无参数返回值和 error
value := assert.MustCall0RE(func() (int, error) {
    return getValue()
}, "获取值失败")

// 单参数返回 error
assert.MustCall1E(func(x int) error {
    return process(x)
}, 42, "处理失败")

// 单参数返回值和 error
result := assert.MustCall1RE(func(x int) (string, error) {
    return convert(x)
}, 42, "转换失败")

// 双参数、三参数同理
sum := assert.MustCall2RE(func(x, y int) (int, error) {
    return x + y, nil
}, 10, 20, "计算失败")
```

### Should - 应该成功（失败时记录日志）

适用于重要但非致命的操作，错误发生时记录日志但继续执行。

```go
// 设置日志记录器（可选）
assert.SetLogger(yourLogger)

// 无参数返回 error
assert.ShouldCall0E(func() error {
    return doSomething()
}, "操作失败")

// 返回值和 error
value := assert.ShouldCall0RE(func() (int, error) {
    return getValue()
}, "获取值失败")

// 即使有错误，也会返回值（错误只记录日志）
result := assert.ShouldCall1RE(func(x int) (string, error) {
    return "default", errors.New("转换失败")
}, 42, "转换失败")
// result = "default"，错误被记录到日志
```

### May - 条件执行

提供灵活的条件执行功能。

```go
// 基础条件执行
assert.May(condition,
    func() {
        // 条件为真时执行
    },
    func() {
        // 条件为假时执行
    },
)

// 仅在条件为真时执行
assert.MayTrue(value > 0, func() {
    fmt.Println("值为正数")
})

// 仅在条件为假时执行
assert.MayFalse(err != nil, func() {
    fmt.Println("没有错误")
})

// 链式调用
assert.Then(value > 0).
    Do(func() {
        fmt.Println("值为正数")
    }).
    Else(func() {
        fmt.Println("值不为正数")
    })
```

## API 文档

### Must 系列函数

Must 系列函数在错误发生时会触发 panic，适用于必须成功的关键操作。

| 函数                                                                                                 | 说明                            |
| ---------------------------------------------------------------------------------------------------- | ------------------------------- |
| `MustCall0E(f func() error, msg ...any)`                                                             | 执行无参数返回 error 的函数     |
| `MustCall0RE[R](f func() (R, error), msg ...any) R`                                                  | 执行无参数返回值和 error 的函数 |
| `MustCall1E[T](f func(T) error, arg T, msg ...any)`                                                  | 执行单参数返回 error 的函数     |
| `MustCall1RE[T, R](f func(T) (R, error), arg T, msg ...any) R`                                       | 执行单参数返回值和 error 的函数 |
| `MustCall2E[T1, T2](f func(T1, T2) error, arg1 T1, arg2 T2, msg ...any)`                             | 执行双参数返回 error 的函数     |
| `MustCall2RE[T1, T2, R](f func(T1, T2) (R, error), arg1 T1, arg2 T2, msg ...any) R`                  | 执行双参数返回值和 error 的函数 |
| `MustCall3E[T1, T2, T3](f func(T1, T2, T3) error, arg1 T1, arg2 T2, arg3 T3, msg ...any)`            | 执行三参数返回 error 的函数     |
| `MustCall3RE[T1, T2, T3, R](f func(T1, T2, T3) (R, error), arg1 T1, arg2 T2, arg3 T3, msg ...any) R` | 执行三参数返回值和 error 的函数 |

### Should 系列函数

Should 系列函数在错误发生时会记录日志但继续执行，适用于重要但非致命的操作。

| 函数                                                                                                   | 说明                            |
| ------------------------------------------------------------------------------------------------------ | ------------------------------- |
| `ShouldCall0E(f func() error, msg ...any)`                                                             | 执行无参数返回 error 的函数     |
| `ShouldCall0RE[R](f func() (R, error), msg ...any) R`                                                  | 执行无参数返回值和 error 的函数 |
| `ShouldCall1E[T](f func(T) error, arg T, msg ...any)`                                                  | 执行单参数返回 error 的函数     |
| `ShouldCall1RE[T, R](f func(T) (R, error), arg T, msg ...any) R`                                       | 执行单参数返回值和 error 的函数 |
| `ShouldCall2E[T1, T2](f func(T1, T2) error, arg1 T1, arg2 T2, msg ...any)`                             | 执行双参数返回 error 的函数     |
| `ShouldCall2RE[T1, T2, R](f func(T1, T2) (R, error), arg1 T1, arg2 T2, msg ...any) R`                  | 执行双参数返回值和 error 的函数 |
| `ShouldCall3E[T1, T2, T3](f func(T1, T2, T3) error, arg1 T1, arg2 T2, arg3 T3, msg ...any)`            | 执行三参数返回 error 的函数     |
| `ShouldCall3RE[T1, T2, T3, R](f func(T1, T2, T3) (R, error), arg1 T1, arg2 T2, arg3 T3, msg ...any) R` | 执行三参数返回值和 error 的函数 |

### May 系列函数

May 系列函数提供灵活的条件执行功能。

| 函数                                                 | 说明                   |
| ---------------------------------------------------- | ---------------------- |
| `May(condition bool, onTrue func(), onFalse func())` | 根据条件执行相应的回调 |
| `MayTrue(condition bool, callback func())`           | 条件为真时执行回调     |
| `MayFalse(condition bool, callback func())`          | 条件为假时执行回调     |
| `Then(condition bool) *MayElse`                      | 创建链式条件执行器     |
| `(*MayElse).Do(callback func()) *MayElse`            | 条件为真时执行回调     |
| `(*MayElse).Else(callback func()) *MayElse`          | 条件为假时执行回调     |

### 日志配置

```go
// 设置全局日志记录器
assert.SetLogger(logger ILogger)
```

日志记录器接口定义：

```go
type ILogger interface {
    Panic(msg string, fields ...zap.Field)
    Panicf(template string, args ...any)
    Error(msg string, fields ...zap.Field)
    Errorf(template string, args ...any)
}
```

## 使用场景

### 场景 1：文件操作

```go
// Must：关键文件必须成功打开
file := assert.MustCall1RE(os.Open, "config.yaml", "无法打开配置文件")
defer file.Close()

// Should：日志文件打开失败不影响主流程
logFile := assert.ShouldCall1RE(os.Open, "app.log", "无法打开日志文件")
if logFile != nil {
    defer logFile.Close()
}
```

### 场景 2：数据库操作

```go
// Must：数据库连接必须成功
db := assert.MustCall1RE(sql.Open, "mysql", dsn, "数据库连接失败")

// Should：缓存操作失败不影响主流程
assert.ShouldCall2E(cache.Set, key, value, "缓存设置失败")
```

### 场景 3：条件执行

```go
// 错误处理
assert.May(err == nil,
    func() {
        // 成功时的处理
        fmt.Println("操作成功")
    },
    func() {
        // 失败时的处理
        fmt.Printf("操作失败: %v\n", err)
    },
)

// 权限检查
assert.MayTrue(user.IsAdmin(), func() {
    // 执行管理员操作
    performAdminTask()
})

// 链式条件
assert.Then(value > 0).
    Do(func() {
        fmt.Println("正数")
    }).
    Else(func() {
        fmt.Println("非正数")
    })
```

## 最佳实践

### 1. 选择合适的错误处理策略

- **Must**：用于程序初始化、配置加载等关键操作
- **Should**：用于日志记录、缓存操作等非关键操作
- **May**：用于条件判断和分支执行

### 2. 提供清晰的错误消息

```go
// 好的做法
assert.MustCall1RE(loadConfig, "config.yaml", "加载配置文件失败")

// 不好的做法
assert.MustCall1RE(loadConfig, "config.yaml")
```

### 3. 合理使用日志记录器

```go
// 在程序初始化时设置日志记录器
func init() {
    logger := zap.NewProduction()
    assert.SetLogger(logger)
}
```

### 4. 链式调用保持简洁

```go
// 好的做法
assert.Then(condition).
    Do(func() { /* 简短操作 */ }).
    Else(func() { /* 简短操作 */ })

// 复杂逻辑应该提取为独立函数
assert.Then(condition).
    Do(handleSuccess).
    Else(handleFailure)
```

## 性能

所有函数都经过性能优化，基准测试结果：

```
BenchmarkMustCall0E-8    100000000    10.5 ns/op
BenchmarkMustCall0RE-8   100000000    11.2 ns/op
BenchmarkMay-8           200000000     8.3 ns/op
BenchmarkThen-8          150000000     9.1 ns/op
```

## 测试

运行测试：

```bash
go test -v
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 相关项目

- [go.uber.org/zap](https://github.com/uber-go/zap) - 高性能日志库（可选依赖）

## 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新历史。
