package logger_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
	"go.uber.org/zap"
)

// TestNewLogger 测试创建 logger
func TestNewLogger(t *testing.T) {
	config := &logger.Config{
		ServiceName: "test-service",
		Environment: "test",
		Level:       "debug",
		Console:     true,
		File:        false,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	defer log.Sync()

	if log == nil {
		t.Fatal("logger 不应该为 nil")
	}
}

// TestLoggerWithFields 测试结构化字段
func TestLoggerWithFields(t *testing.T) {
	config := logger.DefaultConfig()
	config.ServiceName = "test-service"

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	defer log.Sync()

	// 测试结构化日志
	log.Info("测试消息",
		zap.String("key1", "value1"),
		zap.Int("key2", 123),
	)
}

// TestLoggerWithPrefix 测试前缀功能
func TestLoggerWithPrefix(t *testing.T) {
	config := logger.DefaultConfig()
	config.ServiceName = "test-service"

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	defer log.Sync()

	// 测试前缀
	prefixLog := log.WithPrefix("module1")
	prefixLog.Info("带前缀的消息")
}

// TestLoggerFormatted 测试格式化输出
func TestLoggerFormatted(t *testing.T) {
	config := logger.DefaultConfig()
	config.ServiceName = "test-service"

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	defer log.Sync()

	// 测试格式化输出
	log.Infof("格式化消息: %s, %d", "test", 123)
	log.Debugf("调试消息: %v", map[string]int{"a": 1})
}

// TestLoggerFileOutput 测试文件输出
func TestLoggerFileOutput(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	config := &logger.Config{
		ServiceName: "test-service",
		Environment: "test",
		Level:       "debug",
		Console:     false,
		File:        true,
		FilePath:    tmpDir,
		FileName:    "test.log",
		MaxSize:     10,
		MaxAge:      1,
		MaxBackups:  3,
		Compress:    false,
	}

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}

	// 写入日志
	log.Info("测试文件输出")
	log.Error("测试错误日志")
	log.Sync()

	// 检查文件是否存在
	logFile := filepath.Join(tmpDir, "test.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("日志文件不存在: %s", logFile)
	}

	errorLogFile := filepath.Join(tmpDir, "error.log")
	if _, err := os.Stat(errorLogFile); os.IsNotExist(err) {
		t.Errorf("错误日志文件不存在: %s", errorLogFile)
	}
}

// TestDefaultLogger 测试全局 logger
func TestDefaultLogger(t *testing.T) {
	config := logger.DefaultConfig()
	config.ServiceName = "test-global"

	err := logger.InitDefault(config)
	if err != nil {
		t.Fatalf("初始化全局 logger 失败: %v", err)
	}

	// 使用全局 logger
	logger.Info("全局日志测试")
	logger.Infof("格式化全局日志: %s", "test")

	// 获取全局 logger
	log := logger.GetDefault()
	if log == nil {
		t.Fatal("全局 logger 不应该为 nil")
	}
}

// TestMiddleware 测试中间件
func TestMiddleware(t *testing.T) {
	config := logger.DefaultConfig()
	config.ServiceName = "test-middleware"

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	defer log.Sync()

	// 测试 WithRequestID
	requestLog := logger.WithRequestID(log, "req-123")
	requestLog.Info("测试请求 ID")

	// 测试 WithUserID
	userLog := logger.WithUserID(log, "user-456")
	userLog.Info("测试用户 ID")

	// 测试 WithTraceID
	traceLog := logger.WithTraceID(log, "trace-789")
	traceLog.Info("测试追踪 ID")

	// 测试 WithDuration
	logger.WithDuration(log, "测试操作", func() {
		time.Sleep(10 * time.Millisecond)
	})

	// 测试 WithRecover
	logger.WithRecover(log, func() {
		log.Info("测试 recover")
	})
}

// TestContextFields 测试上下文字段构建器
func TestContextFields(t *testing.T) {
	config := logger.DefaultConfig()
	config.ServiceName = "test-context"

	log, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	defer log.Sync()

	// 测试上下文字段构建器
	fields := logger.NewContextFields().
		AddString("key1", "value1").
		AddInt("key2", 123).
		AddBool("key3", true).
		AddFloat64("key4", 3.14).
		AddAny("key5", map[string]any{"nested": "value"})

	contextLog := fields.ApplyToLogger(log)
	contextLog.Info("测试上下文字段")
}

// BenchmarkStructuredLog 基准测试：结构化日志
func BenchmarkStructuredLog(b *testing.B) {
	config := &logger.Config{
		ServiceName: "bench-service",
		Environment: "test",
		Level:       "info",
		Console:     false,
		File:        false,
	}

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("benchmark test",
			zap.String("key", "value"),
			zap.Int("number", 123),
		)
	}
}

// BenchmarkFormattedLog 基准测试：格式化日志
func BenchmarkFormattedLog(b *testing.B) {
	config := &logger.Config{
		ServiceName: "bench-service",
		Environment: "test",
		Level:       "info",
		Console:     false,
		File:        false,
	}

	log, _ := logger.NewLogger(config)
	defer log.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Infof("benchmark test: %s, %d", "value", 123)
	}
}
