package quic

import (
	"fmt"

	"github.com/spelens-gud/trunk/internal/net/conn"
	"github.com/spelens-gud/trunk/internal/net/message"
)

// MessageClient 支持 message 模块的 QUIC 客户端包装器
type MessageClient[T any] struct {
	client  *NetQuicClient
	message *message.Message[T]
	onMsg   func(*message.Message[T]) error
}

// NewMessageClient 创建支持 message 的 QUIC 客户端
func NewMessageClient[T any](
	client *NetQuicClient,
	codec message.Codec[T],
	protocolID, serviceID, messageID uint32,
	onMsg func(*message.Message[T]) error,
) *MessageClient[T] {
	mc := &MessageClient[T]{
		client:  client,
		message: message.NewMessage(codec, protocolID, serviceID, messageID),
		onMsg:   onMsg,
	}

	// 设置原始数据回调
	client.cnf.OnData = func(c *NetQuicClient, data []byte) error {
		// 解码消息
		msg := message.NewMessage(codec, 0, 0, 0)
		if err := msg.Decode(data); err != nil {
			return fmt.Errorf("解码消息失败: %w", err)
		}

		// 调用用户回调
		if mc.onMsg != nil {
			return mc.onMsg(msg)
		}
		return nil
	}

	return mc
}

// SendMessage 发送消息
func (mc *MessageClient[T]) SendMessage(body T, sequence uint64) error {
	mc.message.SetBody(body)
	mc.message.SetSequence(sequence)

	data, err := mc.message.Encode()
	if err != nil {
		return fmt.Errorf("编码消息失败: %w", err)
	}

	return mc.client.Write(data)
}

// GetClient 获取底层 QUIC 客户端
func (mc *MessageClient[T]) GetClient() *NetQuicClient {
	return mc.client
}

// Close 关闭客户端
func (mc *MessageClient[T]) Close() error {
	return mc.client.Close()
}

// MessageServer 支持 message 模块的 QUIC 服务器包装器
type MessageServer[T any] struct {
	server *NetQuicServer
	codec  message.Codec[T]
	onMsg  func(*message.Message[T]) error
}

// NewMessageServer 创建支持 message 的 QUIC 服务器
func NewMessageServer[T any](
	server *NetQuicServer,
	codec message.Codec[T],
	onMsg func(*message.Message[T]) error,
) *MessageServer[T] {
	ms := &MessageServer[T]{
		server: server,
		codec:  codec,
		onMsg:  onMsg,
	}

	// 设置原始数据回调
	server.cnf.OnData = func(c conn.IConn, data []byte) error {
		// 解码消息
		msg := message.NewMessage(codec, 0, 0, 0)
		if err := msg.Decode(data); err != nil {
			return fmt.Errorf("解码消息失败: %w", err)
		}

		// 调用用户回调
		if ms.onMsg != nil {
			return ms.onMsg(msg)
		}
		return nil
	}

	return ms
}

// GetServer 获取底层 QUIC 服务器
func (ms *MessageServer[T]) GetServer() *NetQuicServer {
	return ms.server
}

// Stop 停止服务器
func (ms *MessageServer[T]) Stop() {
	ms.server.Stop()
}
