package conn

import (
	"sync"
	"time"
)

// Conn 连接
type Conn[T any] struct {
	Config[T]                // 配置
	conn         T           // 连接
	sync.RWMutex             // 锁
	writeChan    chan []byte // 写数据通道
	closed       bool        // 是否关闭
	createAt     time.Time   // 创建时间
	updateAt     int64       // 更新时间
}

// NewConn 创建连接
func NewConn[T any](conn T, cfg Config[T]) *Conn[T] {
	c := &Conn[T]{
		conn:      conn,
		Config:    cfg,
		writeChan: make(chan []byte, 64),
		createAt:  time.Now(),
	}
	return c
}

// Start 启动
func (s *Conn[T]) Start() {
	go s.write()
	s.read()
}

// Write 写数据
func (s *Conn[T]) Write(b []byte) {
	s.RLock()
	defer s.RUnlock()
	if s.closed || b == nil {
		return
	}
	s.writeChan <- b
}

// write 写数据
func (s *Conn[T]) write() {
	for {
		select {
		case msg := <-s.writeChan:
			if msg != nil {
				if s.OnWrite != nil {
					if err := s.OnWrite(s.conn, msg); err != nil {
						s.Log.Errorf("Write Error: %v", err)
						s.Close()
						return
					}

				}
			} else {
				return
			}
		}
	}
}

// read 读数据
func (s *Conn[T]) read() {
	for {
		if s.OnRead == nil {
			continue
		}
		if s.closed {
			return
		}
		if _, bs, err := s.OnRead(s.conn); err != nil {
			s.Log.Warnf("conn read err: %p, %p %s ", s, s.conn, err)
			s.Close()
			break
		} else {
			s.OnData(s, bs)
		}
	}
}

// Close 关闭
func (s *Conn[T]) Close() error {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	close(s.writeChan)
	if s.OnClose != nil {
		return s.OnClose(s.conn)
	}
	return nil
}

// GetConn 获取连接
func (s *Conn[T]) GetConn() T {
	return s.conn
}

// SetId 设置id
func (s *Conn[T]) SetId(id uint64) {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return
	}
	s.Config.Id = id
}

// GetId 获取id
func (s *Conn[T]) GetId() uint64 {
	return s.Config.Id
}
