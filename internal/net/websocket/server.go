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
	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/logger"
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
			if r.Body == nil {
				return
			}

			if err := r.Body.Close(); err != nil {
				s.log.Errorf("关闭请求体失败:%s", err)
			}

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

		// 升级为websocket
		wsconn := assert.ShouldCall3RE(upgrader.Upgrade, w, r, nil, "WebSocket升级失败 请求头:", r.Header)

		// 创建连接
		cn := conn.NewConn(wsconn, conn.NetConfig[*websocket.Conn]{
			Name:    s.cnf.Name,
			Host:    fmt.Sprintf("%s:%d", s.cnf.Ip, s.cnf.Port),
			OnWrite: s.onWriteFunc,
			OnRead:  s.onReadFunc,
			OnClose: s.onCloseFunc,
			OnData:  s.cnf.OnData,
		})
		cn.SetLogger(s.log) // 设置 logger

		s.lock.Lock()
		s.nets[cn] = true // 添加连接
		s.connCount++
		s.totalAccepted++
		currentCount := s.connCount
		s.log.Infof("新连接建立:%p 当前连接数:%d 累计接受:%d", cn, currentCount, s.totalAccepted)
		s.lock.Unlock()

		// 启动连接回调函数
		s.cnf.OnConnect(cn)

		// 启动连接的读写循环
		cn.Start()

		// 等待连接关闭（通过监听 context）
		<-cn.GetContext().Done()

		s.lock.Lock()
		// 只有当连接还在 map 中时才删除和减少计数
		if _, exists := s.nets[cn]; exists {
			delete(s.nets, cn)
			s.connCount--
			currentCount := s.connCount
			s.log.Infof("连接断开:%p 当前连接数:%d", cn, currentCount)
		}
		s.lock.Unlock()

		// 启动关闭回调函数
		assert.ShouldCall1E[conn.IConn](s.cnf.OnClose, cn, "关闭连接失败")
	})

	errChan := make(chan error)
	go func(errChan chan error) {
		errChan <- s.httpServer.Serve(s.listener)
	}(errChan)

	s.log.Infof("HTTP服务启动成功")

	select {
	case stopFinished := <-s.stopChan:
		s.log.Infof("开始关闭服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		assert.ShouldCall1E(s.httpServer.Shutdown, ctx, "http关闭连接失败")
		s.closeAllConnections()
		stopFinished <- struct{}{}
	case err := <-errChan:
		s.log.Errorf("HTTP服务异常: %s", err)
	}
}

// HandleFunc 注册路由
func (s *WsNetServer) HandleFunc(pattern string, handle func(w http.ResponseWriter, r *http.Request)) {
	s.mux.HandleFunc(pattern, handle)
}

// Stop 停止服务器
func (s *WsNetServer) Stop() {
	stopDone := make(chan struct{}, 1)
	s.stopChan <- stopDone
	<-stopDone
	s.log.Infof("服务器已停止")
}

// GetConnectionCount 获取当前连接数
func (s *WsNetServer) GetConnectionCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.connCount
}

// ServerStats 服务器统计信息
type ServerStats struct {
	CurrentConnections int    // 当前连接数
	TotalAccepted      uint64 // 累计接受的连接数
	TotalRejected      uint64 // 累计拒绝的连接数
}

// GetStats 获取服务器统计信息
func (s *WsNetServer) GetStats() ServerStats {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return ServerStats{
		CurrentConnections: s.connCount,
		TotalAccepted:      s.totalAccepted,
		TotalRejected:      s.totalRejected,
	}
}

// BroadcastMessage 广播消息给所有连接
func (s *WsNetServer) BroadcastMessage(data []byte) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for cn := range s.nets {
		cn.Write(data)
	}
}

// BroadcastMessageExclude 广播消息给除指定连接外的所有连接
func (s *WsNetServer) BroadcastMessageExclude(data []byte, excludeConn conn.IConn) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for cn := range s.nets {
		if cn == excludeConn {
			continue
		}

		cn.Write(data)
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

// onCloseFunc 关闭连接处理函数
func (s *WsNetServer) onCloseFunc(cn *websocket.Conn) error {
	return cn.Close()
}

// onWriteFunc 写数据处理函数
func (s *WsNetServer) onWriteFunc(cn *websocket.Conn, data []byte) error {
	assert.ShouldCall1E(cn.SetWriteDeadline, time.Now().Add(s.cnf.GetWriteTimeout()), "SetWriteDeadline err:")
	return cn.WriteMessage(websocket.BinaryMessage, data)
}

// onReadFunc 读取数据处理函数
func (s *WsNetServer) onReadFunc(cn *websocket.Conn) (int, []byte, error) {
	assert.ShouldCall1E(cn.SetReadDeadline, time.Now().Add(s.cnf.GetReadTimeout()), "SetReadDeadline err:")
	return cn.ReadMessage()
}

// closeAllConnections 关闭所有连接
func (s *WsNetServer) closeAllConnections() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.log.Infof("开始关闭所有连接，当前连接数:%d", len(s.nets))

	for cn := range s.nets {
		assert.ShouldCall0E(cn.Close, "关闭连接失败")
	}

	// 重置连接数
	s.nets = make(map[*conn.Conn[*websocket.Conn]]bool)
	s.connCount = 0
	s.log.Infof("所有连接已关闭")
}
