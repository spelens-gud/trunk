package conn

import (
	"sync"
	"time"

	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
)

// Conn 连接
type Conn[T any] struct {
	cnf       NetConfig[T]   // 配置
	log       logger.ILogger // 日志
	conn      T              // 连接
	lock      sync.RWMutex   // 锁
	writeChan chan []byte    // 写数据通道
	closed    bool           // 是否关闭
	createAt  time.Time      // 创建时间
	updateAt  int64          // 更新时间
}

// NewConn 创建连接
func NewConn[T any](conn T, cfg NetConfig[T]) *Conn[T] {
	return &Conn[T]{
		conn:      conn,
		cnf:       cfg,
		writeChan: make(chan []byte, 64),
		createAt:  time.Now(),
	}
}

// SetLogger 设置日志
func (s *Conn[T]) SetLogger(logger logger.ILogger) {
	s.log = logger
}

// Start 启动
func (s *Conn[T]) Start() {
	go logger.WithRecover(s.log, func() {
		s.write()
	})

	s.read()
}

// Write 写数据
func (s *Conn[T]) Write(b []byte) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// 如果通道关闭则退出
	if s.closed || b == nil {
		return
	}

	// 数据给到写通道
	s.writeChan <- b
}

// write 写数据
func (s *Conn[T]) write() {
	for {
		select {
		case msg := <-s.writeChan:
			// 如果通道关闭则退出
			if msg == nil {
				return
			}

			// 如果没有写处理函数则退出
			if s.cnf.OnWrite == nil {
				return
			}

			// 写数据
			if err := s.cnf.OnWrite(s.conn, msg); err != nil {
				s.log.Errorf("Write Error: %v", err)
				assert.ShouldCall0E(s.Close, "conn 关闭错误")
				return
			}
		}
	}
}

// read 读数据
func (s *Conn[T]) read() {
	for {
		// 如果没有读处理函数则退出
		if s.cnf.OnRead == nil {
			continue
		}

		// 如果通道关闭则退出
		if s.closed {
			return
		}

		// 读数据
		if _, bs, err := s.cnf.OnRead(s.conn); err != nil {
			s.log.Warnf("conn read err: %p, %p %s ", s, s.conn, err)
			assert.ShouldCall0E(s.Close, "conn 关闭错误")
			break
		} else {
			assert.ShouldCall2E(s.cnf.OnData, s, bs, "处理数据函数发生错误")
		}
	}
}

// Close 关闭
func (s *Conn[T]) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 如果已经关闭则退出
	if s.closed {
		return nil
	}

	s.closed = true
	close(s.writeChan)

	// 关闭处理函数
	if s.cnf.OnClose != nil {
		return s.cnf.OnClose(s.conn)
	}

	return nil
}

// GetConn 获取连接
func (s *Conn[T]) GetConn() T {
	return s.conn
}

// SetId 设置id
func (s *Conn[T]) SetId(id uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 如果已经关闭则退出
	if s.closed {
		return
	}

	s.cnf.Id = id
}

// GetId 获取id
func (s *Conn[T]) GetId() uint64 {
	return s.cnf.Id
}
