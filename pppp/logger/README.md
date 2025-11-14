# Logger 日志组件

基于 zap 封装的高性能日志组件，专为微服务场景设计。

## 目录

- [特性](#特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [使用示例](#使用示例)
- [项目集成](#项目集成)
- [性能说明](#性能说明)
- [最佳实践](#最佳实践)

## 特性

- ✅ 高性能：基于 uber-go/zap 实现
- ✅ 服务标识：支持注入服务名称前缀
- ✅ 分级存储：支持按日志级别分文件存储
- ✅ 环境隔离：支持开发、测试、生产环境配置
- ✅ 多输出：支持控制台和文件同时输出
- ✅ 日志轮转：支持按大小、时间自动轮转
- ✅ 结构化日志：支持结构化字段记录
- ✅ 中间件支持：支持请求 ID、用户 ID、追踪 ID 等上下文注入
- ✅ 调用信息：可选的调用者信息和堆栈跟踪

## 安装

```bash
go get -u go.uber.org/zap
go get -u gopkg.in/natefinch/lumberjack.v2
```

## 快速开始

### 基本使用

```go
package main

import (
	"github.com/spelens-gud/trunk/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 创建日志配置
	config := &logger.Config{
		ServiceName: "user-service",
		Environment: "dev",
		Level:       "info",
		Console:     true,
		File:        true,
		FilePath:    "./logs",
	}

	// 创建 logger
	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// 输出日志
	log.Info("服务启动成功")
	log.Infof("监听端口: %d", 8080)
}
```

### 使用默认配置

```go
// 使用默认配置创建 logger
log, _ := logger.NewLogger(logger.DefaultConfig())
defer log.Sync()

log.Info("使用默认配置")
```

### 全局 Logger

```go
// 初始化全局 logger
config := logger.DefaultConfig()
config.ServiceName = "my-service"
logger.InitDefault(config)

// 在任何地方使用
logger.Info("全局日志")
logger.Infof("用户 %s 登录", "张三")
```

## 配置说明

### Config 结构

```go
type Config struct {
    ServiceName      string  // 服务名称（必填）
    Environment      string  // 环境：dev, test, prod
    Level            string  // 日志级别：debug, info, warn, error
    Console          bool    // 是否输出到控制台
    File             bool    // 是否输出到文件
    FilePath         string  // 文件路径
    FileName         string  // 文件名
    MaxSize          int     // 单个文件最大大小（MB）
    MaxAge           int     // 保留旧文件的最大天数
    MaxBackups       int     // 保留旧文件的最大个数
    Compress         bool    // 是否压缩旧文件
    EnableCaller     bool    // 是否启用调用者信息
    EnableStacktrace bool    // 是否启用堆栈跟踪
}
```

### 环境配置示例

#### 开发环境

```go
devConfig := &logger.Config{
    ServiceName:      "my-service",
    Environment:      "dev",
    Level:            "debug",
    Console:          true,
    File:             false,
    EnableCaller:     true,
    EnableStacktrace: true,
}
```

#### 生产环境

```go
prodConfig := &logger.Config{
    ServiceName:      "my-service",
    Environment:      "prod",
    Level:            "info",
    Console:          false,
    File:             true,
    FilePath:         "/var/log/myapp",
    FileName:         "app.log",
    MaxSize:          500,
    MaxAge:           90,
    MaxBackups:       30,
    Compress:         true,
    EnableCaller:     false,
    EnableStacktrace: false,
}
```

## 使用示例

### 结构化日志

```go
log.Info("用户登录",
    zap.String("user_id", "12345"),
    zap.String("username", "张三"),
    zap.String("ip", "192.168.1.1"),
    zap.Time("login_time", time.Now()),
)
```

### 格式化日志

```go
log.Infof("用户 %s 从 %s 登录", "张三", "192.168.1.1")
log.Debugf("处理请求耗时: %dms", 150)
```

### 添加前缀

```go
// 为不同模块添加前缀
orderLog := log.WithPrefix("order")
orderLog.Info("创建订单", zap.String("order_id", "12345"))

paymentLog := log.WithPrefix("payment")
paymentLog.Info("处理支付", zap.String("payment_id", "67890"))
```

### 添加上下文字段

```go
// 添加请求上下文
requestLog := log.With(
    zap.String("request_id", "req-123"),
    zap.String("user_id", "user-456"),
)

requestLog.Info("处理请求")
requestLog.Info("请求完成")
```

### 中间件使用

#### 请求追踪

```go
// 添加请求 ID
requestLog := logger.WithRequestID(log, "req-123456")

// 添加用户 ID
requestLog = logger.WithUserID(requestLog, "user-789")

// 添加追踪 ID（分布式追踪）
requestLog = logger.WithTraceID(requestLog, "trace-abc")

requestLog.Info("处理 API 请求")
```

#### 记录执行时间

```go
logger.WithDuration(log, "数据库查询", func() {
    // 执行数据库查询
    time.Sleep(100 * time.Millisecond)
})
// 自动记录：操作完成 operation=数据库查询 duration=100ms
```

#### 捕获 Panic

```go
logger.WithRecover(log, func() {
    // 可能会 panic 的代码
    riskyOperation()
})
```

#### 上下文字段构建器

```go
fields := logger.NewContextFields().
    AddString("batch_id", "batch-001").
    AddInt("record_count", 1000).
    AddString("source", "kafka").
    AddAny("metadata", map[string]interface{}{
        "version": "1.0",
        "region":  "us-east-1",
    })

contextLog := fields.ApplyToLogger(log)
contextLog.Info("开始处理数据批次")
```

## 项目集成

### 1. 在 root.go 中初始化全局日志

修改 `cmd/root.go`，在 `initConfig()` 函数中初始化日志：

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spelens-gud/trunk/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile  string
	logLevel string
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "日志级别")
}

func initConfig() {
	// 初始化配置
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("trunk")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}

	// 初始化日志
	initLogger()
}

// initLogger 初始化日志系统
func initLogger() {
	// 从配置文件或环境变量读取日志配置
	logConfig := &logger.Config{
		ServiceName:      viper.GetString("service.name"),
		Environment:      viper.GetString("environment"),
		Level:            logLevel,
		Console:          viper.GetBool("log.console"),
		File:             viper.GetBool("log.file"),
		FilePath:         viper.GetString("log.path"),
		FileName:         viper.GetString("log.filename"),
		MaxSize:          viper.GetInt("log.max_size"),
		MaxAge:           viper.GetInt("log.max_age"),
		MaxBackups:       viper.GetInt("log.max_backups"),
		Compress:         viper.GetBool("log.compress"),
		EnableCaller:     viper.GetBool("log.enable_caller"),
		EnableStacktrace: viper.GetBool("log.enable_stacktrace"),
	}

	// 设置默认值
	if logConfig.ServiceName == "" {
		logConfig.ServiceName = "trunk"
	}
	if logConfig.Environment == "" {
		logConfig.Environment = "dev"
	}
	if !logConfig.Console && !logConfig.File {
		logConfig.Console = true
	}

	// 初始化全局日志
	if err := logger.InitDefault(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	logger.Info("日志系统初始化成功",
		zap.String("service", logConfig.ServiceName),
		zap.String("environment", logConfig.Environment),
		zap.String("level", logConfig.Level),
	)
}
```

### 2. 在各个命令中使用日志

```go
package cmd

import (
	"github.com/spelens-gud/trunk/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var centerCmd = &cobra.Command{
	Use:   "center",
	Short: "启动 center 服务",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 为 center 命令创建专用 logger
		log := logger.GetDefault().WithPrefix("center")

		log.Info("Center 服务启动中...")

		// 业务逻辑
		if err := startCenterService(); err != nil {
			log.Error("Center 服务启动失败", zap.Error(err))
			return err
		}

		log.Info("Center 服务启动成功",
			zap.String("host", "0.0.0.0"),
			zap.Int("port", 8080),
		)

		return nil
	},
}
```

### 3. 配置文件示例

创建 `config/trunk.yaml`：

```yaml
# 服务配置
service:
  name: trunk

# 环境配置
environment: dev # dev, test, prod

# 日志配置
log:
  console: true
  file: true
  path: ./logs
  filename: trunk.log
  max_size: 100 # MB
  max_age: 30 # 天
  max_backups: 10 # 个
  compress: true
  enable_caller: true
  enable_stacktrace: true
```

生产环境配置：

```yaml
service:
  name: trunk

environment: prod

log:
  console: false
  file: true
  path: /var/log/trunk
  filename: trunk.log
  max_size: 500
  max_age: 90
  max_backups: 30
  compress: true
  enable_caller: false
  enable_stacktrace: false
```

### 4. 环境变量配置

也可以通过环境变量配置（优先级高于配置文件）：

```bash
export SERVICE_NAME=trunk
export ENVIRONMENT=prod
export LOG_LEVEL=info
export LOG_CONSOLE=false
export LOG_FILE=true
export LOG_FILE_PATH=/var/log/trunk
export LOG_FILE_NAME=trunk.log
```

### 5. 命令行使用

```bash
# 使用默认配置
./trunk center

# 指定日志级别
./trunk center --log-level debug

# 指定配置文件
./trunk center --config /path/to/config.yaml

# 使用环境变量
LOG_LEVEL=debug ./trunk center
```

## 性能说明

基于 Apple M4 芯片的性能测试结果：

- **结构化日志**：22.01 ns/op，128 B/op，1 allocs/op
- **格式化日志**：17.71 ns/op，24 B/op，1 allocs/op

性能优于标准库 log 和大多数其他日志库。

## 最佳实践

### 1. 服务启动时初始化

```go
func main() {
    // 从配置文件或环境变量读取配置
    config := loadLogConfig()

    log, err := logger.NewLogger(config)
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    // 设置为全局 logger
    logger.InitDefault(config)

    // 启动服务
    startService(log)
}
```

### 2. 为每个请求创建带上下文的 Logger

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 为每个请求创建独立的 logger
    requestLog := logger.GetDefault().With(
        zap.String("request_id", generateRequestID()),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
        zap.String("remote_addr", r.RemoteAddr),
    )

    requestLog.Info("收到请求")

    // 传递给业务逻辑
    processRequest(requestLog, r)

    requestLog.Info("请求完成")
}
```

### 3. 错误处理

```go
result, err := doSomething()
if err != nil {
    log.Error("操作失败",
        zap.Error(err),
        zap.String("operation", "doSomething"),
        zap.Any("input", input),
    )
    return err
}
```

### 4. 性能监控

```go
start := time.Now()
result := expensiveOperation()
duration := time.Since(start)

log.Info("操作完成",
    zap.String("operation", "expensiveOperation"),
    zap.Duration("duration", duration),
    zap.Int("result_count", len(result)),
)
```

## 注意事项

1. 务必在程序退出前调用 `log.Sync()` 刷新缓冲区
2. 生产环境建议将日志级别设置为 `info` 或更高
3. 文件输出路径需要有写入权限
4. 合理设置日志轮转参数，避免磁盘空间不足
5. 敏感信息不要直接记录到日志中
6. 生产环境建议关闭 `EnableCaller` 和 `EnableStacktrace` 以提升性能

## 故障排查

### 日志文件没有生成

- 检查 `log.file` 是否设置为 `true`
- 检查 `log.path` 目录是否有写入权限
- 检查磁盘空间是否充足

### 日志级别不生效

- 检查命令行参数 `--log-level` 是否正确
- 检查配置文件中的日志级别设置
- 环境变量 `LOG_LEVEL` 优先级最高

### 日志输出乱码

- 确保终端支持 UTF-8 编码
- 生产环境建议使用 JSON 格式输出

## License

MIT
