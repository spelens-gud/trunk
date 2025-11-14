package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/spelens-gud/logger"
)

// NetQuicClient QUIC客户端
type NetQuicClient struct {
	cnf            *ClientConfig
	log            logger.ILogger
	conn           *quic.Conn
	stream         *quic.Stream
	stopChan       chan struct{}
	isStop         bool
	reconnectCount int32
	mu             sync.RWMutex
}

// New 初始化客户端
func (c *NetQuicClient) New() {
	c.stopChan = make(chan struct{})
	c.isStop = true
}

// Start 启动客户端
func (c *NetQuicClient) Start() error {
	return c.connect()
}

// connect 建立连接
func (c *NetQuicClient) connect() error {
	ctx := context.Background()

	tlsConf := c.cnf.TLSConfig
	if tlsConf == nil {
		tlsConf = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-trunk"},
		}
	}

	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	}

	conn, err := quic.DialAddr(ctx, c.cnf.Host, tlsConf, quicConfig)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		_ = conn.CloseWithError(0, "")
		return fmt.Errorf("打开流失败: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.stream = stream
	c.isStop = false
	c.mu.Unlock()

	c.log.Infof("QUIC客户端连接成功: %s", c.cnf.Host)

	if c.cnf.FirstPingFunc != nil {
		c.cnf.FirstPingFunc(c)
	}

	go c.readLoop()
	go c.pingLoop()

	return nil
}

// readLoop 读取数据循环
func (c *NetQuicClient) readLoop() {
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

		// 读取消息长度（4字节）
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(stream, lenBuf); err != nil {
			if err != io.EOF && !c.isStop {
				c.log.Errorf("读取消息长度失败: %v", err)
			}
			c.handleDisconnect()
			return
		}

		// 解析消息长度
		msgLen := uint32(lenBuf[0])<<24 | uint32(lenBuf[1])<<16 | uint32(lenBuf[2])<<8 | uint32(lenBuf[3])
		if msgLen == 0 || msgLen > 1024*1024 { // 最大1MB
			c.log.Errorf("无效的消息长度: %d", msgLen)
			c.handleDisconnect()
			return
		}

		// 读取消息内容
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(stream, data); err != nil {
			if err != io.EOF && !c.isStop {
				c.log.Errorf("读取消息内容失败: %v", err)
			}
			c.handleDisconnect()
			return
		}

		if c.cnf.OnData != nil {
			if err := c.cnf.OnData(c, data); err != nil {
				c.log.Errorf("数据处理失败: %v", err)
			}
		}
	}
}

// pingLoop 心跳循环
func (c *NetQuicClient) pingLoop() {
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
func (c *NetQuicClient) handleDisconnect() {
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
func (c *NetQuicClient) reconnect() {
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
func (c *NetQuicClient) Write(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stream == nil {
		return fmt.Errorf("流未连接")
	}

	// 写入消息长度（4字节）
	msgLen := uint32(len(data))
	lenBuf := []byte{
		byte(msgLen >> 24),
		byte(msgLen >> 16),
		byte(msgLen >> 8),
		byte(msgLen),
	}

	if _, err := c.stream.Write(lenBuf); err != nil {
		return fmt.Errorf("写入消息长度失败: %w", err)
	}

	// 写入消息内容
	if _, err := c.stream.Write(data); err != nil {
		return fmt.Errorf("写入消息内容失败: %w", err)
	}

	return nil
}

// Close 关闭客户端
func (c *NetQuicClient) Close() error {
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
		_ = c.conn.CloseWithError(0, "客户端关闭")
	}

	c.log.Infof("QUIC客户端已关闭")
	return nil
}

// IsConnected 检查连接状态
func (c *NetQuicClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.isStop
}

// GetReconnectCount 获取重连次数
func (c *NetQuicClient) GetReconnectCount() int32 {
	return atomic.LoadInt32(&c.reconnectCount)
}
