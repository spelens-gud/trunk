package webSocket

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spelens-gud/trunk/internal/net/conn"
)

type WsNetServer struct {
	cnf           *ServerConfig                        // ws服务端配置
	log           logger.ILogger                       // 日志
	stopChan      chan chan struct{}                   // 停止信号
	mux           *http.ServeMux                       // 路由
	httpServer    *http.Server                         // http服务
	listener      net.Listener                         // 监听
	lock          sync.RWMutex                         // 锁
	nets          map[*conn.Conn[*websocket.Conn]]bool // 所有连接, key: 连接, value: true
	connCount     int                                  // 当前连接数
	totalAccepted uint64                               // 累计接受的连接数
	totalRejected uint64                               // 累计拒绝的连接数
}

// New 创建ws服务端
func (s *WsNetServer) New() {
	s.stopChan = make(chan chan struct{})
	s.mux = http.NewServeMux()
	s.nets = make(map[*conn.Conn[*websocket.Conn]]bool)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", s.cnf.Port),
		Handler:      s.mux,
		ReadTimeout:  s.cnf.GetReadTimeout(),
		WriteTimeout: s.cnf.GetWriteTimeout(),
		IdleTimeout:  s.cnf.GetIdleTimeOut(),
	}

	s.log.Infof("监听地址:%s", s.httpServer.Addr)
	s.listener = assert.ShouldCall2RE(net.Listen, "tcp", s.httpServer.Addr)
}

// RunNet 启动ws服务端
func (s *WsNetServer) RunNet(route string) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout:  3 * time.Second,
		ReadBufferSize:    s.cnf.GetReadBufferSize(),
		WriteBufferSize:   s.cnf.GetWriteBufferSize(),
		EnableCompression: s.cnf.Compression,
		CheckOrigin:       func(r *http.Request) bool { return true },
	}

	assert.MayTrue(s.cnf.Pprof, func() {
		s.initPprof()
	})

	s.log.Infof("启动网络服务 路由:%s%s", route, s.cnf.Route)

	s.mux.HandleFunc(route+s.cnf.Route, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			assert.MayTrue(r.Body != nil && r.Body.Close() != nil, func() {
				s.log.Errorf("关闭请求体失败:%s", err)
			})

			if err := recover(); err != nil {
				s.log.Errorf("处理请求异常:%v", err)
			}
		}()

		// 检查连接数限制
		if s.checkConnectionsLimit(w, r) {
			s.log.Warnf("连接数已达上限(%d)，拒绝新连接 来源:%s", s.cnf.GetMaxConnections(), r.RemoteAddr)
			http.Error(w, "服务器连接数已满", http.StatusServiceUnavailable)
			return
		}

		wsconn := assert.ShouldCall3RE(upgrader.Upgrade, w, r, nil, "WebSocket升级失败 请求头:(%+v) 错误:", r.Header)

		cn := conn.NewConn(wsconn, conn.Config[*websocket.Conn]{
			Name:    s.cnf.Name,
			Host:    fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port),
			Log:     s.log,
			OnWrite: s.onWriteFunc,
			OnRead:  s.onReadFunc,
			OnClose: s.onCloseFunc,
			OnData:  s.cnf.OnData,
		})

		this.lock.Lock()
		this.nets[cn] = true
		this.connCount++
		this.totalAccepted++
		currentCount := this.connCount
		this.Log.Infof("新连接建立:%p 当前连接数:%d 累计接受:%d", cn, currentCount, this.totalAccepted)
		this.lock.Unlock()

		this.OnConnect(cn)
		cn.Start()

		this.lock.Lock()
		delete(this.nets, cn)
		this.connCount--
		currentCount = this.connCount
		this.Log.Infof("连接断开:%p 当前连接数:%d", cn, currentCount)
		this.lock.Unlock()

		this.OnClose(cn)
	})

	errChan := make(chan error)
	go func(errChan chan error) {
		errChan <- this.httpServer.Serve(this.listener)
	}(errChan)
	this.Log.Infof("HTTP服务启动成功")
	select {
	case stopFinished := <-this.stopChan:
		this.Log.Infof("开始关闭服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		this.httpServer.Shutdown(ctx)
		this.closeAllConnections()
		stopFinished <- struct{}{}
		return nil
	case err := <-errChan:
		this.Log.Errorf("HTTP服务异常: %s", err)
		return err
	}
}

// checkConnectionsLimit 检查连接数限制
func (s *WsNetServer) checkConnectionsLimit(w http.ResponseWriter, r *http.Request) bool {
	// 锁的颗粒度控制
	if s.cnf.GetMaxConnections() <= 0 {
		return false
	}

	// 读锁, 避免并发修改, 提高效率
	s.lock.RLock()
	currentCount := s.connCount
	s.lock.RUnlock()

	if currentCount >= s.cnf.GetMaxConnections() {
		// 写锁, 避免数据竞争
		s.lock.Lock()
		s.totalRejected++
		s.lock.Unlock()
		return true
	}

	return false
}

// initPprof 初始化pprof
func (s *WsNetServer) initPprof() {
	if s.mux != nil {
		s.mux.HandleFunc("/debug/pprof/", pprof.Index)
		s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}
func (s *WsNetServer) onCloseFunc(cn *websocket.Conn) error {
	return cn.Close()
}

func (s *WsNetServer) onWriteFunc(cn *websocket.Conn, data []byte) error {
	if err := cn.SetWriteDeadline(time.Now().Add(s.cnf.WriteTimeout)); err != nil {
		s.log.Warnf("SetWriteDeadline err:%s", err)
	}
	return cn.WriteMessage(websocket.BinaryMessage, data)
}

func (s *WsNetServer) onReadFunc(cn *websocket.Conn) (int, []byte, error) {
	if err := cn.SetReadDeadline(time.Now().Add(s.cnf.ReadTimeout)); err != nil {
		s.log.Warnf("SetReadDeadline err:%s", err)
	}
	return cn.ReadMessage()
}
