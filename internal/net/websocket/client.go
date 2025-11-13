package webSocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

type WsNetClient struct {
	conn           *conn.Conn[*websocket.Conn] // 连接
	log            logger.ILogger              // 日志
	cnf            *ClientConfig               // 配置
	stopChan       chan chan struct{}          // 停止通道
	pingTicker     *time.Ticker                // ping定时器
	isStop         bool                        // 是否停止
	isFirstPing    bool                        // 是否首次ping
	reconnectCount int                         // 重连次数
}

// New 创建ws客户端
func (c *WsNetClient) New() {
	c.stopChan = make(chan chan struct{})
	c.isStop = true
}

// Start 运行
func (c *WsNetClient) Start() {
	go c.conn.Start()
	if c.cnf.PingTicker > 0 {
		c.pingTicker = time.NewTicker(c.cnf.PingTicker)
	}
	for {
		select {
		case <-c.pingTicker.C:
			assert.MayTrue(c.cnf.FirstPingFunc != nil && !c.isFirstPing, func() {
				c.cnf.FirstPingFunc(c)
				c.isFirstPing = true
			})
			assert.MayTrue(c.cnf.PingFunc != nil, func() {
				c.cnf.PingFunc(c)
			})
		case ch := <-c.stopChan:
			if ch == nil {
				return
			}

			c.pingTicker.Stop()
			assert.ShouldCall0E(c.conn.Close, "conn关闭失败")
			ch <- struct{}{}
			return
		}
	}
}

// StartWithReconnect 启动客户端并支持自动重连
func (c *WsNetClient) StartWithReconnect() {
	for {
		if err := c.Daily(); err != nil {
			c.log.Errorf("连接失败: %v", err)

			// 如果未启用自动重连，则停止运行
			if !c.cnf.ReconnectEnabled {
				return
			}

			c.reconnectCount++
			if c.cnf.MaxReconnect > 0 && c.reconnectCount > c.cnf.MaxReconnect {
				c.log.Errorf("达到最大重连次数(%d)，停止重连", c.cnf.MaxReconnect)
				return
			}

			c.log.Infof("将在 %v 后进行第 %d 次重连...", c.cnf.ReconnectDelay, c.reconnectCount)
			time.Sleep(c.cnf.ReconnectDelay)
			continue
		}

		// 连接成功，重置重连计数
		if c.reconnectCount > 0 {
			c.log.Infof("重连成功")
			assert.MayTrue(c.cnf.OnReconnect != nil, func() {
				c.cnf.OnReconnect(c)
			})
		}

		c.reconnectCount = 0

		// 启动客户端
		c.Start()

		// 如果连接断开且启用了自动重连，继续循环
		if !c.cnf.ReconnectEnabled {
			break
		}

		// 断线处理
		assert.MayTrue(c.cnf.OnDisconnect != nil, func() {
			c.cnf.OnDisconnect(c)
		})

		c.log.Infof("连接断开，准备重连...")
		time.Sleep(c.cnf.ReconnectDelay)
	}
}

// Daily 建立连接(拨号)
func (c *WsNetClient) Daily() error {
	dialer := &websocket.Dialer{
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		HandshakeTimeout: 30 * time.Second,
	}

	con, _, err := dialer.Dial(c.cnf.Host, nil)
	if err != nil {
		return err
	}
	c.conn = conn.NewConn(con, c.cnf.NetConfig)
	c.isStop = false
	c.isFirstPing = false
	c.log.Infof("连接建立成功: %s", c.cnf.Host)
	return nil
}

// Close 关闭客户端连接
func (c *WsNetClient) Close() error {
	if c.isStop {
		return nil
	}
	defer func() {
		c.isStop = true
	}()

	done := make(chan struct{})
	c.stopChan <- done
	<-done
	c.log.Infof("客户端连接已关闭")
	return nil
}

// IsConnected 检查是否已连接
func (c *WsNetClient) IsConnected() bool {
	return !c.isStop && c.conn != nil
}

// GetReconnectCount 获取重连次数
func (c *WsNetClient) GetReconnectCount() int {
	return c.reconnectCount
}

// SendMsg 发送消息
func (c *WsNetClient) SendMsg(bs []byte) {
	c.conn.Write(bs)
}

// onCloseFunc 关闭连接处理函数
func (c *WsNetClient) onCloseFunc(cn *websocket.Conn) error {
	return cn.Close()
}

// onWriteFunc 写数据处理函数
func (c *WsNetClient) onWriteFunc(cn *websocket.Conn, data []byte) error {
	assert.ShouldCall1E(cn.SetWriteDeadline, time.Now().Add(c.cnf.GetWriteTimeout()), "SetWriteDeadline err:")
	return cn.WriteMessage(websocket.BinaryMessage, data)
}

// onReadFunc 读取数据处理函数
func (c *WsNetClient) onReadFunc(cn *websocket.Conn) (int, []byte, error) {
	assert.ShouldCall1E(cn.SetReadDeadline, time.Now().Add(c.cnf.GetReadTimeout()), "SetReadDeadline err:")
	return cn.ReadMessage()
}
