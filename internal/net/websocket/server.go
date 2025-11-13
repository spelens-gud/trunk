package webSocket

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

// Config ws服务配置
type Config struct {
	Name         string                                  // 服务名称
	Ip           string                                  // 服务ip
	Port         int                                     // 服务端口
	Route        string                                  // 路由
	Log          logger.ILogger                          // 日志
	Pprof        bool                                    // pprof
	OnConnect    func(conn conn.IConn)                   // 连接建立时调用
	OnData       func(conn conn.IConn, raw []byte) error // 数据处理
	OnClose      func(conn conn.IConn) error             // 连接关闭时调用
	WriteTimeout time.Duration                           // 写超时
	ReadTimeout  time.Duration                           // 读超时
	IdleTimeOut  time.Duration                           // 空闲超时
}

// NetServer ws 服务
type NetServer struct {
	Config
	stopChan   chan chan struct{}                   // 停止信号
	mux        *http.ServeMux                       // 路由
	httpServer *http.Server                         // http服务
	listener   net.Listener                         // 监听
	lock       sync.RWMutex                         // 锁
	nets       map[*conn.Conn[*websocket.Conn]]bool // 所有连接
}

func New(cfg Config) (server *NetServer, err error) {
	server = &NetServer{
		Config:   cfg,
		stopChan: make(chan chan struct{}),
		mux:      http.NewServeMux(),
		nets:     make(map[*conn.Conn[*websocket.Conn]]bool),
	}

	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler:      server.mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeOut,
	}
	cfg.Log.Infof("listening on:%s", server.httpServer.Addr)
	server.listener, err = net.Listen("tcp", server.httpServer.Addr)
	if err != nil {
		return nil, err
	}
	server.Port = server.listener.Addr().(*net.TCPAddr).Port
	return server, nil
}

// RunNet 运行
func (this *NetServer) RunNet(route string) error {
	upgrader := websocket.Upgrader{
		HandshakeTimeout:  3 * time.Second,
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin:       func(r *http.Request) bool { return true },
	}
	if this.Pprof {
		this.initPprof()
	}
	this.Log.Infof("RunNet%s%s", route, this.Route)
	this.mux.HandleFunc(route+this.Route, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r.Body != nil {
				if err := r.Body.Close(); err != nil {
					this.Log.Errorf("body close err:%s", err)
				}
			}
			err := recover()
			if err != nil {
				this.Log.Errorf("handle err:%v", err)
			}

		}()
		wsconn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			this.Log.Errorf("Upgrade error（%+v） %v", r.Header, err)
			return
		}

		cn := conn.NewConn(wsconn, conn.Config[*websocket.Conn]{
			Name:    this.Name,
			Host:    fmt.Sprintf("%s:%d", this.Ip, this.Port),
			Log:     this.Log,
			OnWrite: this.onWriteFunc,
			OnRead:  this.onReadFunc,
			OnClose: this.onCloseFunc,
			OnData:  this.Config.OnData,
		})
		this.lock.Lock()
		this.Log.Infof("net server add net:%p", cn)
		this.nets[cn] = true
		this.lock.Unlock()
		this.OnConnect(cn)
		cn.Start()
		this.lock.Lock()
		this.Log.Infof("net server del net:%p", cn)
		delete(this.nets, cn)
		this.lock.Unlock()
		this.OnClose(cn)
	})

	errChan := make(chan error)
	go func(errChan chan error) {
		errChan <- this.httpServer.Serve(this.listener)
	}(errChan)
	this.Log.Infof("http starting")
	select {
	case stopFinished := <-this.stopChan:
		this.httpServer.Shutdown(context.Background())
		stopFinished <- struct{}{}
		return nil
	case err := <-errChan:
		this.Log.Errorf("ListenAndServeErr: %s", err)
		return err
	}
	return nil
}

func (this *NetServer) HandleFunc(pattern string, handle func(w http.ResponseWriter, r *http.Request)) error {
	this.mux.HandleFunc(pattern, handle)
	return nil
}

func (this *NetServer) Stop() {
	stopDone := make(chan struct{}, 1)
	this.stopChan <- stopDone
	<-stopDone
}

func (this *NetServer) initPprof() {
	if this.mux != nil {
		this.mux.HandleFunc("/debug/pprof/", pprof.Index)
		this.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		this.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		this.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		this.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}

func (this *NetServer) onCloseFunc(cn *websocket.Conn) error {
	return cn.Close()
}

func (this *NetServer) onWriteFunc(cn *websocket.Conn, data []byte) error {
	if err := cn.SetWriteDeadline(time.Now().Add(this.WriteTimeout)); err != nil {
		this.Log.Warnf("SetWriteDeadline err:%s", err)
	}
	return cn.WriteMessage(websocket.BinaryMessage, data)
}

func (this *NetServer) onReadFunc(cn *websocket.Conn) (int, []byte, error) {
	if err := cn.SetReadDeadline(time.Now().Add(this.ReadTimeout)); err != nil {
		this.Log.Warnf("SetReadDeadline err:%s", err)
	}
	return cn.ReadMessage()
}
