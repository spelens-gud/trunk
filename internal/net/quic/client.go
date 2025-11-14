package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spelens-gud/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
	"golang.org/x/net/quic"
)

// ClientConfig QUIC客户端配置
type ClientConfig struct {
	NetConfig        conn.NetConfig[quic.Connection]
	TLSConfig        *tls.Config
	PingTicker       time.Duration
	PingFunc         func(client *QuicNetClient)
	FirstPingFunc    func(client *QuicNetClient)
	ReconnectEnabled bool
	ReconnectDelay   time.Duration
	MaxReconnect     int
	OnReconnect      func(client *QuicNetClient)
	OnDisconnect     func(client *QuicNetClient)
}

// QuicNetClient QUIC客户端
type QuicNetClient struct {
	cnf            *ClientConfig
	log            logger.ILogger
	conn           quic.Connection
	stream         quic.Stream
	stopChan       chan struct{}
	isStop         bool
	reconnectCount int32
	mu             sync.RWMutex
}

// New 初始化客户端
func (c *QuicNetClient) New() {
	c.stopChan = make(chan struct{})
	c.isStop = true
}

// Start 启动客户端
func (c *QuicNetClient) Start() error {
	return c.connect()
}

// connect 建立连接
func (c *QuicNetClient) connect() error {
	ctx := context.Background()

	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	}

	conn, err := quic.DialAddr(ctx, c.cnf.NetConfig.Host, c.cnf.TLSConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		conn.CloseWithError(0, "")
		return fmt.Errorf("打开流失败: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.stream = stream
	c.isStop = false
	c.mu.Unlock()

	c.log.Infof("QUIC客户端连接成功: %s", c.cnf.NetConfig.Host)

	if c.cnf.FirstPingFunc != nil {
		c.cnf.FirstPingFunc(c)
	}

	go c.readLoop()
	go c.pingLoop()

	return nil
}

// readLoop 读取数据循环
func (c *QuicNetClient) readLoop() {
	buf := make([]byte, 4096)

	for {
		if c.isStop {
			return
		}

		c.mu.RLock()
		stream := c.stream
		c.mu.RUnlock()

		if stream == nil {
			return
		}

		n, err := stream.Read(buf)
		if err != nil {
			c.log.Errorf("读取数据失败: %v", err)
			c.handleDisconnect()
			return
		}

		if c.cnf.NetConfig.OnData != nil {
			data := make([]byte, n)
			copy(data, buf[:n])

			if err := c.cnf.NetConfig.OnData(nil, data); err != nil {
				c.log.Errorf("数据处理失败: %v", err)
			}
		}
	}
}

// pingLoop 心跳循环
func (c *QuicNetClient) pingLoop() {
	if c.cnf.PingTicker <= 0 {
		return
	}

	ticker := time.NewTicker(c.cnf.PingTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if c.isStop {
				return
			}
			if c.cnf.PingFunc != nil {
				c.cnf.PingFunc(c)
			}
		case <-c.stopChan:
			return
		}
	}
}

// handleDisconnect 处理断开连接
func (c *QuicNetClient) handleDisconnect() {
	c.mu.Lock()
	c.isStop = true
	c.mu.Unlock()

	if c.cnf.OnDisconnect != nil {
		c.cnf.OnDisconnect(c)
	}

	if c.cnf.ReconnectEnabled {
		c.reconnect()
	}
}

// reconnect 重连
func (c *QuicNetClient) reconnect() {
	count := atomic.LoadInt32(&c.reconnectCount)

	if c.cnf.MaxReconnect > 0 && int(count) >= c.cnf.MaxReconnect {
		c.log.Errorf("达到最大重连次数: %d", count)
		return
	}

	atomic.AddInt32(&c.reconnectCount, 1)

	time.Sleep(c.cnf.ReconnectDelay)

	if err := c.connect(); err != nil {
		c.log.Errorf("重连失败: %v", err)
		c.reconnect()
		return
	}

	if c.cnf.OnReconnect != nil {
		c.cnf.OnReconnect(c)
	}
}

// Write 写入数据
func (c *QuicNetClient) Write(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stream == nil {
		return fmt.Errorf("流未连接")
	}

	_, err := c.stream.Write(data)
	return err
}

// Close 关闭客户端
func (c *QuicNetClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isStop {
		return nil
	}

	c.isStop = true
	close(c.stopChan)

	if c.stream != nil {
		c.stream.Close()
	}

	if c.conn != nil {
		c.conn.CloseWithError(0, "客户端关闭")
	}

	c.log.Infof("QUIC客户端已关闭")
	return nil
}

// IsConnected 检查连接状态
func (c *QuicNetClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.isStop && c.conn != nil
}

// GetReconnectCount 获取重连次数
func (c *QuicNetClient) GetReconnectCount() int32 {
	return atomic.LoadInt32(&c.reconnectCount)
}
