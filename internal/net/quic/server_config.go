package quic

import (
	"crypto/tls"
	"time"

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
