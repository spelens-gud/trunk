package grpc

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spelens-gud/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// NetGrpcClient gRPC客户端
type NetGrpcClient struct {
	cnf            *ClientConfig
	log            logger.ILogger
	conn           *grpc.ClientConn
	stopChan       chan struct{}
	isStop         bool
	reconnectCount int32
	mu             sync.RWMutex
}

// New 初始化客户端
func (c *NetGrpcClient) New() {
	c.stopChan = make(chan struct{})
	c.isStop = true
}

// Start 启动客户端
func (c *NetGrpcClient) Start() error {
	return c.connect()
}

// connect 建立连接
func (c *NetGrpcClient) connect() error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                c.cnf.KeepAliveTime,
			Timeout:             c.cnf.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.Dial(c.cnf.Host, opts...)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.isStop = false
	c.mu.Unlock()

	c.log.Infof("gRPC客户端连接成功: %s", c.cnf.Host)

	go c.watchConnection()

	return nil
}

// watchConnection 监控连接状态
func (c *NetGrpcClient) watchConnection() {
	for {
		if c.isStop {
			return
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		state := conn.GetState()

		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			c.log.Warnf("连接状态异常: %v", state)
			c.handleDisconnect()
			return
		}

		time.Sleep(time.Second)
	}
}

// handleDisconnect 处理断开连接
func (c *NetGrpcClient) handleDisconnect() {
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
func (c *NetGrpcClient) reconnect() {
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
func (c *NetGrpcClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isStop {
		return nil
	}

	c.isStop = true
	close(c.stopChan)

	if c.conn != nil {
		c.conn.Close()
	}

	c.log.Infof("gRPC客户端已关闭")
	return nil
}

// IsConnected 检查连接状态
func (c *NetGrpcClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isStop || c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// GetReconnectCount 获取重连次数
func (c *NetGrpcClient) GetReconnectCount() int32 {
	return atomic.LoadInt32(&c.reconnectCount)
}

// GetConn 获取原生gRPC连接
func (c *NetGrpcClient) GetConn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// Invoke 调用RPC方法
func (c *NetGrpcClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("连接未建立")
	}

	return conn.Invoke(ctx, method, args, reply, opts...)
}
