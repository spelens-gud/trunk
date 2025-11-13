package conn

import (
	"sync"
	"time"

	"github.com/spelens-gud/assert"
)

// ConnectionManager 连接管理器
type ConnectionManager struct {
	connections   map[uint64]IConn // 连接
	lock          sync.RWMutex     // 锁
	totalMessages uint64           // 消息数
	totalBytes    uint64           // 字节数
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[uint64]IConn),
	}
}

// AddConnection 添加连接
func (cm *ConnectionManager) AddConnection(id uint64, c IConn) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.connections[id] = c
}

// RemoveConnection 移除连接
func (cm *ConnectionManager) RemoveConnection(id uint64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	delete(cm.connections, id)
}

// GetConnection 获取连接
func (cm *ConnectionManager) GetConnection(id uint64) (IConn, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	c, exists := cm.connections[id]
	return c, exists
}

// GetAllConnections 获取所有连接
func (cm *ConnectionManager) GetAllConnections() []IConn {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	conns := make([]IConn, 0, len(cm.connections))
	for _, c := range cm.connections {
		conns = append(conns, c)
	}
	return conns
}

// Count 获取连接数
func (cm *ConnectionManager) Count() int {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	return len(cm.connections)
}

// Broadcast 广播消息
func (cm *ConnectionManager) Broadcast(data []byte) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	for _, c := range cm.connections {
		c.Write(data)
	}

	cm.totalMessages++
	cm.totalBytes += uint64(len(data))
}

// BroadcastExclude 广播消息(排除指定连接)
func (cm *ConnectionManager) BroadcastExclude(data []byte, excludeID uint64) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	for id, c := range cm.connections {
		if id != excludeID {
			c.Write(data)
		}
	}
}

// SendTo 发送消息给指定连接
func (cm *ConnectionManager) SendTo(id uint64, data []byte) bool {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	if c, exists := cm.connections[id]; exists {
		c.Write(data)
		return true
	}

	return false
}

// CloseAll 关闭所有连接
func (cm *ConnectionManager) CloseAll() {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	for _, c := range cm.connections {
		assert.ShouldCall0E(c.Close, "conn连接关闭失败")
	}
	cm.connections = make(map[uint64]IConn)
}

// GetStats 获取统计信息
func (cm *ConnectionManager) GetStats() ManagerStats {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	return ManagerStats{
		ConnectionCount: len(cm.connections),
		TotalMessages:   cm.totalMessages,
		TotalBytes:      cm.totalBytes,
	}
}

// ManagerStats 管理器统计信息
type ManagerStats struct {
	ConnectionCount int    // 连接数
	TotalMessages   uint64 // 总消息数
	TotalBytes      uint64 // 总字节数
}

// HeartbeatManager 心跳管理器
type HeartbeatManager struct {
	connections map[uint64]*HeartbeatInfo // 连接信息
	lock        sync.RWMutex              // 锁
	timeout     time.Duration             // 超时时间
	onTimeout   func(id uint64)           // 超时回调
}

// HeartbeatInfo 心跳信息
type HeartbeatInfo struct {
	LastHeartbeat time.Time // 最后一次心跳时间
	MissedCount   int       // 丢失次数
}

// NewHeartbeatManager 创建心跳管理器
func NewHeartbeatManager(timeout time.Duration, onTimeout func(id uint64)) *HeartbeatManager {
	return &HeartbeatManager{
		connections: make(map[uint64]*HeartbeatInfo),
		timeout:     timeout,
		onTimeout:   onTimeout,
	}
}

// UpdateHeartbeat 更新心跳
func (hm *HeartbeatManager) UpdateHeartbeat(id uint64) {
	hm.lock.Lock()
	defer hm.lock.Unlock()

	// 如果连接已存在，则更新
	if info, exists := hm.connections[id]; exists {
		info.LastHeartbeat = time.Now()
		info.MissedCount = 0
	} else {
		hm.connections[id] = &HeartbeatInfo{
			LastHeartbeat: time.Now(),
			MissedCount:   0,
		}
	}
}

// RemoveConnection 移除连接
func (hm *HeartbeatManager) RemoveConnection(id uint64) {
	hm.lock.Lock()
	defer hm.lock.Unlock()

	delete(hm.connections, id)
}

// CheckTimeouts 检查超时
func (hm *HeartbeatManager) CheckTimeouts() []uint64 {
	hm.lock.Lock()
	defer hm.lock.Unlock()

	now := time.Now()
	timeouts := make([]uint64, 0)

	// 遍历所有连接
	for id, info := range hm.connections {
		if now.Sub(info.LastHeartbeat) > hm.timeout {
			info.MissedCount++
			timeouts = append(timeouts, id)

			if hm.onTimeout != nil {
				go hm.onTimeout(id)
			}
		}
	}

	return timeouts
}

// StartMonitoring 启动监控
func (hm *HeartbeatManager) StartMonitoring(interval time.Duration) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hm.CheckTimeouts()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}
