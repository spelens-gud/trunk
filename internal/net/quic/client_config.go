package quic

import (
	"crypto/tls"
	"time"
)

// ClientConfig QUIC客户端配置
type ClientConfig struct {
	Name             string                                         // 客户端名称
	Host             string                                         // 服务器地址
	TLSConfig        *tls.Config                                    // TLS配置
	PingTicker       time.Duration                                  // 心跳间隔
	PingFunc         func(client *NetQuicClient)                    // 心跳函数
	FirstPingFunc    func(client *NetQuicClient)                    // 首次连接心跳函数
	ReconnectEnabled bool                                           // 是否启用重连
	ReconnectDelay   time.Duration                                  // 重连延迟
	MaxReconnect     int                                            // 最大重连次数
	OnReconnect      func(client *NetQuicClient)                    // 重连成功回调
	OnDisconnect     func(client *NetQuicClient)                    // 断开连接回调
	OnData           func(client *NetQuicClient, data []byte) error // 数据处理回调
}
