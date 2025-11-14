package grpc

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"context"

	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative test.proto

// waitForPort 等待端口可用
func waitForPort(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// TestServiceImpl 测试服务实现（使用 message 模块）
type TestServiceImpl struct {
	UnimplementedTestServiceServer
	receivedMessages []string
	mu               sync.Mutex
	codec            message.Codec[*EchoRequest]
	respCodec        message.Codec[*EchoResponse]
}

// Echo 实现回显方法（使用 message 模块进行编解码）
func (s *TestServiceImpl) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	// 使用 message 模块包装请求
	reqMsg := message.NewMessage(s.codec, 1001, 2001, 3001)
	reqMsg.SetBody(req)

	// 编码请求（模拟消息传输）
	encodedReq, err := reqMsg.Encode()
	if err != nil {
		return nil, fmt.Errorf("编码请求失败: %w", err)
	}

	// 解码请求（模拟接收消息）
	decodedReqMsg := message.NewMessage(s.codec, 0, 0, 0)
	if err := decodedReqMsg.Decode(encodedReq); err != nil {
		return nil, fmt.Errorf("解码请求失败: %w", err)
	}

	decodedReq := decodedReqMsg.GetBody()

	s.mu.Lock()
	s.receivedMessages = append(s.receivedMessages, decodedReq.Message)
	s.mu.Unlock()

	// 创建响应消息
	respMsg := message.NewMessage(s.respCodec, 1001, 2001, 3002)
	respMsg.SetBody(&EchoResponse{Message: decodedReq.Message})

	// 编码响应
	encodedResp, err := respMsg.Encode()
	if err != nil {
		return nil, fmt.Errorf("编码响应失败: %w", err)
	}

	// 解码响应（模拟客户端接收）
	decodedRespMsg := message.NewMessage(s.respCodec, 0, 0, 0)
	if err := decodedRespMsg.Decode(encodedResp); err != nil {
		return nil, fmt.Errorf("解码响应失败: %w", err)
	}

	resp := decodedRespMsg.GetBody()
	return resp, nil
}

// GetReceivedMessages 获取接收到的消息
func (s *TestServiceImpl) GetReceivedMessages() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := make([]string, len(s.receivedMessages))
	copy(messages, s.receivedMessages)
	return messages
}

// TestIntegration_ServerClientCommunication 集成测试：服务器与客户端通信
func TestIntegration_ServerClientCommunication(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 60000
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建并启动服务器
	serverConfig := &ServerConfig{
		Name:                 "integration-server",
		Ip:                   "127.0.0.1",
		Port:                 port,
		MaxConnections:       10,
		MaxConcurrentStreams: 100,
		KeepAliveTime:        10 * time.Second,
		KeepAliveTimeout:     3 * time.Second,
		MaxConnectionAge:     0, // 不限制连接时长
	}

	server := &NetGrpcServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()

	// 注册测试服务（使用 Protobuf 编解码器）
	testService := &TestServiceImpl{
		receivedMessages: make([]string, 0),
		codec:            message.NewProtobufCodec[*EchoRequest](),
		respCodec:        message.NewProtobufCodec[*EchoResponse](),
	}
	RegisterTestServiceServer(server.GetServer(), testService)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	// 等待服务器端口可用
	if !waitForPort(port, 5*time.Second) {
		t.Fatal("服务器启动超时")
	}

	// 额外等待确保服务器完全就绪
	time.Sleep(1 * time.Second)

	t.Log("服务器端口已就绪")

	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("客户端连接失败: %v", err)
	}
	defer conn.Close()

	// 等待连接就绪
	time.Sleep(1 * time.Second)

	t.Log("客户端连接成功")

	// 创建客户端
	client := NewTestServiceClient(conn)

	// 发送测试消息
	ctx := context.Background()
	testMessage := "Hello, gRPC with Message Module!"

	req := &EchoRequest{Message: testMessage}
	resp, err := client.Echo(ctx, req)
	if err != nil {
		t.Fatalf("调用 Echo 方法失败: %v", err)
	}

	// 验证响应
	if resp.Message != testMessage {
		t.Errorf("期望收到 '%s', 实际收到 '%s'", testMessage, resp.Message)
	}

	t.Logf("成功收到响应: %s", resp.Message)

	// 验证服务器收到消息
	messages := testService.GetReceivedMessages()
	if len(messages) != 1 {
		t.Errorf("期望服务器收到 1 条消息, 实际收到 %d 条", len(messages))
	}
	if len(messages) > 0 && messages[0] != testMessage {
		t.Errorf("期望服务器收到 '%s', 实际收到 '%s'", testMessage, messages[0])
	}

	// 清理
	server.Stop()
}

