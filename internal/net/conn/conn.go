package conn

import (
	"context"
	"sync"
	"time"

	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/logger"
)

// OnConnectFunc 连接建立时调用
type OnConnectFunc[T any] func(conn T)

// OnWriteFunc 写数据处理
type OnWriteFunc[T any] func(conn T, raw []byte) error

// OnReadFunc 读数据处理
type OnReadFunc[T any] func(conn T) (int, []byte, error)

// OnDataFunc 数据处理
type OnDataFunc func(conn IConn, raw []byte) error

// OnCloseFunc 关闭处理
type OnCloseFunc[T any] func(conn T) error

// IConn 连接接口
type IConn interface {
	// Start 启动连接(启动读写 goroutine)
	Start()
	// Write 写数据(非阻塞,带超时)
	Write(b []byte)
	// Close 关闭连接
	Close() error
	// SetId 设置连接 ID
	SetId(id uint64)
	// GetId 获取连接 ID
	GetId() uint64
	// IsClosed 检查连接是否已关闭
	IsClosed() bool
	// GetCreateTime 获取创建时间
	GetCreateTime() time.Time
	// GetLastActiveTime 获取最后活跃时间
	GetLastActiveTime() time.Time
}

// Conn 连接
type Conn[T any] struct {
	cnf        NetConfig[T]       // 配置
	log        logger.ILogger     // 日志
	conn       T                  // 连接
	lock       sync.RWMutex       // 锁
	writeChan  chan []byte        // 写数据通道
	closed     bool               // 是否关闭
	createAt   time.Time          // 创建时间
	lastActive time.Time          // 最后活跃时间
	ctx        context.Context    // 上下文
	cancel     context.CancelFunc // 取消函数
}

// NewConn 创建连接
func NewConn[T any](conn T, cfg NetConfig[T]) *Conn[T] {
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now()

	return &Conn[T]{
		conn:       conn,
		cnf:        cfg,
		writeChan:  make(chan []byte, 64),
		createAt:   now,
		lastActive: now,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetLogger 设置日志
func (s *Conn[T]) SetLogger(logger logger.ILogger) {
	s.log = logger
}

// Start 启动
func (s *Conn[T]) Start() {
	// 启动写 goroutine
	go logger.WithRecover(s.log, func() {
		s.write()
	})

	// 启动读 goroutine
	go logger.WithRecover(s.log, func() {
		s.read()
	})

	// 如果配置了空闲超时,启动空闲检测
	if s.cnf.IdleTimeOut > 0 {
		go logger.WithRecover(s.log, func() {
			s.checkIdle()
		})
	}
}

// Write 写数据(非阻塞,带超时)
func (s *Conn[T]) Write(b []byte) {
	s.lock.RLock()
	closed := s.closed
	s.lock.RUnlock()

	// 如果通道关闭或数据为空则退出
	if closed || b == nil {
		s.log.Warnf("写数据失败: 连接已关闭或数据为空")
		return
	}

	// 使用 select 实现非阻塞写入,带超时控制
	timeout := s.cnf.GetWriteTimeout()

	select {
	case s.writeChan <- b:
		// 写入成功,更新活跃时间
		s.updateActiveTime()
	case <-time.After(timeout):
		s.log.Errorf("写数据超时")
	case <-s.ctx.Done():
		s.log.Warnf("连接已取消,写数据失败")
	}
}

// write 写数据循环
func (s *Conn[T]) write() {
	for {
		select {
		case msg, ok := <-s.writeChan:
			// 如果通道关闭则退出
			if !ok {
				s.log.Debugf("写通道已关闭,退出写循环")
				return
			}

			// 如果没有写处理函数则跳过
			if s.cnf.OnWrite == nil {
				s.log.Warnf("OnWrite 回调函数未设置")
				continue
			}

			// 写数据
			if err := s.cnf.OnWrite(s.conn, msg); err != nil {
				s.log.Errorf("写数据错误: %v", err)
				assert.ShouldCall0E(s.Close, "conn 关闭错误")
				return
			}

			// 更新活跃时间
			s.updateActiveTime()

		case <-s.ctx.Done():
			s.log.Debugf("上下文已取消,退出写循环")
			return
		}
	}
}

// read 读数据循环
func (s *Conn[T]) read() {
	// 如果没有读处理函数则退出
	if s.cnf.OnRead == nil {
		s.log.Warnf("OnRead 回调函数未设置,退出读循环")
		return
	}

	for {
		// 检查连接是否已关闭
		s.lock.RLock()
		closed := s.closed
		s.lock.RUnlock()

		if closed {
			s.log.Debugf("连接已关闭,退出读循环")
			return
		}

		// 检查上下文是否已取消
		select {
		case <-s.ctx.Done():
			s.log.Debugf("上下文已取消,退出读循环")
			return
		default:
		}

		// 读数据
		if _, bs, err := s.cnf.OnRead(s.conn); err != nil {
			s.log.Warnf("读数据错误: %v", err)
			assert.ShouldCall0E(s.Close, "conn 关闭错误")
			return
		} else {
			// 更新活跃时间
			s.updateActiveTime()

			// 如果没有数据处理函数则退出
			if s.cnf.OnData == nil {
				return
			}

			// 调用数据处理函数
			assert.ShouldCall2E(s.cnf.OnData, IConn(s), bs, "处理数据错误")
		}
	}
}

// Close 关闭连接
func (s *Conn[T]) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 如果已经关闭则退出
	if s.closed {
		return nil
	}

	s.log.Debugf("关闭连接: id=%d", s.cnf.Id)

	// 标记为已关闭
	s.closed = true

	// 取消上下文,通知所有 goroutine 退出
	if s.cancel != nil {
		s.cancel()
	}

	// 关闭写通道
	close(s.writeChan)

	// 调用关闭处理函数
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

// updateActiveTime 更新最后活跃时间
func (s *Conn[T]) updateActiveTime() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.lastActive = time.Now()
}

// checkIdle 检查空闲超时
func (s *Conn[T]) checkIdle() {
	ticker := time.NewTicker(time.Second * 10) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.lock.RLock()
			lastActive := s.lastActive
			idleTimeout := s.cnf.IdleTimeOut
			closed := s.closed
			s.lock.RUnlock()

			if closed {
				return
			}

			// 检查是否超时
			if time.Since(lastActive) > idleTimeout {
				s.log.Warnf("连接空闲超时: id=%d, 空闲时间=%v", s.cnf.Id, time.Since(lastActive))
				assert.ShouldCall0E(s.Close, "conn 关闭错误")
				return
			}

		case <-s.ctx.Done():
			return
		}
	}
}

// IsClosed 检查连接是否已关闭
func (s *Conn[T]) IsClosed() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.closed
}

// GetCreateTime 获取创建时间
func (s *Conn[T]) GetCreateTime() time.Time {
	return s.createAt
}

// GetLastActiveTime 获取最后活跃时间
func (s *Conn[T]) GetLastActiveTime() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.lastActive
}
