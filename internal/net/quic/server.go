package quic

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/quic-go/quic-go"
	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// NetQuicServer QUIC服务器
type NetQuicServer struct {
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
func (s *NetQuicServer) New() {
	s.stopChan = make(chan chan struct{})
	s.nets = sync.Map{}
}

// Start 启动服务器
func (s *NetQuicServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port)

	quicConfig := &quic.Config{
		MaxIdleTimeout:  s.cnf.IdleTimeout,
		KeepAlivePeriod: s.cnf.KeepAlivePeriod,
	}

	listener, err := quic.ListenAddr(addr, s.cnf.TLSConfig, quicConfig)
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
func (s *NetQuicServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			// 区分服务器关闭和其他错误
			if err.Error() == "quic: server closed" {
				s.log.Infof("服务器已关闭，停止接受新连接")
			} else {
				s.log.Errorf("接受连接失败: %v", err)
			}
			return
		}

		if !s.checkConnectionLimit() {
			_ = conn.CloseWithError(1000, "连接数已达上限")
			atomic.AddInt64(&s.totalRejected, 1)
			s.log.Warnf("拒绝新连接: 已达到最大连接数限制")
			continue
		}

		atomic.AddInt32(&s.connCount, 1)
		atomic.AddInt64(&s.totalAccepted, 1)

		go s.handleConnection(conn)
	}
}

// checkConnectionLimit 检查连接数限制
func (s *NetQuicServer) checkConnectionLimit() bool {
	if s.cnf.MaxConnections <= 0 {
		return true
	}
	return atomic.LoadInt32(&s.connCount) < int32(s.cnf.MaxConnections)
}

// handleConnection 处理连接
func (s *NetQuicServer) handleConnection(qconn *quic.Conn) {
	defer func() {
		atomic.AddInt32(&s.connCount, -1)
		_ = qconn.CloseWithError(0, "")
	}()

	if s.cnf.OnConnect != nil {
		s.cnf.OnConnect(nil)
	}

	// 接受流
	for {
		stream, err := qconn.AcceptStream(context.Background())
		if err != nil {
			// 区分客户端正常关闭和异常错误
			if isNormalClose(err) {
				s.log.Debugf("客户端关闭连接: %v", err)
			} else {
				s.log.Errorf("接受流失败: %v", err)
			}
			if s.cnf.OnClose != nil {
				_ = s.cnf.OnClose(nil)
			}
			return
		}

		go s.handleStream(stream)
	}
}

// handleStream 处理流
func (s *NetQuicServer) handleStream(stream *quic.Stream) {
	defer stream.Close()

	for {
		// 读取消息长度（4字节）
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(stream, lenBuf); err != nil {
			return
		}

		// 解析消息长度
		msgLen := uint32(lenBuf[0])<<24 | uint32(lenBuf[1])<<16 | uint32(lenBuf[2])<<8 | uint32(lenBuf[3])
		if msgLen == 0 || msgLen > 1024*1024 { // 最大1MB
			s.log.Errorf("无效的消息长度: %d", msgLen)
			return
		}

		// 读取消息内容
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(stream, data); err != nil {
			s.log.Errorf("读取消息内容失败: %v", err)
			return
		}

		if s.cnf.OnData != nil {
			if err := s.cnf.OnData(nil, data); err != nil {
				s.log.Errorf("数据处理失败: %v", err)
				return
			}
		}
	}
}

// handleStop 处理停止信号
func (s *NetQuicServer) handleStop() {
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
func (s *NetQuicServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("QUIC服务器已停止")
}

// GetStats 获取统计信息
func (s *NetQuicServer) GetStats() ServerStats {
	return ServerStats{
		CurrentConnections: int64(atomic.LoadInt32(&s.connCount)),
		TotalAccepted:      atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:      atomic.LoadInt64(&s.totalRejected),
	}
}

// GetConnectionCount 获取当前连接数
func (s *NetQuicServer) GetConnectionCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}

// isNormalClose 判断是否为正常关闭
func isNormalClose(err error) bool {
	if err == nil {
		return true
	}
	errStr := err.Error()
	// 客户端正常关闭的错误模式
	return errStr == "Application error 0x0 (remote): 客户端关闭" ||
		errStr == "Application error 0x0" ||
		errStr == "NO_ERROR" ||
		errStr == "quic: server closed"
}
