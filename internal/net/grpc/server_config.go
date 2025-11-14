package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/peer"
)

// ServerConfig gRPC服务器配置
type ServerConfig struct {
	Name                 string
	Ip                   string
	Port                 int
	MaxConnections       int
	MaxConcurrentStreams uint32
	KeepAliveTime        time.Duration
	KeepAliveTimeout     time.Duration
	MaxConnectionIdle    time.Duration
	MaxConnectionAge     time.Duration
	OnConnect            func(ctx context.Context, peer *peer.Peer)
	OnDisconnect         func(ctx context.Context, peer *peer.Peer)
}

// GetMaxConnections 获取最大连接数
func (c *ServerConfig) GetMaxConnections() int {
	return c.MaxConnections
}

// GetMaxConcurrentStreams 获取最大并发流数
func (c *ServerConfig) GetMaxConcurrentStreams() uint32 {
	return c.MaxConcurrentStreams
}
