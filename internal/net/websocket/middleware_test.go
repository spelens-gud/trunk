package webSocket

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestRateLimitMiddleware 测试限流中间件
func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		maxRequests  int
		window       time.Duration
		requestCount int
		expectPass   int
		expectFail   int
	}{
		{
			name:         "允许范围内的请求",
			maxRequests:  5,
			window:       time.Second,
			requestCount: 3,
			expectPass:   3,
			expectFail:   0,
		},
		{
			name:         "超出限制的请求",
			maxRequests:  3,
			window:       time.Second,
			requestCount: 5,
			expectPass:   3,
			expectFail:   2,
		},
		{
			name:         "单个请求",
			maxRequests:  1,
			window:       time.Second,
			requestCount: 1,
			expectPass:   1,
			expectFail:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passCount := 0
			failCount := 0

			handler := func(w http.ResponseWriter, r *http.Request) {
				passCount++
				w.WriteHeader(http.StatusOK)
			}

			middleware := RateLimitMiddleware(tt.maxRequests, tt.window)
			wrappedHandler := middleware(handler)

			for i := 0; i < tt.requestCount; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()

				wrappedHandler(w, req)

				if w.Code == http.StatusTooManyRequests {
					failCount++
				}
			}

			if passCount != tt.expectPass {
				t.Errorf("期望通过 %d 个请求, 实际通过 %d 个", tt.expectPass, passCount)
			}
			if failCount != tt.expectFail {
				t.Errorf("期望拒绝 %d 个请求, 实际拒绝 %d 个", tt.expectFail, failCount)
			}
		})
	}
}

// TestRateLimitMiddleware_TimeWindow 测试时间窗口重置
func TestRateLimitMiddleware_TimeWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过时间窗口测试")
	}

	maxRequests := 2
	window := 200 * time.Millisecond

	passCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		passCount++
		w.WriteHeader(http.StatusOK)
	}

	middleware := RateLimitMiddleware(maxRequests, window)
	wrappedHandler := middleware(handler)

	// 第一批请求
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		wrappedHandler(w, req)
	}

	if passCount != 2 {
		t.Errorf("第一批: 期望通过 2 个请求, 实际通过 %d 个", passCount)
	}

	// 第三个请求应该被拒绝
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	wrappedHandler(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Error("第三个请求应该被拒绝")
	}

	// 等待时间窗口过期
	time.Sleep(window + 50*time.Millisecond)

	// 时间窗口重置后，应该可以再次请求
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	wrappedHandler(w, req)

	if w.Code != http.StatusOK {
		t.Error("时间窗口重置后的请求应该被允许")
	}
}

// TestRateLimitMiddleware_MultipleClients 测试多客户端限流
func TestRateLimitMiddleware_MultipleClients(t *testing.T) {
	maxRequests := 3
	window := time.Second

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	middleware := RateLimitMiddleware(maxRequests, window)
	wrappedHandler := middleware(handler)

	clients := []string{
		"192.168.1.1:12345",
		"192.168.1.2:12345",
		"192.168.1.3:12345",
	}

	for _, client := range clients {
		for i := 0; i < maxRequests; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = client
			w := httptest.NewRecorder()
			wrappedHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("客户端 %s 的第 %d 个请求应该被允许", client, i+1)
			}
		}

		// 超出限制的请求
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = client
		w := httptest.NewRecorder()
		wrappedHandler(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("客户端 %s 的超限请求应该被拒绝", client)
		}
	}
}

// TestIPWhitelistMiddleware 测试IP白名单中间件
func TestIPWhitelistMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		whitelist  []string
		clientIP   string
		expectCode int
	}{
		{
			name:       "白名单中的IP",
			whitelist:  []string{"127.0.0.1:12345", "192.168.1.1:12345"},
			clientIP:   "127.0.0.1:12345",
			expectCode: http.StatusOK,
		},
		{
			name:       "不在白名单中的IP",
			whitelist:  []string{"127.0.0.1:12345", "192.168.1.1:12345"},
			clientIP:   "10.0.0.1:12345",
			expectCode: http.StatusForbidden,
		},
		{
			name:       "空白名单允许所有",
			whitelist:  []string{},
			clientIP:   "10.0.0.1:12345",
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}

			middleware := IPWhitelistMiddleware(tt.whitelist)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.clientIP
			w := httptest.NewRecorder()

			wrappedHandler(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

// TestMiddlewareChain 测试中间件链
func TestMiddlewareChain(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	// 组合中间件
	rateLimitMW := RateLimitMiddleware(5, time.Second)
	whitelistMW := IPWhitelistMiddleware([]string{"127.0.0.1:12345"})

	// 应用中间件链
	wrappedHandler := whitelistMW(rateLimitMW(handler))

	// 测试白名单IP + 限流
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		wrappedHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("第 %d 个请求应该成功", i+1)
		}
	}

	// 超出限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	wrappedHandler(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Error("超限请求应该被拒绝")
	}

	// 测试非白名单IP
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w = httptest.NewRecorder()
	wrappedHandler(w, req)

	if w.Code != http.StatusForbidden {
		t.Error("非白名单IP应该被拒绝")
	}
}

// BenchmarkRateLimitMiddleware 基准测试：限流中间件
func BenchmarkRateLimitMiddleware(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	middleware := RateLimitMiddleware(1000, time.Second)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler(w, req)
	}
}

// BenchmarkIPWhitelistMiddleware 基准测试：IP白名单中间件
func BenchmarkIPWhitelistMiddleware(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	whitelist := []string{"127.0.0.1:12345", "192.168.1.1:12345", "10.0.0.1:12345"}
	middleware := IPWhitelistMiddleware(whitelist)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler(w, req)
	}
}

// BenchmarkMiddlewareChain 基准测试：中间件链
func BenchmarkMiddlewareChain(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	rateLimitMW := RateLimitMiddleware(1000, time.Second)
	whitelistMW := IPWhitelistMiddleware([]string{"127.0.0.1:12345"})
	wrappedHandler := whitelistMW(rateLimitMW(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler(w, req)
	}
}

// TestRateLimitMiddleware_ConcurrentRequests 性能测试：并发请求限流
func TestRateLimitMiddleware_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	maxRequests := 100
	window := time.Second

	var passCount, failCount int
	var mu sync.Mutex

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		passCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}

	middleware := RateLimitMiddleware(maxRequests, window)
	wrappedHandler := middleware(handler)

	concurrency := 200
	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()
			wrappedHandler(w, req)

			if w.Code == http.StatusTooManyRequests {
				mu.Lock()
				failCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("并发限流测试完成:")
	t.Logf("  并发数: %d", concurrency)
	t.Logf("  限流阈值: %d", maxRequests)
	t.Logf("  通过请求: %d", passCount)
	t.Logf("  拒绝请求: %d", failCount)
	t.Logf("  耗时: %v", elapsed)

	if passCount > maxRequests {
		t.Errorf("通过的请求数 (%d) 超过了限流阈值 (%d)", passCount, maxRequests)
	}
}
