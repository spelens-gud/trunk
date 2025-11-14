package grpc

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/spelens-gud/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NetGrpcServer gRPC服务器
type NetGrpcServer struct {
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
func (s *NetGrpcServer) New() {
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
	}

	// 只有在配置了 MaxConnectionAge 时才添加 ConnectionTimeout
	if s.cnf.MaxConnectionAge > 0 {
		opts = append(opts, grpc.ConnectionTimeout(s.cnf.MaxConnectionAge))
	}

	s.server = grpc.NewServer(opts...)
}

// Start 启动服务器
func (s *NetGrpcServer) Start() error {
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
func (s *NetGrpcServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
	s.services.Store(desc.ServiceName, impl)
	s.log.Infof("注册gRPC服务: %s", desc.ServiceName)
}

// handleStop 处理停止信号
func (s *NetGrpcServer) handleStop() {
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
func (s *NetGrpcServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("gRPC服务器已停止")
}

// GetStats 获取统计信息
func (s *NetGrpcServer) GetStats() ServerStats {
	return ServerStats{
		CurrentConnections: int64(atomic.LoadInt32(&s.connCount)),
		TotalAccepted:      atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:      atomic.LoadInt64(&s.totalRejected),
	}
}

// GetConnectionCount 获取当前连接数
func (s *NetGrpcServer) GetConnectionCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}

// GetServer 获取原生gRPC服务器实例
func (s *NetGrpcServer) GetServer() *grpc.Server {
	return s.server
}
