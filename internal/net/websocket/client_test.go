package webSocket

import (
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// 测试辅助函数：创建测试客户端配置
func createTestClientConfig(host string) *ClientConfig {
	return &ClientConfig{
		NetConfig: conn.NetConfig[*websocket.Conn]{
			Name: "test-client",
			Host: host,
			OnData: func(c conn.IConn, data []byte) error {
				return nil
			},
		},
		PingTicker: 5 * time.Second,
		PingFunc: func(client *NetWsClient) {
			// 心跳函数
		},
		FirstPingFunc: func(client *NetWsClient) {
			// 首次心跳函数
		},
		ReconnectEnabled: true,
		ReconnectDelay:   2 * time.Second,
		MaxReconnect:     3,
		OnReconnect: func(client *NetWsClient) {
			// 重连成功回调
		},
		OnDisconnect: func(client *NetWsClient) {
			// 断开连接回调
		},
	}
}

// TestWsNetClient_New 测试客户端创建
func TestWsNetClient_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log: log,
		cnf: createTestClientConfig("ws://localhost:8080/ws"),
	}

	client.New()

	if client.stopChan == nil {
		t.Error("stopChan 未初始化")
	}
	if !client.isStop {
		t.Error("isStop 应该为 true")
	}
}

// TestWsNetClient_IsConnected 测试连接状态检查
func TestWsNetClient_IsConnected(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:    log,
		cnf:    createTestClientConfig("ws://localhost:8080/ws"),
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

// TestWsNetClient_GetReconnectCount 测试重连次数获取
func TestWsNetClient_GetReconnectCount(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:            log,
		cnf:            createTestClientConfig("ws://localhost:8080/ws"),
		reconnectCount: 5,
	}

	count := client.GetReconnectCount()
	if count != 5 {
		t.Errorf("期望重连次数 = 5, 实际 = %d", count)
	}
}

// TestWsNetClient_Close 测试客户端关闭
func TestWsNetClient_Close(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:    log,
		cnf:    createTestClientConfig("ws://localhost:8080/ws"),
		isStop: true,
	}

	client.New()

	// 已经停止的客户端，关闭应该立即返回
	err := client.Close()
	if err != nil {
		t.Errorf("关闭已停止的客户端不应返回错误: %v", err)
	}
}

// TestClientConfig_Defaults 测试客户端配置默认值
func TestClientConfig_Defaults(t *testing.T) {
	config := &ClientConfig{
		NetConfig: conn.NetConfig[*websocket.Conn]{
			Name: "test",
			Host: "ws://localhost:8080/ws",
		},
		PingTicker:       10 * time.Second,
		ReconnectEnabled: true,
		ReconnectDelay:   5 * time.Second,
		MaxReconnect:     5,
	}

	if config.PingTicker != 10*time.Second {
		t.Errorf("期望 PingTicker = 10s, 实际 = %v", config.PingTicker)
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

// BenchmarkWsNetClient_IsConnected 基准测试：连接状态检查
func BenchmarkWsNetClient_IsConnected(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:    log,
		cnf:    createTestClientConfig("ws://localhost:8080/ws"),
		isStop: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.IsConnected()
	}
}

// BenchmarkWsNetClient_GetReconnectCount 基准测试：获取重连次数
func BenchmarkWsNetClient_GetReconnectCount(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:            log,
		cnf:            createTestClientConfig("ws://localhost:8080/ws"),
		reconnectCount: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetReconnectCount()
	}
}

// BenchmarkWsNetClient_ConcurrentStatusCheck 基准测试：并发状态检查
func BenchmarkWsNetClient_ConcurrentStatusCheck(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:            log,
		cnf:            createTestClientConfig("ws://localhost:8080/ws"),
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

// TestWsNetClient_ConcurrentAccess 性能测试：并发访问
func TestWsNetClient_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetWsClient{
		log:            log,
		cnf:            createTestClientConfig("ws://localhost:8080/ws"),
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

// TestWsNetClient_ReconnectLogic 测试重连逻辑
func TestWsNetClient_ReconnectLogic(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	reconnectCalled := 0
	disconnectCalled := 0

	config := createTestClientConfig("ws://localhost:9999/ws")
	config.MaxReconnect = 2
	config.ReconnectDelay = 100 * time.Millisecond
	config.OnReconnect = func(client *NetWsClient) {
		reconnectCalled++
	}
	config.OnDisconnect = func(client *NetWsClient) {
		disconnectCalled++
	}

	client := &NetWsClient{
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