// TestIntegration_MultipleClients 集成测试：多客户端连接
func TestIntegration_MultipleClients(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 60001
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建并启动服务器
	serverConfig := &ServerConfig{
		Name:                 "multi-client-server",
		Ip:                   "127.0.0.1",
		Port:                 port,
		MaxConnections:       10,
		MaxConcurrentStreams: 100,
		KeepAliveTime:        10 * time.Second,
		KeepAliveTimeout:     3 * time.Second,
		MaxConnectionAge:     0,
	}

	server := &NetGrpcServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()

	// 注册测试服务
	testService := &TestServiceImpl{
		receivedMessages: make([]string, 0),
		codec:            message.NewProtobufCodec[*EchoRequest](),
		respCodec:        message.NewProtobufCodec[*EchoResponse](),
	}
	RegisterTestServiceServer(server.GetServer(), testService)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	// 创建多个客户端
	clientCount := 5
	var wg sync.WaitGroup
	wg.Add(clientCount)

	for i := range clientCount {
		go func(id int) {
			defer wg.Done()

			// 创建客户端连接
			conn, err := grpc.NewClient(
				fmt.Sprintf("127.0.0.1:%d", port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				t.Errorf("客户端 %d 连接失败: %v", id, err)
				return
			}
			defer conn.Close()

			client := NewTestServiceClient(conn)

			// 发送消息
			ctx := context.Background()
			req := &EchoRequest{Message: fmt.Sprintf("Message from client %d", id)}

			resp, err := client.Echo(ctx, req)
			if err != nil {
				t.Errorf("客户端 %d 调用失败: %v", id, err)
				return
			}

			if resp.Message != req.Message {
				t.Errorf("客户端 %d: 期望 '%s', 实际 '%s'", id, req.Message, resp.Message)
			}

			t.Logf("客户端 %d 成功收到响应: %s", id, resp.Message)
		}(i)
	}

	wg.Wait()

	// 验证服务器收到所有消息
	messages := testService.GetReceivedMessages()
	if len(messages) != clientCount {
		t.Errorf("期望服务器收到 %d 条消息, 实际收到 %d 条", clientCount, len(messages))
	}

	t.Logf("服务器总共收到 %d 条消息", len(messages))

	// 清理
	server.Stop()
}

// TestIntegration_ConcurrentRequests 集成测试：并发请求
func TestIntegration_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 60002
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建并启动服务器
	serverConfig := &ServerConfig{
		Name:                 "concurrent-server",
		Ip:                   "127.0.0.1",
		Port:                 port,
		MaxConnections:       100,
		MaxConcurrentStreams: 100,
		KeepAliveTime:        10 * time.Second,
		KeepAliveTimeout:     3 * time.Second,
		MaxConnectionAge:     0,
	}

	server := &NetGrpcServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()

	// 注册测试服务
	testService := &TestServiceImpl{
		receivedMessages: make([]string, 0),
		codec:            message.NewProtobufCodec[*EchoRequest](),
		respCodec:        message.NewProtobufCodec[*EchoResponse](),
	}
	RegisterTestServiceServer(server.GetServer(), testService)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	// 创建客户端连接
	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("客户端连接失败: %v", err)
	}
	defer conn.Close()

	client := NewTestServiceClient(conn)

	// 并发发送请求
	concurrency := 50
	var wg sync.WaitGroup
	var successCount, failCount int32

	wg.Add(concurrency)
	start := time.Now()

	for i := range concurrency {
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			req := &EchoRequest{Message: fmt.Sprintf("Concurrent message %d", id)}

			resp, err := client.Echo(ctx, req)
			if err != nil {
				atomic.AddInt32(&failCount, 1)
				t.Logf("请求 %d 失败: %v", id, err)
			} else {
				atomic.AddInt32(&successCount, 1)
				if resp.Message != req.Message {
					t.Errorf("请求 %d: 期望 '%s', 实际 '%s'", id, req.Message, resp.Message)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("并发请求测试完成:")
	t.Logf("  并发数: %d", concurrency)
	t.Logf("  成功请求: %d", successCount)
	t.Logf("  失败请求: %d", failCount)
	t.Logf("  耗时: %v", elapsed)
	t.Logf("  平均延迟: %v", elapsed/time.Duration(concurrency))

	// 验证服务器收到的消息数
	messages := testService.GetReceivedMessages()
	if len(messages) != int(successCount) {
		t.Errorf("期望服务器收到 %d 条消息, 实际收到 %d 条", successCount, len(messages))
	}

	// 清理
	server.Stop()
}

// TestIntegration_DataTransfer 集成测试：数据传输
func TestIntegration_DataTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 60003
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建并启动服务器
	serverConfig := &ServerConfig{
		Name:                 "data-transfer-server",
		Ip:                   "127.0.0.1",
		Port:                 port,
		MaxConnections:       10,
		MaxConcurrentStreams: 100,
		KeepAliveTime:        10 * time.Second,
		KeepAliveTimeout:     3 * time.Second,
		MaxConnectionAge:     0,
	}

	server := &NetGrpcServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()

	// 注册测试服务
	testService := &TestServiceImpl{
		receivedMessages: make([]string, 0),
		codec:            message.NewProtobufCodec[*EchoRequest](),
		respCodec:        message.NewProtobufCodec[*EchoResponse](),
	}
	RegisterTestServiceServer(server.GetServer(), testService)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	// 创建客户端连接
	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("客户端连接失败: %v", err)
	}
	defer conn.Close()

	client := NewTestServiceClient(conn)

	// 测试不同大小的数据传输
	testCases := []struct {
		name    string
		message string
	}{
		{"短消息", "Hello"},
		{"中等消息", "This is a medium length message for testing gRPC communication with message module"},
		{"长消息", string(make([]byte, 1024))}, // 1KB 数据
		{"特殊字符", "测试中文 🚀 Special chars: !@#$%^&*()"},
		{"空消息", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			req := &EchoRequest{Message: tc.message}

			resp, err := client.Echo(ctx, req)
			if err != nil {
				t.Fatalf("调用失败: %v", err)
			}

			if resp.Message != tc.message {
				t.Errorf("数据不匹配: 期望长度 %d, 实际长度 %d",
					len(tc.message), len(resp.Message))
			}

			t.Logf("成功传输 %d 字节数据", len(tc.message))
		})
	}

	// 清理
	server.Stop()
}

