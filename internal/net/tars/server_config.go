package tars

import (
	"context"
)

// ServerConfig TARS服务器配置
type ServerConfig struct {
	Name           string
	Ip             string
	Port           int
	Protocol       string // tcp, udp
	MaxConnections int
	OnConnect      func(ctx context.Context)
	OnDisconnect   func(ctx context.Context)
}

// GetMaxConnections 获取最大连接数
func (c *ServerConfig) GetMaxConnections() int {
	return c.MaxConnections
}
