package grpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
)

// 测试辅助函数：创建测试客户端配置
func createTestClientConfig(host string) *ClientConfig {
	return &ClientConfig{
		Name:             "test-client",
		Host:             host,
		KeepAliveTime:    10 * time.Second,
		KeepAliveTimeout: 3 * time.Second,
		ReconnectEnabled: true,
		ReconnectDelay:   2 * time.Second,
		MaxReconnect:     3,
		OnReconnect: func(client *NetGrpcClient) {
			// 重连成功回调
		},
		OnDisconnect: func(client *NetGrpcClient) {
			// 断开连接回调
		},
	}
}

// TestNetGrpcClient_New 测试客户端创建
func TestNetGrpcClient_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log: log,
		cnf: createTestClientConfig("localhost:50051"),
	}

	client.New()

	if client.stopChan == nil {
		t.Error("stopChan 未初始化")
	}
	if !client.isStop {
		t.Error("isStop 应该为 true")
	}
}

// TestNetGrpcClient_IsConnected 测试连接状态检查
func TestNetGrpcClient_IsConnected(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:    log,
		cnf:    createTestClientConfig("localhost:50051"),
		isStop: true,
		conn:   nil,
	}

	client.New()

	// 未连接状态
	if client.IsConnected() {
		t.Error("期望未连接，但返回已连接")
	}

	// 模拟已连接状态
	client.isStop = false
	// 注意：这里不创建真实连接，只是测试逻辑
	// 实际使用中 conn 不应为 nil

	if client.IsConnected() {
		t.Log("连接状态检查通过")
	}
}

// TestNetGrpcClient_GetReconnectCount 测试重连次数获取
func TestNetGrpcClient_GetReconnectCount(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:            log,
		cnf:            createTestClientConfig("localhost:50051"),
		reconnectCount: 5,
	}

	count := client.GetReconnectCount()
	if count != 5 {
		t.Errorf("期望重连次数 = 5, 实际 = %d", count)
	}
}

// TestNetGrpcClient_Close 测试客户端关闭
func TestNetGrpcClient_Close(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:    log,
		cnf:    createTestClientConfig("localhost:50051"),
		isStop: true,
	}

	client.New()

	// 已经停止的客户端，关闭应该立即返回
	err := client.Close()
	if err != nil {
		t.Errorf("关闭已停止的客户端不应返回错误: %v", err)
	}
}

// TestNetGrpcClient_GetConn 测试获取连接
func TestNetGrpcClient_GetConn(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:  log,
		cnf:  createTestClientConfig("localhost:50051"),
		conn: nil,
	}

	conn := client.GetConn()
	if conn != nil {
		t.Error("未连接时应该返回 nil")
	}
}

// TestNetGrpcClient_Invoke_NoConnection 测试无连接时调用
func TestNetGrpcClient_Invoke_NoConnection(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:  log,
		cnf:  createTestClientConfig("localhost:50051"),
		conn: nil,
	}

	ctx := context.Background()
	err := client.Invoke(ctx, "/test.Service/Method", nil, nil)

	if err == nil {
		t.Error("无连接时调用应该返回错误")
	}
}

// TestClientConfig_Defaults 测试客户端配置默认值
func TestClientConfig_Defaults(t *testing.T) {
	config := &ClientConfig{
		Name:             "test",
		Host:             "localhost:50051",
		KeepAliveTime:    10 * time.Second,
		KeepAliveTimeout: 3 * time.Second,
		ReconnectEnabled: true,
		ReconnectDelay:   5 * time.Second,
		MaxReconnect:     5,
	}

	if config.KeepAliveTime != 10*time.Second {
		t.Errorf("期望 KeepAliveTime = 10s, 实际 = %v", config.KeepAliveTime)
	}
	if !config.ReconnectEnabled {
		t.Error("期望 ReconnectEnabled = true")
	}
	if config.ReconnectDelay != 5*time.Second {
		t.Errorf("期望 ReconnectDelay = 5s, 实际 = %v", config.ReconnectDelay)
	}
	if config.MaxReconnect != 5 {
		t.Errorf("期望 MaxReconnect = 5, 实际 = %d", config.MaxReconnect)
	}
}