// TestIntegration_HighThroughput 集成测试：高吞吐量
func TestIntegration_HighThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 60005
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建并启动服务器
	serverConfig := &ServerConfig{
		Name:                 "throughput-server",
		Ip:                   "127.0.0.1",
		Port:                 port,
		MaxConnections:       100,
		MaxConcurrentStreams: 100,
		KeepAliveTime:        10 * time.Second,
		KeepAliveTimeout:     3 * time.Second,
		MaxConnectionAge:     0,
	}

	server := &NetGrpcServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()

	// 注册测试服务
	testService := &TestServiceImpl{
		receivedMessages: make([]string, 0),
		codec:            message.NewProtobufCodec[*EchoRequest](),
		respCodec:        message.NewProtobufCodec[*EchoResponse](),
	}
	RegisterTestServiceServer(server.GetServer(), testService)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	// 创建多个客户端连接
	clientCount := 10
	requestsPerClient := 100

	var wg sync.WaitGroup
	var totalSuccess int32

	start := time.Now()

	for i := range clientCount {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			// 创建客户端连接
			conn, err := grpc.NewClient(
				fmt.Sprintf("127.0.0.1:%d", port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				t.Errorf("客户端 %d 连接失败: %v", clientID, err)
				return
			}
			defer conn.Close()

			client := NewTestServiceClient(conn)

			// 发送多个请求
			for j := range requestsPerClient {
				ctx := context.Background()
				req := &EchoRequest{Message: fmt.Sprintf("Client %d - Request %d", clientID, j)}

				resp, err := client.Echo(ctx, req)
				if err == nil && resp.Message == req.Message {
					atomic.AddInt32(&totalSuccess, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalRequests := clientCount * requestsPerClient
	throughput := float64(totalSuccess) / elapsed.Seconds()

	t.Logf("高吞吐量测试完成:")
	t.Logf("  客户端数: %d", clientCount)
	t.Logf("  每客户端请求数: %d", requestsPerClient)
	t.Logf("  总请求数: %d", totalRequests)
	t.Logf("  成功请求数: %d", totalSuccess)
	t.Logf("  耗时: %v", elapsed)
	t.Logf("  吞吐量: %.2f requests/s", throughput)

	if totalSuccess < int32(totalRequests*9/10) {
		t.Errorf("成功率过低: %d/%d (%.2f%%)",
			totalSuccess, totalRequests, float64(totalSuccess)/float64(totalRequests)*100)
	}

	// 清理
	server.Stop()
}
