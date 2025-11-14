package tars

import (
	"time"
)

// ClientConfig TARS客户端配置
type ClientConfig struct {
	Name             string
	Host             string
	Obj              string // TARS对象名
	ReconnectEnabled bool
	ReconnectDelay   time.Duration
	MaxReconnect     int
	OnReconnect      func(client *NetTarsClient)
	OnDisconnect     func(client *NetTarsClient)
}
