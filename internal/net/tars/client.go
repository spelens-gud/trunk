package tars

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/spelens-gud/logger"
)

// NetTarsClient TARS客户端
type NetTarsClient struct {
	cnf            *ClientConfig
	log            logger.ILogger
	comm           *tars.Communicator
	stopChan       chan struct{}
	isStop         bool
	reconnectCount int32
	mu             sync.RWMutex
}

// New 初始化客户端
func (c *NetTarsClient) New() {
	c.stopChan = make(chan struct{})
	c.isStop = true
	c.comm = tars.NewCommunicator()
}

// Start 启动客户端
func (c *NetTarsClient) Start() error {
	return c.connect()
}

// connect 建立连接
func (c *NetTarsClient) connect() error {
	c.mu.Lock()
	c.isStop = false
	c.mu.Unlock()

	c.log.Infof("TARS客户端连接成功: %s", c.cnf.Host)
	return nil
}

// handleDisconnect 处理断开连接
func (c *NetTarsClient) handleDisconnect() {
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
func (c *NetTarsClient) reconnect() {
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

// Close 关闭客户端
func (c *NetTarsClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isStop {
		return nil
	}

	c.isStop = true
	close(c.stopChan)

	c.log.Infof("TARS客户端已关闭")
	return nil
}

// IsConnected 检查连接状态
func (c *NetTarsClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.isStop
}

// GetReconnectCount 获取重连次数
func (c *NetTarsClient) GetReconnectCount() int32 {
	return atomic.LoadInt32(&c.reconnectCount)
}

// GetCommunicator 获取通信器
func (c *NetTarsClient) GetCommunicator() *tars.Communicator {
	return c.comm
}

// StringToProxy 获取代理对象
func (c *NetTarsClient) StringToProxy(obj string, proxy interface{}) error {
	if c.comm == nil {
		return fmt.Errorf("通信器未初始化")
	}
	c.comm.StringToProxy(obj, proxy)
	return nil
}
