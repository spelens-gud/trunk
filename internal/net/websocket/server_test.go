package webSocket

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// 测试辅助函数：创建测试服务器配置
func createTestServerConfig(port int) *ServerConfig {
	return &ServerConfig{
		Name:  "test-server",
		Ip:    "127.0.0.1",
		Port:  port,
		Route: "/ws",
		Pprof: false,
		OnConnect: func(c conn.IConn) {
			// 连接建立回调
		},
		OnData: func(c conn.IConn, data []byte) error {
			// 数据处理回调
			c.Write(data) // 回显数据
			return nil
		},
		OnClose: func(c conn.IConn) error {
			// 连接关闭回调
			return nil
		},
		WriteTimeout:    5 * time.Second,
		ReadTimeout:     30 * time.Second,
		IdleTimeOut:     60 * time.Second,
		MaxConnections:  100,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		Compression:     true,
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
				if cfg.GetReadBufferSize() != 4096 {
					t.Errorf("期望 ReadBufferSize = 4096, 实际 = %d", cfg.GetReadBufferSize())
				}
				if cfg.GetWriteBufferSize() != 4096 {
					t.Errorf("期望 WriteBufferSize = 4096, 实际 = %d", cfg.GetWriteBufferSize())
				}
				if cfg.GetWriteTimeout() != 10*time.Second {
					t.Errorf("期望 WriteTimeout = 10s, 实际 = %v", cfg.GetWriteTimeout())
				}
				if cfg.GetReadTimeout() != 60*time.Second {
					t.Errorf("期望 ReadTimeout = 60s, 实际 = %v", cfg.GetReadTimeout())
				}
				if cfg.GetIdleTimeOut() != 120*time.Second {
					t.Errorf("期望 IdleTimeOut = 120s, 实际 = %v", cfg.GetIdleTimeOut())
				}
			},
		},
		{
			name: "自定义值测试",
			config: &ServerConfig{
				ReadBufferSize:  8192,
				WriteBufferSize: 8192,
				WriteTimeout:    20 * time.Second,
				ReadTimeout:     90 * time.Second,
				IdleTimeOut:     180 * time.Second,
				MaxConnections:  50,
			},
			checks: func(t *testing.T, cfg *ServerConfig) {
				if cfg.GetReadBufferSize() != 8192 {
					t.Errorf("期望 ReadBufferSize = 8192, 实际 = %d", cfg.GetReadBufferSize())
				}
				if cfg.GetMaxConnections() != 50 {
					t.Errorf("期望 MaxConnections = 50, 实际 = %d", cfg.GetMaxConnections())
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

// TestWsNetServer_New 测试服务器创建
func TestWsNetServer_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &WsNetServer{
		cnf: createTestServerConfig(18080),
		log: log,
	}

	server.New()

	if server.stopChan == nil {
		t.Error("stopChan 未初始化")
	}
	if server.mux == nil {
		t.Error("mux 未初始化")
	}
	if server.nets == nil {
		t.Error("nets 未初始化")
	}
	if server.httpServer == nil {
		t.Error("httpServer 未初始化")
	}
	if server.listener == nil {
		t.Error("listener 未初始化")
	}
}

// TestWsNetServer_Stats 测试服务器统计信息
func TestWsNetServer_Stats(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &WsNetServer{
		cnf:           createTestServerConfig(18081),
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

// TestWsNetServer_ConnectionLimit 测试连接数限制
func TestWsNetServer_ConnectionLimit(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	config := createTestServerConfig(18082)
	config.MaxConnections = 2

	server := &WsNetServer{
		cnf:       config,
		log:       log,
		connCount: 2,
	}

	server.New()

	// 模拟请求
	req, _ := http.NewRequest("GET", "/ws", nil)
	w := &mockResponseWriter{}

	// 应该被拒绝
	rejected := server.checkConnectionsLimit(w, req)
	if !rejected {
		t.Error("期望连接被拒绝，但实际未被拒绝")
	}

	stats := server.GetStats()
	if stats.TotalRejected != 1 {
		t.Errorf("期望 TotalRejected = 1, 实际 = %d", stats.TotalRejected)
	}
}

// mockResponseWriter 模拟 ResponseWriter
type mockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	m.body = append(m.body, b...)
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

// BenchmarkServerConfig_GetMethods 基准测试：配置获取方法
func BenchmarkServerConfig_GetMethods(b *testing.B) {
	config := createTestServerConfig(18083)

	b.Run("GetReadBufferSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = config.GetReadBufferSize()
		}
	})

	b.Run("GetWriteBufferSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = config.GetWriteBufferSize()
		}
	})

	b.Run("GetWriteTimeout", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = config.GetWriteTimeout()
		}
	})
}

// BenchmarkWsNetServer_GetStats 基准测试：获取统计信息
func BenchmarkWsNetServer_GetStats(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &WsNetServer{
		cnf:           createTestServerConfig(18084),
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

// BenchmarkWsNetServer_ConcurrentStats 基准测试：并发获取统计信息
func BenchmarkWsNetServer_ConcurrentStats(b *testing.B) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &WsNetServer{
		cnf:           createTestServerConfig(18085),
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

// TestWsNetServer_ConcurrentAccess 性能测试：并发访问
func TestWsNetServer_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &WsNetServer{
		cnf: createTestServerConfig(18086),
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

// TestWsNetServer_HandleFunc 测试路由注册
func TestWsNetServer_HandleFunc(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &WsNetServer{
		cnf: createTestServerConfig(18087),
		log: log,
	}

	server.New()

	called := false
	server.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := &mockResponseWriter{}

	server.mux.ServeHTTP(w, req)

	if !called {
		t.Error("路由处理函数未被调用")
	}
}
