package conn

import (
	"time"
)

// DefaultWriteTimeOut 默认写超时
const DefaultWriteTimeOut = time.Second * 30

// DefaultReadTimeOut 默认读超时
const DefaultReadTimeOut = time.Minute * 5

// OnConnectFunc 连接建立时调用
type OnConnectFunc[T any] func(conn T)

// OnWriteFunc 写数据处理
type OnWriteFunc[T any] func(conn T, raw []byte) error

// OnReadFunc 读数据处理
type OnReadFunc[T any] func(conn T) (int, []byte, error)

// OnDataFunc 数据处理
type OnDataFunc func(conn IConn, raw []byte) error

// OnCloseFunc 关闭处理
type OnCloseFunc[T any] func(conn T) error

// IConn 连接接口
type IConn interface {
	// Start 启动
	Start()
	// Write 写数据
	Write(b []byte)
	// Close 关闭
	Close() error
	// SetId 设置id
	SetId(id uint64)
	// GetId 获取id
	GetId() uint64
}

// NetConfig 配置
type NetConfig[T any] struct {
	Id           uint64         //  id
	Name         string         // 服务名称
	Host         string         // 服务地址
	OnWrite      OnWriteFunc[T] // 写数据处理
	OnRead       OnReadFunc[T]  // 读数据处理
	OnClose      OnCloseFunc[T] // 关闭处理
	OnData       OnDataFunc     // 数据处理
	WriteTimeout time.Duration  // 写超时
	ReadTimeout  time.Duration  // 读超时
	IdleTimeOut  time.Duration  // 空闲超时
}
