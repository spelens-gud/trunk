package tars

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/spelens-gud/logger"
)

// NetTarsServer TARS服务器
type NetTarsServer struct {
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
func (s *NetTarsServer) New() {
	s.stopChan = make(chan chan struct{})
	s.servants = sync.Map{}
	s.comm = tars.NewCommunicator()
}

// Start 启动服务器
func (s *NetTarsServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port)
	s.log.Infof("TARS服务器启动成功: %s", addr)

	go s.handleStop()
	return nil
}

// AddServant 添加Servant
func (s *NetTarsServer) AddServant(obj string, servant interface{}) {
	s.servants.Store(obj, servant)
	s.log.Infof("注册TARS Servant: %s", obj)
}

// handleStop 处理停止信号
func (s *NetTarsServer) handleStop() {
	stopDone := <-s.stopChan
	close(stopDone)
}

// Stop 停止服务器
func (s *NetTarsServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("TARS服务器已停止")
}

// GetStats 获取统计信息
func (s *NetTarsServer) GetStats() ServerStats {
	return ServerStats{
		CurrentConnections: int64(atomic.LoadInt32(&s.connCount)),
		TotalAccepted:      atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:      atomic.LoadInt64(&s.totalRejected),
	}
}

// GetConnectionCount 获取当前连接数
func (s *NetTarsServer) GetConnectionCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}
