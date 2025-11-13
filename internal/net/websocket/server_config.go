package webSocket

import (
	"time"

	"github.com/spelens-gud/trunk/internal/net/conn"
)

type ServerConfig struct {
	Name            string                         // 服务名称
	Ip              string                         // 服务ip
	Port            int                            // 服务端口
	Route           string                         // 路由
	Pprof           bool                           // 是否开启pprof
	OnConnect       func(conn.IConn)               // 连接建立时调用
	OnData          func(conn.IConn, []byte) error // 数据处理
	OnClose         func(conn.IConn) error         // 连接关闭时调用
	WriteTimeout    time.Duration                  // 写超时
	ReadTimeout     time.Duration                  // 读超时
	IdleTimeOut     time.Duration                  // 空闲超时
	MaxConnections  int                            // 最大连接数限制，0表示不限制
	ReadBufferSize  int                            // 读缓冲区大小，默认4096
	WriteBufferSize int                            // 写缓冲区大小，默认4096
	Compression     bool                           // 是否启用压缩，默认true
}

// GetMaxConnections 获取最大连接数限制
func (s *ServerConfig) GetMaxConnections() int {
	return s.MaxConnections
}

// GetReadBufferSize 获取读缓冲区大小
func (s *ServerConfig) GetReadBufferSize() int {
	if s.ReadBufferSize == 0 {
		return 4096
	}
	return s.ReadBufferSize
}

// GetWriteBufferSize 获取读缓冲区大小
func (s *ServerConfig) GetWriteBufferSize() int {
	if s.WriteBufferSize == 0 {
		return 4096
	}
	return s.WriteBufferSize
}

// GetWriteTimeout 获取写超时
func (s *ServerConfig) GetWriteTimeout() time.Duration {
	if s.WriteTimeout <= 0 {
		return 10 * time.Second
	}
	return s.WriteTimeout
}

// GetReadTimeout 获取读超时
func (s *ServerConfig) GetReadTimeout() time.Duration {
	if s.ReadTimeout <= 0 {
		return 60 * time.Second
	}
	return s.ReadTimeout
}

// GetIdleTimeOut 获取空闲超时
func (s *ServerConfig) GetIdleTimeOut() time.Duration {
	if s.IdleTimeOut <= 0 {
		return 120 * time.Second
	}
	return s.IdleTimeOut
}
