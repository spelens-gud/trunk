package webSocket

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// TestIntegration_ServerClientCommunication 集成测试：服务器与客户端通信
func TestIntegration_ServerClientCommunication(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 19000
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建服务器
	serverConfig := &ServerConfig{
		Name:  "integration-server",
		Ip:    "127.0.0.1",
		Port:  port,
		Route: "/ws",
		OnConnect: func(c conn.IConn) {
			t.Logf("服务器: 客户端已连接")
		},
		OnData: func(c conn.IConn, data []byte) error {
			t.Logf("服务器: 收到数据 %s", string(data))
			// 回显数据
			c.Write(data)
			return nil
		},
		OnClose: func(c conn.IConn) error {
			t.Logf("服务器: 客户端已断开")
			return nil
		},
		MaxConnections: 10,
	}

	server := &WsNetServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()

	// 启动服务器
	go server.RunNet("")

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 创建客户端
	receivedData := make(chan []byte, 1)
	clientConfig := &ClientConfig{
		NetConfig: conn.NetConfig[*websocket.Conn]{
			Name: "integration-client",
			Host: fmt.Sprintf("ws://127.0.0.1:%d/ws", port),
			OnWrite: func(cn *websocket.Conn, data []byte) error {
				return cn.WriteMessage(websocket.BinaryMessage, data)
			},
			OnRead: func(cn *websocket.Conn) (int, []byte, error) {
				return cn.ReadMessage()
			},
			OnClose: func(cn *websocket.Conn) error {
				return cn.Close()
			},
			OnData: func(c conn.IConn, data []byte) error {
				t.Logf("客户端: 收到数据 %s", string(data))
				receivedData <- data
				return nil
			},
		},
		PingTicker:       10 * time.Second,
		ReconnectEnabled: false,
	}

	client := &WsNetClient{
		cnf: clientConfig,
		log: log,
	}

	client.New()

	// 连接服务器
	if err := client.Daily(); err != nil {
		t.Fatalf("客户端连接失败: %v", err)
	}

	// 启动客户端
	go client.Start()

	// 等待连接建立
	time.Sleep(200 * time.Millisecond)

	// 发送测试数据
	testData := []byte("Hello, WebSocket!")
	client.SendMsg(testData)

	// 等待接收数据
	select {
	case data := <-receivedData:
		if string(data) != string(testData) {
			t.Errorf("期望收到 '%s', 实际收到 '%s'", string(testData), string(data))
		}
	case <-time.After(2 * time.Second):
		t.Error("超时：未收到服务器响应")
	}

	// 清理
	if err := client.Close(); err != nil {
		return
	}

	server.Stop()
}

// TestIntegration_MultipleClients 集成测试：多客户端连接
func TestIntegration_MultipleClients(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 19001
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	// 创建服务器
	var connectedClients int
	var mu sync.Mutex

	serverConfig := &ServerConfig{
		Name:  "multi-client-server",
		Ip:    "127.0.0.1",
		Port:  port,
		Route: "/ws",
		OnConnect: func(c conn.IConn) {
			mu.Lock()
			connectedClients++
			mu.Unlock()
			t.Logf("客户端已连接，当前连接数: %d", connectedClients)
		},
		OnData: func(c conn.IConn, data []byte) error {
			c.Write(data)
			return nil
		},
		OnClose: func(c conn.IConn) error {
			mu.Lock()
			connectedClients--
			mu.Unlock()
			return nil
		},
		MaxConnections: 10,
	}

	server := &WsNetServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	go server.RunNet("")
	time.Sleep(500 * time.Millisecond)

	// 创建多个客户端
	clientCount := 5
	clients := make([]*WsNetClient, clientCount)

	for i := 0; i < clientCount; i++ {
		clientConfig := &ClientConfig{
			NetConfig: conn.NetConfig[*websocket.Conn]{
				Name: fmt.Sprintf("client-%d", i),
				Host: fmt.Sprintf("ws://127.0.0.1:%d/ws", port),
				OnWrite: func(cn *websocket.Conn, data []byte) error {
					return cn.WriteMessage(websocket.BinaryMessage, data)
				},
				OnRead: func(cn *websocket.Conn) (int, []byte, error) {
					return cn.ReadMessage()
				},
				OnClose: func(cn *websocket.Conn) error {
					return cn.Close()
				},
				OnData: func(c conn.IConn, data []byte) error {
					return nil
				},
			},
			ReconnectEnabled: false,
		}

		client := &WsNetClient{
			cnf: clientConfig,
			log: log,
		}

		client.New()
		if err := client.Daily(); err != nil {
			t.Fatalf("客户端 %d 连接失败: %v", i, err)
		}

		go client.Start()
		clients[i] = client
	}

	// 等待所有客户端连接
	time.Sleep(500 * time.Millisecond)

	// 验证连接数
	stats := server.GetStats()
	if stats.CurrentConnections != clientCount {
		t.Errorf("期望连接数 = %d, 实际 = %d", clientCount, stats.CurrentConnections)
	}

	// 清理
	for _, client := range clients {
		client.Close()
	}
	time.Sleep(200 * time.Millisecond)
	server.Stop()
}

