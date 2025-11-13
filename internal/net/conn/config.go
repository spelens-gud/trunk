package conn

import (
	"errors"
	"time"
)

// DefaultWriteTimeOut 默认写超时
const DefaultWriteTimeOut = time.Second * 30

// DefaultReadTimeOut 默认读超时
const DefaultReadTimeOut = time.Minute * 5

// NetConfig 网络连接配置
type NetConfig[T any] struct {
	Id           uint64         // 连接唯一标识
	Name         string         // 服务名称
	Host         string         // 服务地址
	OnWrite      OnWriteFunc[T] // 写数据处理回调(必须)
	OnRead       OnReadFunc[T]  // 读数据处理回调(必须)
	OnClose      OnCloseFunc[T] // 关闭处理回调(可选)
	OnData       OnDataFunc     // 数据处理回调(必须)
	WriteTimeout time.Duration  // 写超时时间(默认 30s)
	ReadTimeout  time.Duration  // 读超时时间(默认 5m)
	IdleTimeOut  time.Duration  // 空闲超时时间(0 表示不检测)
}

// Validate 验证配置有效性
func (c *NetConfig[T]) Validate() error {
	if c.OnWrite == nil {
		return errors.New("OnWrite回调函数未设置")
	}

	if c.OnRead == nil {
		return errors.New("OnRead回调函数没有设置")
	}

	if c.OnData == nil {
		return errors.New("OnData回调函数没有设置")
	}

	return nil
}

// GetWriteTimeout 获取写超时
func (c *NetConfig[T]) GetWriteTimeout() time.Duration {
	if c.WriteTimeout == 0 {
		c.WriteTimeout = DefaultWriteTimeOut
	}
	return c.WriteTimeout
}

// GetReadTimeout 获取读超时
func (c *NetConfig[T]) GetReadTimeout() time.Duration {
	if c.ReadTimeout == 0 {
		c.ReadTimeout = DefaultReadTimeOut
	}
	return c.ReadTimeout
}
