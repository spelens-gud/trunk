package webSocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	conn.NetConfig[*websocket.Conn]                    // 基础配置
	PingTicker                      time.Duration      // 心跳间隔
	PingFunc                        func(*NetWsClient) // 心跳函数
	FirstPingFunc                   func(*NetWsClient) // 首次心跳函数
	ReconnectEnabled                bool               // 是否启用自动重连
	ReconnectDelay                  time.Duration      // 重连延迟
	MaxReconnect                    int                // 最大重连次数，0表示无限重连
	OnReconnect                     func(*NetWsClient) // 重连成功回调
	OnDisconnect                    func(*NetWsClient) // 断开连接回调
}
