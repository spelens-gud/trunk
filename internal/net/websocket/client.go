package webSocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

type ClientConfig struct {
	conn.Config[*websocket.Conn]
	PingTicker    time.Duration
	PingFunc      func(*Client)
	FirstPingFunc func(*Client)
}

type Client struct {
	*conn.Conn[*websocket.Conn]
	cfg ClientConfig

	stopChan    chan chan struct{}
	pingTicker  *time.Ticker
	isStop      bool
	isFirstPing bool
}

func NewClient(cfg ClientConfig) (*Client, error) {

	c := &Client{
		cfg:      cfg,
		stopChan: make(chan chan struct{}),
		isStop:   true,
	}
	if cfg.OnWrite == nil {
		cfg.OnWrite = c.OnWrite
	}
	if cfg.OnRead == nil {
		cfg.OnRead = c.OnRead
	}
	if cfg.OnClose == nil {
		cfg.OnClose = c.OnClose
	}
	return c, nil
}

func (this *Client) Start() {
	go this.Conn.Start()
	if this.cfg.PingTicker > 0 {
		this.pingTicker = time.NewTicker(this.cfg.PingTicker)
	}
	for {
		select {
		case <-this.pingTicker.C:
			if this.cfg.FirstPingFunc != nil && !this.isFirstPing {
				this.cfg.FirstPingFunc(this)
				this.isFirstPing = true
			}
			if this.cfg.PingFunc != nil {
				this.cfg.PingFunc(this)
			}
		case ch := <-this.stopChan:
			if ch != nil {
				this.pingTicker.Stop()
				this.Conn.Close()
				ch <- struct{}{}
			}
			return
		}
	}
}

func (this *Client) Daily() error {
	dialer := &websocket.Dialer{
		ReadBufferSize:   2048,
		WriteBufferSize:  2048,
		HandshakeTimeout: 30 * time.Second,
	}
	c, _, err := dialer.Dial(this.cfg.Host, nil)
	if err != nil {
		return err
	}
	this.Conn = conn.NewConn(c, this.cfg.Config)
	this.isStop = false
	return nil
}

func (this *Client) Close() error {
	if this.isStop {
		return nil
	}
	defer func() {
		this.isStop = true
	}()
	done := make(chan struct{})
	this.stopChan <- done
	<-done
	return nil
}

func (this *Client) SendMsg(bs []byte) {
	this.Conn.Write(bs)
}

func (this *Client) onCloseFunc(cn *websocket.Conn) error {
	return cn.Close()
}

func (this *Client) onWriteFunc(cn *websocket.Conn, data []byte) error {
	if err := cn.SetWriteDeadline(time.Now().Add(this.WriteTimeout)); err != nil {
		this.Log.Warnf("SetWriteDeadline err:%s", err)
	}
	return cn.WriteMessage(websocket.BinaryMessage, data)
}

func (this *Client) onReadFunc(cn *websocket.Conn) (int, []byte, error) {
	if err := cn.SetReadDeadline(time.Now().Add(this.ReadTimeout)); err != nil {
		this.Log.Warnf("SetReadDeadline err:%s", err)
	}
	return cn.ReadMessage()
}
