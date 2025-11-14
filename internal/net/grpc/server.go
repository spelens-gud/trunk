package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

// GrpcNetServer gRPC服务器
type GrpcNetServer struct {
	cnf           *ServerConfig
	log           logger.ILogger
	server        *grpc.Server
	listener      net.Listener
	stopChan      chan chan struct{}
	connCount     int32
	totalAccepted int64
	totalRejected int64
	services      sync.Map
}

// ServerStats 服务器统计信息
type ServerStats struct {
	CurrentConnections int64
	TotalAccepted      int64
	TotalRejected      int64
}

// New 初始化服务器
func (s *GrpcNetServer) New() {
	s.stopChan = make(chan chan struct{})
	s.services = sync.Map{}

	// 配置gRPC服务器选项
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(s.cnf.MaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    s.cnf.KeepAliveTime,
			Timeout: s.cnf.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             s.cnf.KeepAliveTime,
			PermitWithoutStream: true,
		}),
		grpc.ConnectionTimeout(s.cnf.MaxConnectionAge),
	}

	s.server = grpc.NewServer(opts...)
}

// Start 启动服务器
func (s *GrpcNetServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听失败: %w", err)
	}

	s.listener = listener
	s.log.Infof("gRPC服务器启动成功: %s", addr)

	go s.handleStop()

	// 启动gRPC服务器
	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("服务器启动失败: %w", err)
	}

	return nil
}

// RegisterService 注册服务
func (s *GrpcNetServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
	s.services.Store(desc.ServiceName, impl)
	s.log.Infof("注册gRPC服务: %s", desc.ServiceName)
}

// handleStop 处理停止信号
func (s *GrpcNetServer) handleStop() {
	stopDone := <-s.stopChan

	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		s.listener.Close()
	}

	close(stopDone)
}

// Stop 停止服务器
func (s *GrpcNetServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("gRPC服务器已停止")
}

// GetStats 获取统计信息
func (s *GrpcNetServer) GetStats() ServerStats {
	return ServerStats{
		CurrentConnections: int64(atomic.LoadInt32(&s.connCount)),
		TotalAccepted:      atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:      atomic.LoadInt64(&s.totalRejected),
	}
}

// GetConnectionCount 获取当前连接数
func (s *GrpcNetServer) GetConnectionCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}

// GetServer 获取原生gRPC服务器实例
func (s *GrpcNetServer) GetServer() *grpc.Server {
	return s.server
}
