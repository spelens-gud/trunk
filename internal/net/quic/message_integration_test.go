package quic

import (
	"crypto/tls"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/message"
)

// UserRequest 用户请求消息
type UserRequest struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Action   string `json:"action"`
}

// UserResponse 用户响应消息
type UserResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// TestIntegration_MessageJSON 测试使用 JSON 编解码的消息通信
func TestIntegration_MessageJSON(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18450
	var receivedCount int32

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "message-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	// 创建 JSON 编解码器
	codec := message.NewJSONCodec[UserRequest]()

	// 创建 message 服务器包装器
	msgServer := NewMessageServer(server, codec, func(msg *message.Message[UserRequest]) error {
		atomic.AddInt32(&receivedCount, 1)
		header := msg.GetHeader()
		body := msg.GetBody()

		t.Logf("服务器收到消息 - 协议:%d, 服务:%d, 消息ID:%d, 序列:%d",
			header.ProtocolID, header.ServiceID, header.MessageID, header.Sequence)
		t.Logf("消息内容 - UserID:%d, Username:%s, Action:%s",
			body.UserID, body.Username, body.Action)

		return nil
	})

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer msgServer.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "message-client",
		Host: fmt.Sprintf("127.0.0.1:%d", port),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-trunk"},
		},
	}

	client := &NetQuicClient{
		cnf: clientConfig,
		log: log,
	}

	// 创建 message 客户端包装器
	msgClient := NewMessageClient(
		client,
		codec,
		1001, // ProtocolID
		2001, // ServiceID
		3001, // MessageID
		func(msg *message.Message[UserRequest]) error {
			t.Logf("客户端收到响应消息")
			return nil
		},
	)

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer msgClient.Close()

	time.Sleep(100 * time.Millisecond)

	// 发送消息
	request := UserRequest{
		UserID:   12345,
		Username: "alice",
		Action:   "login",
	}

	if err := msgClient.SendMessage(request, 1); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// 验证
	count := atomic.LoadInt32(&receivedCount)
	if count != 1 {
		t.Errorf("期望接收 1 条消息, 实际接收 %d 条", count)
	}
}

// TestIntegration_MessageRaw 测试使用原始字节编解码的消息通信
func TestIntegration_MessageRaw(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18451
	var receivedData []byte

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "raw-message-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	// 创建原始字节编解码器
	codec := message.NewRawCodec()

	// 创建 message 服务器包装器
	msgServer := NewMessageServer(server, codec, func(msg *message.Message[[]byte]) error {
		header := msg.GetHeader()
		body := msg.GetBody()
		receivedData = body

		t.Logf("服务器收到原始消息 - 协议:%d, 服务:%d, 消息ID:%d, 序列:%d",
			header.ProtocolID, header.ServiceID, header.MessageID, header.Sequence)
		t.Logf("消息内容: %s", string(body))

		return nil
	})

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer msgServer.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "raw-message-client",
		Host: fmt.Sprintf("127.0.0.1:%d", port),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-trunk"},
		},
	}

	client := &NetQuicClient{
		cnf: clientConfig,
		log: log,
	}

	// 创建 message 客户端包装器
	msgClient := NewMessageClient(
		client,
		codec,
		1002, // ProtocolID
		2002, // ServiceID
		3002, // MessageID
		nil,
	)

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer msgClient.Close()

	time.Sleep(100 * time.Millisecond)

	// 发送原始字节消息
	rawData := []byte("Hello QUIC with Message Module")
	if err := msgClient.SendMessage(rawData, 100); err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// 验证
	if string(receivedData) != string(rawData) {
		t.Errorf("期望接收 %s, 实际接收 %s", rawData, receivedData)
	}
}

// TestIntegration_MessageMultiple 测试多条消息通信
func TestIntegration_MessageMultiple(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18452
	var receivedCount int32

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "multi-message-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	codec := message.NewJSONCodec[UserRequest]()

	msgServer := NewMessageServer(server, codec, func(msg *message.Message[UserRequest]) error {
		atomic.AddInt32(&receivedCount, 1)
		body := msg.GetBody()
		t.Logf("收到消息 #%d - UserID:%d, Action:%s",
			msg.GetHeader().Sequence, body.UserID, body.Action)
		return nil
	})

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer msgServer.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "multi-message-client",
		Host: fmt.Sprintf("127.0.0.1:%d", port),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-trunk"},
		},
	}

	client := &NetQuicClient{
		cnf: clientConfig,
		log: log,
	}

	msgClient := NewMessageClient(
		client,
		codec,
		1003,
		2003,
		3003,
		nil,
	)

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer msgClient.Close()

	time.Sleep(100 * time.Millisecond)

	// 发送多条消息
	messageCount := 10
	for i := 0; i < messageCount; i++ {
		request := UserRequest{
			UserID:   uint64(1000 + i),
			Username: fmt.Sprintf("user_%d", i),
			Action:   fmt.Sprintf("action_%d", i),
		}

		if err := msgClient.SendMessage(request, uint64(i+1)); err != nil {
			t.Errorf("发送消息 %d 失败: %v", i, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	// 验证
	count := atomic.LoadInt32(&receivedCount)
	if count != int32(messageCount) {
		t.Errorf("期望接收 %d 条消息, 实际接收 %d 条", messageCount, count)
	}
}
