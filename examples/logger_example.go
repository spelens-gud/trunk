package main

import (
	"time"

	"github.com/spelens-gud/logger"
	"go.uber.org/zap"
)

func main() {
	// ========== 示例 1: 基本使用 ==========
	basicUsage()

	// ========== 示例 2: 使用前缀 ==========
	withPrefix()

	// ========== 示例 3: 使用中间件 ==========
	withMiddleware()

	// ========== 示例 4: 上下文字段构建器 ==========
	contextFields()

	// ========== 示例 5: 不同环境配置 ==========
	environmentSpecific()

	// ========== 示例 6: 全局 Logger ==========
	globalLogger()

	// ========== 示例 7: 结构化日志 ==========
	structuredLogging()
}

// basicUsage 基本使用示例
func basicUsage() {
	// 创建日志配置
	config := &logger.Config{
		ServiceName:      "user-service",
		Environment:      "dev",
		Level:            "debug",
		Console:          true,
		File:             true,
		FilePath:         "./logs",
		FileName:         "app.log",
		MaxSize:          100,
		MaxAge:           30,
		MaxBackups:       10,
		Compress:         true,
		EnableCaller:     true,
		EnableStacktrace: true,
	}

	// 创建 logger
	log, err := logger.NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// 基本日志输出
	log.Info("服务启动")
	log.Debug("调试信息", zap.String("module", "main"))
	log.Warn("警告信息", zap.Int("retry_count", 3))

	// 格式化输出
	log.Infof("用户 %s 登录成功", "张三")
	log.Debugf("处理请求耗时: %dms", 150)
}

// withPrefix 使用前缀示例
func withPrefix() {
	config := logger.DefaultConfig()
	config.ServiceName = "order-service"

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	// 为不同模块添加前缀
	orderLog := log.WithPrefix("order-module")
	orderLog.Info("创建订单", zap.String("order_id", "12345"))

	paymentLog := log.WithPrefix("payment-module")
	paymentLog.Info("处理支付", zap.String("payment_id", "67890"))
}

// withMiddleware 使用中间件示例
func withMiddleware() {
	config := logger.DefaultConfig()
	config.ServiceName = "api-gateway"

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	// 添加请求上下文
	requestLog := logger.WithRequestID(log, "req-123456")
	requestLog = logger.WithUserID(requestLog, "user-789")
	requestLog = logger.WithTraceID(requestLog, "trace-abc")

	requestLog.Info("处理 API 请求",
		zap.String("method", "POST"),
		zap.String("path", "/api/users"),
	)

	// 记录执行时间
	logger.WithDuration(log, "数据库查询", func() {
		time.Sleep(100 * time.Millisecond)
		log.Info("查询完成")
	})

	// 捕获 panic
	logger.WithRecover(log, func() {
		log.Info("执行危险操作")
		// panic("something went wrong")
	})
}

// contextFields 使用上下文字段构建器示例
func contextFields() {
	config := logger.DefaultConfig()
	config.ServiceName = "data-processor"

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	// 构建上下文字段
	fields := logger.NewContextFields().
		AddString("batch_id", "batch-001").
		AddInt("record_count", 1000).
		AddString("source", "kafka").
		AddAny("metadata", map[string]interface{}{
			"version": "1.0",
			"region":  "us-east-1",
		})

	// 应用到 logger
	contextLog := fields.ApplyToLogger(log)
	contextLog.Info("开始处理数据批次")
	contextLog.Info("数据处理完成")
}

// environmentSpecific 不同环境配置示例
func environmentSpecific() {
	// 开发环境配置
	devConfig := &logger.Config{
		ServiceName:      "my-service",
		Environment:      "dev",
		Level:            "debug",
		Console:          true,
		File:             false,
		EnableCaller:     true,
		EnableStacktrace: true,
	}

	// 生产环境配置
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

	// 根据环境选择配置
	env := "dev" // 从环境变量读取
	var config *logger.Config
	if env == "prod" {
		config = prodConfig
	} else {
		config = devConfig
	}

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	log.Info("服务启动", zap.String("environment", env))
}

// globalLogger 使用全局 logger 示例
func globalLogger() {
	// 初始化全局 logger
	config := &logger.Config{
		ServiceName: "global-service",
		Environment: "dev",
		Level:       "info",
		Console:     true,
	}

	logger.InitDefault(config)

	// 在任何地方使用全局 logger
	logger.Info("这是全局日志")
	logger.Infof("用户 %s 执行了操作", "李四")

	// 获取全局 logger 实例进行扩展
	log := logger.GetDefault()
	moduleLog := log.WithPrefix("auth-module")
	moduleLog.Info("用户认证成功")
}

// structuredLogging 结构化日志示例
func structuredLogging() {
	config := logger.DefaultConfig()
	config.ServiceName = "transaction-service"

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	// 记录结构化数据
	log.Info("交易创建",
		zap.String("transaction_id", "txn-123"),
		zap.String("user_id", "user-456"),
		zap.Float64("amount", 99.99),
		zap.String("currency", "USD"),
		zap.String("status", "pending"),
		zap.Time("created_at", time.Now()),
		zap.Strings("tags", []string{"online", "credit-card"}),
	)

	// 记录错误详情
	log.Error("交易失败",
		zap.String("transaction_id", "txn-124"),
		zap.String("error_code", "INSUFFICIENT_FUNDS"),
		zap.String("error_message", "账户余额不足"),
		zap.Float64("required_amount", 199.99),
		zap.Float64("available_balance", 50.00),
	)
}
