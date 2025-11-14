package webSocket

import (
	"net/http"
	"sync"
	"time"
)

// Middleware 中间件函数类型
type Middleware func(next http.HandlerFunc) http.HandlerFunc

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []Middleware // 中间件列表
}

// NewMiddlewareChain 创建中间件链
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

// Use 添加中间件
func (mc *MiddlewareChain) Use(middleware Middleware) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, middleware)
	return mc
}

// Apply 应用中间件链
func (mc *MiddlewareChain) Apply(handler http.HandlerFunc) http.HandlerFunc {
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		handler = mc.middlewares[i](handler)
	}
	return handler
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(maxRequests int, window time.Duration) Middleware {
	type client struct {
		count    int       // 请求计数
		lastTime time.Time // 最后请求时间
	}

	var mu sync.Mutex
	clients := make(map[string]*client)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now()

			mu.Lock()
			if c, exists := clients[ip]; exists {
				// 检查时间间隔
				if now.Sub(c.lastTime) > window {
					c.count = 1
					c.lastTime = now
				} else {
					c.count++
					if c.count > maxRequests {
						mu.Unlock()
						http.Error(w, "请求过于频繁", http.StatusTooManyRequests)
						return
					}
				}
			} else {
				clients[ip] = &client{count: 1, lastTime: now}
			}
			mu.Unlock()

			next(w, r)
		}
	}
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(whitelist []string) Middleware {
	whitelistMap := make(map[string]bool)

	for _, ip := range whitelist {
		whitelistMap[ip] = true
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if len(whitelistMap) > 0 && !whitelistMap[r.RemoteAddr] {
				http.Error(w, "访问被拒绝", http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}
}
