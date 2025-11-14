package message

import (
	"errors"
)

// Codec 编解码器接口
type Codec[T any] interface {
	// Encode 编码消息
	Encode(msg T) ([]byte, error)
	// Decode 解码消息
	Decode(data []byte) (T, error)
}

// Header 消息头信息
type Header struct {
	// ProtocolID 协议号
	ProtocolID uint32
	// ServiceID 分布式服务 ID
	ServiceID uint32
	// MessageID 消息 ID
	MessageID uint32
	// Sequence 序列号
	Sequence uint64
}

// Message 泛型消息结构
type Message[T any] struct {
	// Header 消息头
	Header Header
	// Body 消息体
	Body T
	// codec 编解码器
	codec Codec[T]
}

// NewMessage 创建新消息
func NewMessage[T any](codec Codec[T], protocolID, serviceID, messageID uint32) *Message[T] {
	return &Message[T]{
		Header: Header{
			ProtocolID: protocolID,
			ServiceID:  serviceID,
			MessageID:  messageID,
		},
		codec: codec,
	}
}

// SetBody 设置消息体
func (m *Message[T]) SetBody(body T) {
	m.Body = body
}

// GetBody 获取消息体
func (m *Message[T]) GetBody() T {
	return m.Body
}

// SetSequence 设置序列号
func (m *Message[T]) SetSequence(seq uint64) {
	m.Header.Sequence = seq
}

// GetHeader 获取消息头
func (m *Message[T]) GetHeader() Header {
	return m.Header
}

// Encode 编码消息（包含消息头和消息体）
func (m *Message[T]) Encode() ([]byte, error) {
	if m.codec == nil {
		return nil, errors.New("编码器不存在")
	}

	// 编码消息体
	bodyData, err := m.codec.Encode(m.Body)
	if err != nil {
		return nil, err
	}

	// 编码消息头(简单的二进制格式)
	headerData := make([]byte, 20) // 4+4+4+8 字节
	putUint32(headerData[0:4], m.Header.ProtocolID)
	putUint32(headerData[4:8], m.Header.ServiceID)
	putUint32(headerData[8:12], m.Header.MessageID)
	putUint64(headerData[12:20], m.Header.Sequence)

	// 组合消息头和消息体
	result := make([]byte, len(headerData)+len(bodyData))
	copy(result, headerData)
	copy(result[len(headerData):], bodyData)

	return result, nil
}

// Decode 解码消息（包含消息头和消息体）
func (m *Message[T]) Decode(data []byte) error {
	if m.codec == nil {
		return errors.New("编码器不存在")
	}

	if len(data) < 20 {
		return errors.New("编码数据数据小于编码限定值")
	}

	// 解码消息头
	m.Header.ProtocolID = getUint32(data[0:4])
	m.Header.ServiceID = getUint32(data[4:8])
	m.Header.MessageID = getUint32(data[8:12])
	m.Header.Sequence = getUint64(data[12:20])

	// 解码消息体
	body, err := m.codec.Decode(data[20:])
	if err != nil {
		return err
	}

	m.Body = body
	return nil
}

// putUint32 将 uint32 写入字节数组(大端序)
func putUint32(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

// getUint32 从字节数组读取 uint32(大端序)
func getUint32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

// putUint64 将 uint64 写入字节数组(大端序)
func putUint64(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

// getUint64 从字节数组读取 uint64(大端序)
func getUint64(b []byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}
