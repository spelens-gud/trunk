package message_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/spelens-gud/trunk/internal/net/message"
)

// UserData 示例用户数据结构
type UserData struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Level    int    `json:"level"`
}

// TestJSONCodec 测试 JSON 编解码器
func TestJSONCodec(t *testing.T) {
	// 创建 JSON 编解码器
	codec := message.NewJSONCodec[UserData]()

	// 创建消息，指定协议号、服务 ID 和消息 ID
	msg := message.NewMessage(codec, 1001, 2001, 3001)

	// 设置消息体
	userData := UserData{
		UserID:   12345,
		Username: "test_user",
		Level:    10,
	}
	msg.SetBody(userData)
	msg.SetSequence(100)

	// 编码消息
	data, err := msg.Encode()
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("编码后的数据为空")
	}

	// 解码消息
	newMsg := message.NewMessage(codec, 0, 0, 0)
	err = newMsg.Decode(data)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证消息头
	header := newMsg.GetHeader()
	if header.ProtocolID != 1001 {
		t.Errorf("协议号不匹配: 期望 1001, 实际 %d", header.ProtocolID)
	}
	if header.ServiceID != 2001 {
		t.Errorf("服务 ID 不匹配: 期望 2001, 实际 %d", header.ServiceID)
	}
	if header.MessageID != 3001 {
		t.Errorf("消息 ID 不匹配: 期望 3001, 实际 %d", header.MessageID)
	}
	if header.Sequence != 100 {
		t.Errorf("序列号不匹配: 期望 100, 实际 %d", header.Sequence)
	}

	// 验证消息体
	body := newMsg.GetBody()
	if body.UserID != userData.UserID {
		t.Errorf("用户 ID 不匹配: 期望 %d, 实际 %d", userData.UserID, body.UserID)
	}
	if body.Username != userData.Username {
		t.Errorf("用户名不匹配: 期望 %s, 实际 %s", userData.Username, body.Username)
	}
	if body.Level != userData.Level {
		t.Errorf("等级不匹配: 期望 %d, 实际 %d", userData.Level, body.Level)
	}
}

// TestRawCodec 测试原始字节编解码器
func TestRawCodec(t *testing.T) {
	// 创建原始字节编解码器
	codec := message.NewRawCodec()

	// 创建消息
	msg := message.NewMessage(codec, 1002, 2002, 3002)

	// 设置原始字节数据
	rawData := []byte("Hello, World!")
	msg.SetBody(rawData)
	msg.SetSequence(200)

	// 编码消息
	data, err := msg.Encode()
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("编码后的数据为空")
	}

	// 解码消息
	newMsg := message.NewMessage(codec, 0, 0, 0)
	err = newMsg.Decode(data)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证消息头
	header := newMsg.GetHeader()
	if header.ProtocolID != 1002 {
		t.Errorf("协议号不匹配: 期望 1002, 实际 %d", header.ProtocolID)
	}
	if header.ServiceID != 2002 {
		t.Errorf("服务 ID 不匹配: 期望 2002, 实际 %d", header.ServiceID)
	}
	if header.MessageID != 3002 {
		t.Errorf("消息 ID 不匹配: 期望 3002, 实际 %d", header.MessageID)
	}
	if header.Sequence != 200 {
		t.Errorf("序列号不匹配: 期望 200, 实际 %d", header.Sequence)
	}

	// 验证消息体
	body := newMsg.GetBody()
	if !bytes.Equal(body, rawData) {
		t.Errorf("消息体不匹配: 期望 %v, 实际 %v", rawData, body)
	}
}

// CustomCodec 自定义编解码器示例
type CustomCodec struct {
	prefix string
}

// NewCustomCodec 创建自定义编解码器
func NewCustomCodec(prefix string) *CustomCodec {
	return &CustomCodec{prefix: prefix}
}

// Encode 自定义编码
func (c *CustomCodec) Encode(msg string) ([]byte, error) {
	return []byte(c.prefix + msg), nil
}

// Decode 自定义解码
func (c *CustomCodec) Decode(data []byte) (string, error) {
	str := string(data)
	if len(str) < len(c.prefix) {
		return "", errors.New("编码数据不足")
	}
	return str[len(c.prefix):], nil
}

// TestCustomCodec 测试自定义编解码器
func TestCustomCodec(t *testing.T) {
	// 创建自定义编解码器
	codec := NewCustomCodec("PREFIX:")

	// 创建消息
	msg := message.NewMessage(codec, 1003, 2003, 3003)

	// 设置消息体
	msg.SetBody("custom message")
	msg.SetSequence(300)

	// 编码消息
	data, err := msg.Encode()
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("编码后的数据为空")
	}

	// 解码消息
	newMsg := message.NewMessage(codec, 0, 0, 0)
	err = newMsg.Decode(data)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证消息头
	header := newMsg.GetHeader()
	if header.ProtocolID != 1003 {
		t.Errorf("协议号不匹配: 期望 1003, 实际 %d", header.ProtocolID)
	}
	if header.ServiceID != 2003 {
		t.Errorf("服务 ID 不匹配: 期望 2003, 实际 %d", header.ServiceID)
	}
	if header.MessageID != 3003 {
		t.Errorf("消息 ID 不匹配: 期望 3003, 实际 %d", header.MessageID)
	}
	if header.Sequence != 300 {
		t.Errorf("序列号不匹配: 期望 300, 实际 %d", header.Sequence)
	}

	// 验证消息体
	body := newMsg.GetBody()
	if body != "custom message" {
		t.Errorf("消息体不匹配: 期望 'custom message', 实际 '%s'", body)
	}
}

// ExampleMessage 演示如何使用泛型消息
func TestExampleMessage(t *testing.T) {
	// 1. 使用 JSON 编解码器
	jsonCodec := message.NewJSONCodec[UserData]()
	jsonMsg := message.NewMessage(jsonCodec, 1001, 2001, 3001)
	jsonMsg.SetBody(UserData{UserID: 123, Username: "alice", Level: 5})

	data, _ := jsonMsg.Encode()
	fmt.Printf("编码后的 JSON 消息: %d 字节\n", len(data))

	// 2. 使用原始字节编解码器
	rawCodec := message.NewRawCodec()
	rawMsg := message.NewMessage(rawCodec, 1002, 2002, 3002)
	rawMsg.SetBody([]byte("raw data"))

	data, _ = rawMsg.Encode()
	fmt.Printf("编码后的原始消息: %d 字节\n", len(data))

	// 3. 使用自定义编解码器
	customCodec := NewCustomCodec("CUSTOM:")
	customMsg := message.NewMessage(customCodec, 1003, 2003, 3003)
	customMsg.SetBody("hello")

	data, _ = customMsg.Encode()
	fmt.Printf("编码后的自定义消息: %d 字节\n", len(data))
}
