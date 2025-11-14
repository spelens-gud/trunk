package grpc

import (
	"time"
)

// ClientConfig gRPC客户端配置
type ClientConfig struct {
	Name             string                      // 客户端名称
	Host             string                      // 服务地址
	KeepAliveTime    time.Duration               // keepalive时间间隔
	KeepAliveTimeout time.Duration               // keepalive超时时间
	ReconnectEnabled bool                        // 是否启用重连
	ReconnectDelay   time.Duration               // 重连间隔
	MaxReconnect     int                         // 最大重连次数
	OnReconnect      func(client *NetGrpcClient) // 重连成功回调
	OnDisconnect     func(client *NetGrpcClient) // 断开连接回调
}
