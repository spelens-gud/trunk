package grpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// 测试辅助函数：创建测试服务器配置
func createTestServerConfig(port int) *ServerConfig {
	return &ServerConfig{
		Name:                 "test-server",
		Ip:                   "127.0.0.1",
		Port:                 port,
		MaxConnections:       100,
		MaxConcurrentStreams: 100,
		KeepAliveTime:        10 * time.Second,
		KeepAliveTimeout:     3 * time.Second,
		MaxConnectionIdle:    5 * time.Minute,
		MaxConnectionAge:     30 * time.Minute,
		OnConnect: func(ctx context.Context, peer *peer.Peer) {
			// 连接建立回调
		},
		OnDisconnect: func(ctx context.Context, peer *peer.Peer) {
			// 连接关闭回调
		},
	}
}

// TestServerConfig_GetMethods 测试配置获取方法
func TestServerConfig_GetMethods(t *testing.T) {
	tests := []struct {
		name   string
		config *ServerConfig
		checks func(*testing.T, *ServerConfig)
	}{
		{
			name:   "默认值测试",
			config: &ServerConfig{},
			checks: func(t *testing.T, cfg *ServerConfig) {
				if cfg.GetMaxConnections() != 0 {
					t.Errorf("期望 MaxConnections = 0, 实际 = %d", cfg.GetMaxConnections())
				}
				if cfg.GetMaxConcurrentStreams() != 0 {
					t.Errorf("期望 MaxConcurrentStreams = 0, 实际 = %d", cfg.GetMaxConcurrentStreams())
				}
			},
		},
		{
			name: "自定义值测试",
			config: &ServerConfig{
				MaxConnections:       50,
				MaxConcurrentStreams: 200,
			},
			checks: func(t *testing.T, cfg *ServerConfig) {
				if cfg.GetMaxConnections() != 50 {
					t.Errorf("期望 MaxConnections = 50, 实际 = %d", cfg.GetMaxConnections())
				}
				if cfg.GetMaxConcurrentStreams() != 200 {
					t.Errorf("期望 MaxConcurrentStreams = 200, 实际 = %d", cfg.GetMaxConcurrentStreams())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checks(t, tt.config)
		})
	}
}

// TestNetGrpcServer_New 测试服务器创建
func TestNetGrpcServer_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf: createTestServerConfig(50051),
		log: log,
	}

	server.New()

	if server.stopChan == nil {
		t.Error("stopChan 未初始化")
	}
	if server.server == nil {
		t.Error("gRPC 服务器未初始化")
	}
}

// TestNetGrpcServer_GetServer 测试获取原生服务器实例
func TestNetGrpcServer_GetServer(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf: createTestServerConfig(50052),
		log: log,
	}

	server.New()

	grpcServer := server.GetServer()
	if grpcServer == nil {
		t.Error("应该返回有效的 gRPC 服务器实例")
	}
}

// TestNetGrpcServer_Stats 测试服务器统计信息
func TestNetGrpcServer_Stats(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf:           createTestServerConfig(50053),
		log:           log,
		connCount:     5,
		totalAccepted: 100,
		totalRejected: 10,
	}

	server.New()

	stats := server.GetStats()
	if stats.CurrentConnections != 5 {
		t.Errorf("期望 CurrentConnections = 5, 实际 = %d", stats.CurrentConnections)
	}
	if stats.TotalAccepted != 100 {
		t.Errorf("期望 TotalAccepted = 100, 实际 = %d", stats.TotalAccepted)
	}
	if stats.TotalRejected != 10 {
		t.Errorf("期望 TotalRejected = 10, 实际 = %d", stats.TotalRejected)
	}

	count := server.GetConnectionCount()
	if count != 5 {
		t.Errorf("期望 ConnectionCount = 5, 实际 = %d", count)
	}
}

// TestNetGrpcServer_RegisterService 测试注册服务
func TestNetGrpcServer_RegisterService(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf: createTestServerConfig(50054),
		log: log,
	}

	server.New()

	// 创建一个测试服务描述
	desc := &grpc.ServiceDesc{
		ServiceName: "test.TestService",
		HandlerType: (*interface{})(nil),
		Methods:     []grpc.MethodDesc{},
		Streams:     []grpc.StreamDesc{},
	}

	// 注册服务不应该 panic
	server.RegisterService(desc, nil)

	// 验证服务已注册
	_, ok := server.services.Load("test.TestService")
	if !ok {
		t.Error("服务应该被注册到 services map 中")
	}
}

// TestServerConfig_Callbacks 测试服务器回调配置
func TestServerConfig_Callbacks(t *testing.T) {
	connectCalled := false
	disconnectCalled := false

	config := &ServerConfig{
		Name: "test-server",
		Ip:   "127.0.0.1",
		Port: 50055,
		OnConnect: func(ctx context.Context, peer *peer.Peer) {
			connectCalled = true
		},
		OnDisconnect: func(ctx context.Context, peer *peer.Peer) {
			disconnectCalled = true
		},
	}

	// 测试回调是否被正确设置
	if config.OnConnect == nil {
		t.Error("OnConnect 回调应该被设置")
	}

	if config.OnDisconnect == nil {
		t.Error("OnDisconnect 回调应该被设置")
	}

	// 执行回调
	ctx := context.Background()
	config.OnConnect(ctx, nil)
	config.OnDisconnect(ctx, nil)

	if !connectCalled {
		t.Error("OnConnect 回调应该被调用")
	}

	if !disconnectCalled {
		t.Error("OnDisconnect 回调应该被调用")
	}
}

// TestServerStats 测试服务器统计结构
func TestServerStats(t *testing.T) {
	stats := ServerStats{
		CurrentConnections: 10,
		TotalAccepted:      100,
		TotalRejected:      5,
	}

	if stats.CurrentConnections != 10 {
		t.Errorf("期望当前连接数为 10，实际为 %d", stats.CurrentConnections)
	}

	if stats.TotalAccepted != 100 {
		t.Errorf("期望总接受连接数为 100，实际为 %d", stats.TotalAccepted)
	}

	if stats.TotalRejected != 5 {
		t.Errorf("期望总拒绝连接数为 5，实际为 %d", stats.TotalRejected)
	}
}

// TestNetGrpcServer_ConcurrentAccess 性能测试：并发访问
func TestNetGrpcServer_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf: createTestServerConfig(50059),
		log: log,
	}

	server.New()

	// 并发读取统计信息
	concurrency := 100
	iterations := 1000

	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = server.GetStats()
				_ = server.GetConnectionCount()
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

// BenchmarkServerConfig_GetMethods 基准测试：配置获取方法
func BenchmarkServerConfig_GetMethods(b *testing.B) {
	config := createTestServerConfig(50056)

	b.Run("GetMaxConnections", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = config.GetMaxConnections()
		}
	})

	b.Run("GetMaxConcurrentStreams", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = config.GetMaxConcurrentStreams()
		}
	})
}

// BenchmarkNetGrpcServer_GetStats 基准测试：获取统计信息
func BenchmarkNetGrpcServer_GetStats(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf:           createTestServerConfig(50057),
		log:           log,
		connCount:     10,
		totalAccepted: 1000,
		totalRejected: 50,
	}

	server.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.GetStats()
	}
}

// BenchmarkNetGrpcServer_ConcurrentStats 基准测试：并发获取统计信息
func BenchmarkNetGrpcServer_ConcurrentStats(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetGrpcServer{
		cnf:           createTestServerConfig(50058),
		log:           log,
		connCount:     10,
		totalAccepted: 1000,
		totalRejected: 50,
	}

	server.New()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = server.GetStats()
		}
	})
}
