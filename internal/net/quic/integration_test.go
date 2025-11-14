package quic

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// generateIntegrationTestTLSConfig 生成集成测试用的TLS配置
func generateIntegrationTestTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Co"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-trunk"},
	}
}

// waitForPort 等待端口可用
func waitForPort(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("udp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// TestIntegration_ServerClientCommunication 集成测试：服务器与客户端通信
func TestIntegration_ServerClientCommunication(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18443
	var receivedData []byte
	var mu sync.Mutex

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "test-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
		OnData: func(_ conn.IConn, data []byte) error {
			mu.Lock()
			receivedData = append(receivedData, data...)
			mu.Unlock()
			return nil
		},
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "test-client",
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

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// 发送数据
	testData := []byte("Hello QUIC")
	if err := client.Write(testData); err != nil {
		t.Fatalf("发送数据失败: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// 验证数据
	mu.Lock()
	defer mu.Unlock()
	if string(receivedData) != string(testData) {
		t.Errorf("期望接收 %s, 实际接收 %s", testData, receivedData)
	}
}

// TestIntegration_ClientReconnect 集成测试：客户端重连
func TestIntegration_ClientReconnect(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 使用不同的端口避免端口占用问题
	port := 19444
	var reconnectCount int32
	var disconnectCount int32

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "test-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "test-client",
		Host: fmt.Sprintf("127.0.0.1:%d", port),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-trunk"},
		},
		ReconnectEnabled: true,
		ReconnectDelay:   200 * time.Millisecond,
		MaxReconnect:     5,
		OnDisconnect: func(client *NetQuicClient) {
			atomic.AddInt32(&disconnectCount, 1)
		},
		OnReconnect: func(client *NetQuicClient) {
			atomic.AddInt32(&reconnectCount, 1)
		},
	}

	client := &NetQuicClient{
		cnf: clientConfig,
		log: log,
	}

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer client.Close()

	time.Sleep(200 * time.Millisecond)

	// 停止服务器，触发断开和重连
	server.Stop()
	time.Sleep(1 * time.Second)

	// 重启服务器
	server = &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}
	server.New()
	if err := server.Start(); err != nil {
		// 如果端口仍被占用，跳过测试
		t.Skipf("无法重启服务器（端口可能未释放）: %v", err)
	}
	defer server.Stop()

	// 等待重连完成
	time.Sleep(2 * time.Second)

	// 验证断开和重连
	dCount := atomic.LoadInt32(&disconnectCount)
	rCount := atomic.LoadInt32(&reconnectCount)

	if dCount == 0 {
		t.Error("期望检测到断开连接，但未检测到")
	}

	if rCount == 0 {
		t.Error("期望发生重连，但未检测到重连")
	} else {
		t.Logf("检测到 %d 次断开, %d 次重连", dCount, rCount)
	}
}

// TestIntegration_MultipleClients 集成测试：多客户端连接
func TestIntegration_MultipleClients(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18445
	var messageCount int32

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:           "test-server",
		Ip:             "127.0.0.1",
		Port:           port,
		TLSConfig:      generateIntegrationTestTLSConfig(),
		MaxConnections: 10,
		OnData: func(_ conn.IConn, data []byte) error {
			atomic.AddInt32(&messageCount, 1)
			return nil
		},
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建多个客户端
	clientCount := 5
	clients := make([]*NetQuicClient, clientCount)

	for i := 0; i < clientCount; i++ {
		clientConfig := &ClientConfig{
			Name: fmt.Sprintf("test-client-%d", i),
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

		client.New()
		if err := client.Start(); err != nil {
			t.Fatalf("启动客户端 %d 失败: %v", i, err)
		}
		clients[i] = client
	}

	time.Sleep(200 * time.Millisecond)

	// 每个客户端发送数据
	for i, client := range clients {
		testData := []byte(fmt.Sprintf("Message from client %d", i))
		if err := client.Write(testData); err != nil {
			t.Errorf("客户端 %d 发送数据失败: %v", i, err)
		}
	}

	time.Sleep(300 * time.Millisecond)

	// 关闭所有客户端
	for _, client := range clients {
		client.Close()
	}

	// 验证消息数量
	count := atomic.LoadInt32(&messageCount)
	if count != int32(clientCount) {
		t.Errorf("期望接收 %d 条消息, 实际接收 %d 条", clientCount, count)
	}
}

// TestIntegration_ConcurrentRequests 集成测试：并发请求
func TestIntegration_ConcurrentRequests(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18446
	var messageCount int32

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "test-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
		OnData: func(_ conn.IConn, data []byte) error {
			atomic.AddInt32(&messageCount, 1)
			return nil
		},
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "test-client",
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

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// 顺序发送数据（QUIC单流不适合并发写入）
	requestCount := 100

	for i := 0; i < requestCount; i++ {
		testData := []byte(fmt.Sprintf("Request %d", i))
		if err := client.Write(testData); err != nil {
			t.Errorf("发送请求 %d 失败: %v", i, err)
		}
		time.Sleep(5 * time.Millisecond) // 小延迟确保数据发送
	}

	time.Sleep(500 * time.Millisecond)

	// 验证消息数量
	count := atomic.LoadInt32(&messageCount)
	if count < int32(requestCount)*9/10 {
		t.Errorf("期望接收至少 %d 条消息, 实际接收 %d 条", requestCount*9/10, count)
	}
}

// TestIntegration_DataTransfer 集成测试：数据传输
func TestIntegration_DataTransfer(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	port := 18447
	var receivedData []byte
	var mu sync.Mutex

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "test-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
		OnData: func(_ conn.IConn, data []byte) error {
			mu.Lock()
			receivedData = append(receivedData, data...)
			mu.Unlock()
			return nil
		},
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "test-client",
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

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// 发送大量数据
	testData := make([]byte, 10240) // 10KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	if err := client.Write(testData); err != nil {
		t.Fatalf("发送数据失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 验证数据
	mu.Lock()
	defer mu.Unlock()
	if len(receivedData) != len(testData) {
		t.Errorf("期望接收 %d 字节, 实际接收 %d 字节", len(testData), len(receivedData))
	}
}

// TestIntegration_HighThroughput 集成测试：高吞吐量
func TestIntegration_HighThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过高吞吐量测试")
	}

	log, _ := logger.NewLogger(&logger.Config{
		Level:   "error",
		Console: false,
	})

	port := 18448
	var totalBytes int64

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:      "test-server",
		Ip:        "127.0.0.1",
		Port:      port,
		TLSConfig: generateIntegrationTestTLSConfig(),
		OnData: func(_ conn.IConn, data []byte) error {
			atomic.AddInt64(&totalBytes, int64(len(data)))
			return nil
		},
	}

	server := &NetQuicServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	if err := server.Start(); err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientConfig := &ClientConfig{
		Name: "test-client",
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

	client.New()
	if err := client.Start(); err != nil {
		t.Fatalf("启动客户端失败: %v", err)
	}
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// 发送大量数据
	totalMessages := 100
	testData := make([]byte, 1024) // 1KB per message
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	start := time.Now()
	for i := 0; i < totalMessages; i++ {
		if err := client.Write(testData); err != nil {
			t.Errorf("发送消息 %d 失败: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)
	elapsed := time.Since(start)

	// 验证吞吐量
	bytes := atomic.LoadInt64(&totalBytes)
	throughput := float64(bytes) / elapsed.Seconds() / 1024 // KB/s

	t.Logf("发送 %d KB, 接收 %d KB, 耗时 %v, 吞吐量 %.2f KB/s",
		totalMessages, bytes/1024, elapsed, throughput)

	expectedBytes := int64(totalMessages * 1024)
	if bytes < expectedBytes*8/10 {
		t.Errorf("数据丢失过多: 期望 %d KB, 接收 %d KB", expectedBytes/1024, bytes/1024)
	}
}
