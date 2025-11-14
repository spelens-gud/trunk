package message

//go:generate protoc --go_out=. demo.proto

import (
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/proto"
)

// ProtobufCodec Protobuf 编解码器
type ProtobufCodec[T proto.Message] struct{}

// NewProtobufCodec 创建 Protobuf 编解码器
func NewProtobufCodec[T proto.Message]() *ProtobufCodec[T] {
	return &ProtobufCodec[T]{}
}

// Encode 编码 Protobuf 消息
func (c *ProtobufCodec[T]) Encode(msg T) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, errors.New("错误编码失败")
	}
	return data, nil
}

// Decode 解码 Protobuf 消息
func (c *ProtobufCodec[T]) Decode(data []byte) (T, error) {
	var msg T
	// 创建消息实例
	msg = msg.ProtoReflect().New().Interface().(T)

	if err := proto.Unmarshal(data, msg); err != nil {
		return msg, errors.New("错误解码失败")
	}
	return msg, nil
}

// JSONCodec JSON 编解码器
type JSONCodec[T any] struct{}

// NewJSONCodec 创建 JSON 编解码器
func NewJSONCodec[T any]() *JSONCodec[T] {
	return &JSONCodec[T]{}
}

// Encode 编码 JSON 消息
func (c *JSONCodec[T]) Encode(msg T) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.New("错误编码失败")
	}
	return data, nil
}

// Decode 解码 JSON 消息
func (c *JSONCodec[T]) Decode(data []byte) (T, error) {
	var msg T
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, errors.New("错误解码失败")
	}
	return msg, nil
}

// RawCodec 原始字节编解码器(不做任何转换)
type RawCodec struct{}

// NewRawCodec 创建原始字节编解码器
func NewRawCodec() *RawCodec {
	return &RawCodec{}
}

// Encode 编码原始字节
func (c *RawCodec) Encode(msg []byte) ([]byte, error) {
	return msg, nil
}

// Decode 解码原始字节
func (c *RawCodec) Decode(data []byte) ([]byte, error) {
	return data, nil
}