// TestIntegration_ConnectionLimit 集成测试：连接数限制
func TestIntegration_ConnectionLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 19002
	maxConnections := 3

	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	serverConfig := &ServerConfig{
		Name:  "limit-server",
		Ip:    "127.0.0.1",
		Port:  port,
		Route: "/ws",
		OnConnect: func(c conn.IConn) {
			t.Logf("客户端已连接")
		},
		OnData: func(c conn.IConn, data []byte) error {
			return nil
		},
		OnClose: func(c conn.IConn) error {
			return nil
		},
		MaxConnections: maxConnections,
	}

	server := &WsNetServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	go server.RunNet("")
	time.Sleep(500 * time.Millisecond)

	// 尝试创建超过限制的客户端
	attemptCount := maxConnections + 2
	successCount := 0
	failCount := 0

	for i := 0; i < attemptCount; i++ {
		clientConfig := &ClientConfig{
			NetConfig: conn.NetConfig[*websocket.Conn]{
				Name: fmt.Sprintf("client-%d", i),
				Host: fmt.Sprintf("ws://127.0.0.1:%d/ws", port),
				OnWrite: func(cn *websocket.Conn, data []byte) error {
					return cn.WriteMessage(websocket.BinaryMessage, data)
				},
				OnRead: func(cn *websocket.Conn) (int, []byte, error) {
					return cn.ReadMessage()
				},
				OnClose: func(cn *websocket.Conn) error {
					return cn.Close()
				},
				OnData: func(c conn.IConn, data []byte) error {
					return nil
				},
			},
			ReconnectEnabled: false,
		}

		client := &WsNetClient{
			cnf: clientConfig,
			log: log,
		}

		client.New()
		if err := client.Daily(); err != nil {
			failCount++
			t.Logf("客户端 %d 连接失败（预期）: %v", i, err)
		} else {
			successCount++
			go client.Start()
			defer client.Close()
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Logf("成功连接: %d, 失败连接: %d", successCount, failCount)

	stats := server.GetStats()
	t.Logf("服务器统计: 当前连接=%d, 累计接受=%d, 累计拒绝=%d",
		stats.CurrentConnections, stats.TotalAccepted, stats.TotalRejected)

	if stats.CurrentConnections > maxConnections {
		t.Errorf("当前连接数 (%d) 超过了限制 (%d)", stats.CurrentConnections, maxConnections)
	}

	server.Stop()
}

// TestIntegration_BroadcastMessage 集成测试：广播消息
func TestIntegration_BroadcastMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	port := 19003
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	serverConfig := &ServerConfig{
		Name:  "broadcast-server",
		Ip:    "127.0.0.1",
		Port:  port,
		Route: "/ws",
		OnConnect: func(c conn.IConn) {
			t.Logf("客户端已连接")
		},
		OnData: func(c conn.IConn, data []byte) error {
			return nil
		},
		OnClose: func(c conn.IConn) error {
			return nil
		},
		MaxConnections: 10,
	}

	server := &WsNetServer{
		cnf: serverConfig,
		log: log,
	}

	server.New()
	go server.RunNet("")
	time.Sleep(500 * time.Millisecond)

	// 创建多个客户端
	clientCount := 3
	receivedCounts := make([]int, clientCount)
	var mu sync.Mutex

	for i := 0; i < clientCount; i++ {
		idx := i
		clientConfig := &ClientConfig{
			NetConfig: conn.NetConfig[*websocket.Conn]{
				Name: fmt.Sprintf("client-%d", idx),
				Host: fmt.Sprintf("ws://127.0.0.1:%d/ws", port),
				OnWrite: func(cn *websocket.Conn, data []byte) error {
					return cn.WriteMessage(websocket.BinaryMessage, data)
				},
				OnRead: func(cn *websocket.Conn) (int, []byte, error) {
					return cn.ReadMessage()
				},
				OnClose: func(cn *websocket.Conn) error {
					return cn.Close()
				},
				OnData: func(c conn.IConn, data []byte) error {
					mu.Lock()
					receivedCounts[idx]++
					mu.Unlock()
					t.Logf("客户端 %d 收到广播: %s", idx, string(data))
					return nil
				},
			},
			ReconnectEnabled: false,
		}

		client := &WsNetClient{
			cnf: clientConfig,
			log: log,
		}

		client.New()
		if err := client.Daily(); err != nil {
			t.Fatalf("客户端 %d 连接失败: %v", idx, err)
		}

		go client.Start()
		defer client.Close()
	}

	time.Sleep(500 * time.Millisecond)

	// 广播消息
	broadcastMsg := []byte("Broadcast Message")
	server.BroadcastMessage(broadcastMsg)

	// 等待消息传递
	time.Sleep(500 * time.Millisecond)

	// 验证所有客户端都收到了消息
	for i, count := range receivedCounts {
		if count != 1 {
			t.Errorf("客户端 %d 期望收到 1 条消息, 实际收到 %d 条", i, count)
		}
	}

	server.Stop()
}
