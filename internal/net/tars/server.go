package tars

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/spelens-gud/logger"
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

// TarsNetServer TARS服务器
type TarsNetServer struct {
	cnf           *ServerConfig
	log           logger.ILogger
	comm          *tars.Communicator
	stopChan      chan chan struct{}
	connCount     int32
	totalAccepted int64
	totalRejected int64
	servants      sync.Map
}

// ServerStats 服务器统计信息
type ServerStats struct {
	CurrentConnections int64
	TotalAccepted      int64
	TotalRejected      int64
}

// New 初始化服务器
func (s *TarsNetServer) New() {
	s.stopChan = make(chan chan struct{})
	s.servants = sync.Map{}
	s.comm = tars.NewCommunicator()
}

// Start 启动服务器
func (s *TarsNetServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port)
	s.log.Infof("TARS服务器启动成功: %s", addr)

	go s.handleStop()
	return nil
}

// AddServant 添加Servant
func (s *TarsNetServer) AddServant(obj string, servant interface{}) {
	s.servants.Store(obj, servant)
	s.log.Infof("注册TARS Servant: %s", obj)
}

// handleStop 处理停止信号
func (s *TarsNetServer) handleStop() {
	stopDone := <-s.stopChan
	close(stopDone)
}

// Stop 停止服务器
func (s *TarsNetServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("TARS服务器已停止")
}

// GetStats 获取统计信息
func (s *TarsNetServer) GetStats() ServerStats {
	return ServerStats{
		CurrentConnections: int64(atomic.LoadInt32(&s.connCount)),
		TotalAccepted:      atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:      atomic.LoadInt64(&s.totalRejected),
	}
}

// GetConnectionCount 获取当前连接数
func (s *TarsNetServer) GetConnectionCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}
