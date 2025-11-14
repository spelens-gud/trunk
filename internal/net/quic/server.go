package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// ServerConfig QUIC服务器配置
type ServerConfig struct {
	Name            string                         // 服务器名称
	Ip              string                         // 监听IP
	Port            int                            // 监听端口
	TLSConfig       *tls.Config                    // TLS配置
	OnConnect       func(conn.IConn)               // 连接建立回调
	OnData          func(conn.IConn, []byte) error // 数据处理回调
	OnClose         func(conn.IConn) error         // 连接关闭回调
	MaxConnections  int                            // 最大连接数
	IdleTimeout     time.Duration                  // 空闲超时
	MaxStreamCount  int64                          // 最大流数量
	KeepAlivePeriod time.Duration                  // 保活周期
}

// GetMaxConnections 获取最大连接数
func (c *ServerConfig) GetMaxConnections() int {
	return c.MaxConnections
}

// GetIdleTimeout 获取空闲超时
func (c *ServerConfig) GetIdleTimeout() time.Duration {
	return c.IdleTimeout
}

// QuicNetServer QUIC服务器
type QuicNetServer struct {
	cnf           *ServerConfig
	log           logger.ILogger
	listener      *quic.Listener
	stopChan      chan chan struct{}
	nets          sync.Map
	connCount     int32
	totalAccepted int64
	totalRejected int64
}

// ServerStats 服务器统计信息
type ServerStats struct {
	CurrentConnections int64
	TotalAccepted      int64
	TotalRejected      int64
}

// New 初始化服务器
func (s *QuicNetServer) New() {
	s.stopChan = make(chan chan struct{})
	s.nets = sync.Map{}
}

// Start 启动服务器
func (s *QuicNetServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("解析地址失败: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("监听UDP失败: %w", err)
	}

	quicConfig := &quic.Config{
		MaxIdleTimeout:     s.cnf.IdleTimeout,
		KeepAlivePeriod:    s.cnf.KeepAlivePeriod,
		MaxIncomingStreams: s.cnf.MaxStreamCount,
	}

	listener, err := quic.Listen(udpConn, s.cnf.TLSConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("创建QUIC监听器失败: %w", err)
	}

	s.listener = listener
	s.log.Infof("QUIC服务器启动成功: %s", addr)

	go s.acceptLoop()
	go s.handleStop()

	return nil
}

// acceptLoop 接受连接循环
func (s *QuicNetServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			s.log.Errorf("接受连接失败: %v", err)
			return
		}

		if !s.checkConnectionLimit() {
			conn.CloseWithError(1000, "连接数已达上限")
			atomic.AddInt64(&s.totalRejected, 1)
			continue
		}

		atomic.AddInt32(&s.connCount, 1)
		atomic.AddInt64(&s.totalAccepted, 1)

		go s.handleConnection(conn)
	}
}

// checkConnectionLimit 检查连接数限制
func (s *QuicNetServer) checkConnectionLimit() bool {
	if s.cnf.MaxConnections <= 0 {
		return true
	}
	return atomic.LoadInt32(&s.connCount) < int32(s.cnf.MaxConnections)
}

// handleConnection 处理连接
func (s *QuicNetServer) handleConnection(qconn quic.Connection) {
	defer func() {
		atomic.AddInt32(&s.connCount, -1)
		qconn.CloseWithError(0, "")
	}()

	// 接受流
	for {
		stream, err := qconn.AcceptStream(context.Background())
		if err != nil {
			s.log.Errorf("接受流失败: %v", err)
			return
		}

		go s.handleStream(stream)
	}
}

// handleStream 处理流
func (s *QuicNetServer) handleStream(stream quic.Stream) {
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			return
		}

		if s.cnf.OnData != nil {
			data := make([]byte, n)
			copy(data, buf[:n])

			if err := s.cnf.OnData(nil, data); err != nil {
				s.log.Errorf("数据处理失败: %v", err)
				return
			}
		}
	}
}

// handleStop 处理停止信号
func (s *QuicNetServer) handleStop() {
	stopDone := <-s.stopChan

	if s.listener != nil {
		s.listener.Close()
	}

	s.nets.Range(func(key, value interface{}) bool {
		if c, ok := value.(conn.IConn); ok {
			c.Close()
		}
		return true
	})

	close(stopDone)
}

// Stop 停止服务器
func (s *QuicNetServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("QUIC服务器已停止")
}

// GetStats 获取统计信息
func (s *QuicNetServer) GetStats() ServerStats {
	return ServerStats{
		CurrentConnections: int64(atomic.LoadInt32(&s.connCount)),
		TotalAccepted:      atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:      atomic.LoadInt64(&s.totalRejected),
	}
}

// GetConnectionCount 获取当前连接数
func (s *QuicNetServer) GetConnectionCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}