// TestClientConfig_Callbacks 测试客户端回调配置
func TestClientConfig_Callbacks(t *testing.T) {
	reconnectCalled := false
	disconnectCalled := false

	config := &ClientConfig{
		Name: "test-client",
		Host: "localhost:50051",
		OnReconnect: func(client *NetGrpcClient) {
			reconnectCalled = true
		},
		OnDisconnect: func(client *NetGrpcClient) {
			disconnectCalled = true
		},
	}

	// 测试回调是否被正确设置
	if config.OnReconnect == nil {
		t.Error("OnReconnect 回调应该被设置")
	}

	if config.OnDisconnect == nil {
		t.Error("OnDisconnect 回调应该被设置")
	}

	// 执行回调
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})
	client := &NetGrpcClient{cnf: config, log: log}
	config.OnReconnect(client)
	config.OnDisconnect(client)

	if !reconnectCalled {
		t.Error("OnReconnect 回调应该被调用")
	}

	if !disconnectCalled {
		t.Error("OnDisconnect 回调应该被调用")
	}
}

// TestNetGrpcClient_ReconnectLogic 测试重连逻辑
func TestNetGrpcClient_ReconnectLogic(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	reconnectCalled := 0
	disconnectCalled := 0

	config := createTestClientConfig("localhost:9999")
	config.MaxReconnect = 2
	config.ReconnectDelay = 100 * time.Millisecond
	config.OnReconnect = func(client *NetGrpcClient) {
		reconnectCalled++
	}
	config.OnDisconnect = func(client *NetGrpcClient) {
		disconnectCalled++
	}

	client := &NetGrpcClient{
		log: log,
		cnf: config,
	}

	client.New()

	// 测试重连次数递增
	client.reconnectCount = 0
	for i := 0; i < 3; i++ {
		client.reconnectCount++
	}

	if client.GetReconnectCount() != 3 {
		t.Errorf("期望重连次数 = 3, 实际 = %d", client.GetReconnectCount())
	}
}

// TestNetGrpcClient_ConcurrentAccess 性能测试：并发访问
func TestNetGrpcClient_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:            log,
		cnf:            createTestClientConfig("localhost:50051"),
		isStop:         false,
		reconnectCount: 0,
	}

	client.New()

	// 并发读取状态
	concurrency := 100
	iterations := 1000

	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = client.IsConnected()
				_ = client.GetReconnectCount()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalOps := concurrency * iterations * 2
	opsPerSec := float64(totalOps) / elapsed.Seconds()

	t.Logf("并发测试完成:")
	t.Logf("  并发数: %d", concurrency)
	t.Logf("  每个协程迭代: %d", iterations)
	t.Logf("  总操作数: %d", totalOps)
	t.Logf("  耗时: %v", elapsed)
	t.Logf("  吞吐量: %.2f ops/s", opsPerSec)

	if opsPerSec < 10000 {
		t.Logf("警告: 吞吐量较低 (%.2f ops/s)", opsPerSec)
	}
}

// BenchmarkNetGrpcClient_IsConnected 基准测试：连接状态检查
func BenchmarkNetGrpcClient_IsConnected(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:    log,
		cnf:    createTestClientConfig("localhost:50051"),
		isStop: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.IsConnected()
	}
}

// BenchmarkNetGrpcClient_GetReconnectCount 基准测试：获取重连次数
func BenchmarkNetGrpcClient_GetReconnectCount(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:            log,
		cnf:            createTestClientConfig("localhost:50051"),
		reconnectCount: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetReconnectCount()
	}
}

// BenchmarkNetGrpcClient_ConcurrentStatusCheck 基准测试：并发状态检查
func BenchmarkNetGrpcClient_ConcurrentStatusCheck(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetGrpcClient{
		log:            log,
		cnf:            createTestClientConfig("localhost:50051"),
		isStop:         false,
		reconnectCount: 5,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = client.IsConnected()
			_ = client.GetReconnectCount()
		}
	})
}
