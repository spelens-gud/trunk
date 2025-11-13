package conn

import (
	"sync"
	"testing"
	"time"
)

// mockConnForManager 模拟连接对象，用于管理器测试
type mockConnForManager struct {
	id             uint64
	closed         bool
	writtenData    [][]byte
	createTime     time.Time
	lastActiveTime time.Time
	mu             sync.Mutex
}

// newMockConnForManager 创建模拟连接
func newMockConnForManager(id uint64) *mockConnForManager {
	now := time.Now()
	return &mockConnForManager{
		id:             id,
		closed:         false,
		writtenData:    make([][]byte, 0),
		createTime:     now,
		lastActiveTime: now,
	}
}

func (m *mockConnForManager) Start() {}
func (m *mockConnForManager) Write(b []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writtenData = append(m.writtenData, b)
	m.lastActiveTime = time.Now()
}
func (m *mockConnForManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
func (m *mockConnForManager) SetId(id uint64) { m.id = id }
func (m *mockConnForManager) GetId() uint64   { return m.id }
func (m *mockConnForManager) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}
func (m *mockConnForManager) GetCreateTime() time.Time     { return m.createTime }
func (m *mockConnForManager) GetLastActiveTime() time.Time { return m.lastActiveTime }

// TestNewConnectionManager 测试创建连接管理器
func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager()
	if cm == nil {
		t.Fatal("创建连接管理器失败")
	}
	if cm.Count() != 0 {
		t.Errorf("初始连接数应该为 0，实际为 %d", cm.Count())
	}
}

// TestConnectionManager_AddConnection 测试添加连接
func TestConnectionManager_AddConnection(t *testing.T) {
	cm := NewConnectionManager()
	conn := newMockConnForManager(1)

	cm.AddConnection(1, conn)

	if cm.Count() != 1 {
		t.Errorf("添加连接后，连接数应该为 1，实际为 %d", cm.Count())
	}

	// 测试获取连接
	c, exists := cm.GetConnection(1)
	if !exists {
		t.Error("应该能够获取到已添加的连接")
	}
	if c.GetId() != 1 {
		t.Errorf("连接 ID 应该为 1，实际为 %d", c.GetId())
	}
}

// TestConnectionManager_RemoveConnection 测试移除连接
func TestConnectionManager_RemoveConnection(t *testing.T) {
	cm := NewConnectionManager()
	conn := newMockConnForManager(1)

	cm.AddConnection(1, conn)
	cm.RemoveConnection(1)

	if cm.Count() != 0 {
		t.Errorf("移除连接后，连接数应该为 0，实际为 %d", cm.Count())
	}

	_, exists := cm.GetConnection(1)
	if exists {
		t.Error("移除后不应该能够获取到连接")
	}
}

// TestConnectionManager_GetAllConnections 测试获取所有连接
func TestConnectionManager_GetAllConnections(t *testing.T) {
	cm := NewConnectionManager()

	// 添加多个连接
	for i := uint64(1); i <= 5; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	conns := cm.GetAllConnections()
	if len(conns) != 5 {
		t.Errorf("应该有 5 个连接，实际为 %d", len(conns))
	}
}

// TestConnectionManager_Broadcast 测试广播消息
func TestConnectionManager_Broadcast(t *testing.T) {
	cm := NewConnectionManager()

	// 添加多个连接
	conns := make([]*mockConnForManager, 3)
	for i := uint64(0); i < 3; i++ {
		conns[i] = newMockConnForManager(i + 1)
		cm.AddConnection(i+1, conns[i])
	}

	// 广播消息
	testData := []byte("test message")
	cm.Broadcast(testData)

	// 验证每个连接都收到了消息
	for i, conn := range conns {
		conn.mu.Lock()
		if len(conn.writtenData) != 1 {
			t.Errorf("连接 %d 应该收到 1 条消息，实际收到 %d 条", i+1, len(conn.writtenData))
		}
		conn.mu.Unlock()
	}

	// 验证统计信息
	stats := cm.GetStats()
	if stats.TotalMessages != 1 {
		t.Errorf("总消息数应该为 1，实际为 %d", stats.TotalMessages)
	}
	if stats.TotalBytes != uint64(len(testData)) {
		t.Errorf("总字节数应该为 %d，实际为 %d", len(testData), stats.TotalBytes)
	}
}

// TestConnectionManager_BroadcastExclude 测试排除广播
func TestConnectionManager_BroadcastExclude(t *testing.T) {
	cm := NewConnectionManager()

	// 添加多个连接
	conns := make([]*mockConnForManager, 3)
	for i := uint64(0); i < 3; i++ {
		conns[i] = newMockConnForManager(i + 1)
		cm.AddConnection(i+1, conns[i])
	}

	// 广播消息，排除连接 2
	testData := []byte("test message")
	cm.BroadcastExclude(testData, 2)

	// 验证连接 1 和 3 收到消息，连接 2 没有收到
	for i, conn := range conns {
		conn.mu.Lock()
		expectedCount := 1
		if i == 1 { // 连接 2 (索引 1)
			expectedCount = 0
		}
		if len(conn.writtenData) != expectedCount {
			t.Errorf("连接 %d 应该收到 %d 条消息，实际收到 %d 条", i+1, expectedCount, len(conn.writtenData))
		}
		conn.mu.Unlock()
	}
}

// TestConnectionManager_SendTo 测试发送消息给指定连接
func TestConnectionManager_SendTo(t *testing.T) {
	cm := NewConnectionManager()
	conn := newMockConnForManager(1)
	cm.AddConnection(1, conn)

	testData := []byte("test message")

	// 发送给存在的连接
	if !cm.SendTo(1, testData) {
		t.Error("发送给存在的连接应该返回 true")
	}

	// 发送给不存在的连接
	if cm.SendTo(999, testData) {
		t.Error("发送给不存在的连接应该返回 false")
	}

	// 验证连接收到消息
	conn.mu.Lock()
	if len(conn.writtenData) != 1 {
		t.Errorf("连接应该收到 1 条消息，实际收到 %d 条", len(conn.writtenData))
	}
	conn.mu.Unlock()
}

// TestConnectionManager_CloseAll 测试关闭所有连接
func TestConnectionManager_CloseAll(t *testing.T) {
	cm := NewConnectionManager()

	// 添加多个连接
	conns := make([]*mockConnForManager, 3)
	for i := uint64(0); i < 3; i++ {
		conns[i] = newMockConnForManager(i + 1)
		cm.AddConnection(i+1, conns[i])
	}

	// 关闭所有连接
	cm.CloseAll()

	// 验证连接数为 0
	if cm.Count() != 0 {
		t.Errorf("关闭所有连接后，连接数应该为 0，实际为 %d", cm.Count())
	}

	// 验证所有连接都被关闭
	for i, conn := range conns {
		if !conn.IsClosed() {
			t.Errorf("连接 %d 应该被关闭", i+1)
		}
	}
}

// TestConnectionManager_Concurrent 测试并发操作
func TestConnectionManager_Concurrent(t *testing.T) {
	cm := NewConnectionManager()

	// 并发添加连接
	var wg sync.WaitGroup
	for i := uint64(0); i < 100; i++ {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			cm.AddConnection(id, newMockConnForManager(id))
		}(i)
	}
	wg.Wait()

	if cm.Count() != 100 {
		t.Errorf("并发添加后应该有 100 个连接，实际为 %d", cm.Count())
	}

	// 并发读取
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cm.GetAllConnections()
		}()
	}
	wg.Wait()

	// 并发移除
	for i := uint64(0); i < 50; i++ {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			cm.RemoveConnection(id)
		}(i)
	}
	wg.Wait()

	if cm.Count() != 50 {
		t.Errorf("并发移除后应该有 50 个连接，实际为 %d", cm.Count())
	}
}

// TestNewHeartbeatManager 测试创建心跳管理器
func TestNewHeartbeatManager(t *testing.T) {
	timeout := 5 * time.Second
	onTimeout := func(id uint64) {
		// 超时回调
	}

	hm := NewHeartbeatManager(timeout, onTimeout)
	if hm == nil {
		t.Fatal("创建心跳管理器失败")
	}
	if hm.timeout != timeout {
		t.Errorf("超时时间应该为 %v，实际为 %v", timeout, hm.timeout)
	}
}

// TestHeartbeatManager_UpdateHeartbeat 测试更新心跳
func TestHeartbeatManager_UpdateHeartbeat(t *testing.T) {
	hm := NewHeartbeatManager(5*time.Second, nil)

	// 更新心跳
	hm.UpdateHeartbeat(1)

	// 验证心跳信息存在
	hm.lock.RLock()
	info, exists := hm.connections[1]
	hm.lock.RUnlock()

	if !exists {
		t.Error("更新心跳后应该存在心跳信息")
	}
	if info.MissedCount != 0 {
		t.Errorf("初始丢失次数应该为 0，实际为 %d", info.MissedCount)
	}
}

// TestHeartbeatManager_RemoveConnection 测试移除连接
func TestHeartbeatManager_RemoveConnection(t *testing.T) {
	hm := NewHeartbeatManager(5*time.Second, nil)

	hm.UpdateHeartbeat(1)
	hm.RemoveConnection(1)

	hm.lock.RLock()
	_, exists := hm.connections[1]
	hm.lock.RUnlock()

	if exists {
		t.Error("移除后不应该存在心跳信息")
	}
}

// TestHeartbeatManager_CheckTimeouts 测试检查超时
func TestHeartbeatManager_CheckTimeouts(t *testing.T) {
	timeoutCalled := false
	var timeoutID uint64

	hm := NewHeartbeatManager(100*time.Millisecond, func(id uint64) {
		timeoutCalled = true
		timeoutID = id
	})

	// 添加一个连接
	hm.UpdateHeartbeat(1)

	// 等待超时
	time.Sleep(150 * time.Millisecond)

	// 检查超时
	timeouts := hm.CheckTimeouts()

	if len(timeouts) != 1 {
		t.Errorf("应该有 1 个超时连接，实际为 %d", len(timeouts))
	}
	if timeouts[0] != 1 {
		t.Errorf("超时连接 ID 应该为 1，实际为 %d", timeouts[0])
	}

	// 等待回调执行
	time.Sleep(50 * time.Millisecond)

	if !timeoutCalled {
		t.Error("超时回调应该被调用")
	}
	if timeoutID != 1 {
		t.Errorf("超时回调的连接 ID 应该为 1，实际为 %d", timeoutID)
	}
}

// TestHeartbeatManager_StartMonitoring 测试启动监控
func TestHeartbeatManager_StartMonitoring(t *testing.T) {
	checkCount := 0
	var mu sync.Mutex

	hm := NewHeartbeatManager(50*time.Millisecond, func(id uint64) {
		mu.Lock()
		checkCount++
		mu.Unlock()
	})

	// 添加一个连接
	hm.UpdateHeartbeat(1)

	// 启动监控
	stopChan := hm.StartMonitoring(100 * time.Millisecond)

	// 等待几次检查
	time.Sleep(350 * time.Millisecond)

	// 停止监控
	close(stopChan)

	mu.Lock()
	count := checkCount
	mu.Unlock()

	// 应该至少检查到 2-3 次超时
	if count < 2 {
		t.Errorf("应该至少检查到 2 次超时，实际为 %d", count)
	}
}

// BenchmarkConnectionManager_AddConnection 基准测试：添加连接
func BenchmarkConnectionManager_AddConnection(b *testing.B) {
	cm := NewConnectionManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.AddConnection(uint64(i), newMockConnForManager(uint64(i)))
	}
}

// BenchmarkConnectionManager_GetConnection 基准测试：获取连接
func BenchmarkConnectionManager_GetConnection(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加 1000 个连接
	for i := uint64(0); i < 1000; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.GetConnection(uint64(i % 1000))
	}
}

// BenchmarkConnectionManager_RemoveConnection 基准测试：移除连接
func BenchmarkConnectionManager_RemoveConnection(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加足够多的连接
	for i := uint64(0); i < uint64(b.N); i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.RemoveConnection(uint64(i))
	}
}

// BenchmarkConnectionManager_Broadcast 基准测试：广播消息
func BenchmarkConnectionManager_Broadcast(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加 100 个连接
	for i := uint64(0); i < 100; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	testData := []byte("benchmark test message")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.Broadcast(testData)
	}
}

// BenchmarkConnectionManager_BroadcastExclude 基准测试：排除广播
func BenchmarkConnectionManager_BroadcastExclude(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加 100 个连接
	for i := uint64(0); i < 100; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	testData := []byte("benchmark test message")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.BroadcastExclude(testData, 50)
	}
}

// BenchmarkConnectionManager_SendTo 基准测试：发送消息给指定连接
func BenchmarkConnectionManager_SendTo(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加 1000 个连接
	for i := uint64(0); i < 1000; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	testData := []byte("benchmark test message")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.SendTo(uint64(i%1000), testData)
	}
}

// BenchmarkConnectionManager_GetAllConnections 基准测试：获取所有连接
func BenchmarkConnectionManager_GetAllConnections(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加 1000 个连接
	for i := uint64(0); i < 1000; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.GetAllConnections()
	}
}

// BenchmarkConnectionManager_Concurrent 基准测试：并发操作
func BenchmarkConnectionManager_Concurrent(b *testing.B) {
	cm := NewConnectionManager()

	// 预先添加 1000 个连接
	for i := uint64(0); i < 1000; i++ {
		cm.AddConnection(i, newMockConnForManager(i))
	}

	testData := []byte("benchmark test message")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := uint64(0)
		for pb.Next() {
			switch i % 4 {
			case 0:
				cm.GetConnection(i % 1000)
			case 1:
				cm.SendTo(i%1000, testData)
			case 2:
				cm.GetAllConnections()
			case 3:
				cm.Broadcast(testData)
			}
			i++
		}
	})
}

// BenchmarkHeartbeatManager_UpdateHeartbeat 基准测试：更新心跳
func BenchmarkHeartbeatManager_UpdateHeartbeat(b *testing.B) {
	hm := NewHeartbeatManager(5*time.Second, nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hm.UpdateHeartbeat(uint64(i % 1000))
	}
}

// BenchmarkHeartbeatManager_CheckTimeouts 基准测试：检查超时
func BenchmarkHeartbeatManager_CheckTimeouts(b *testing.B) {
	hm := NewHeartbeatManager(5*time.Second, nil)

	// 预先添加 1000 个连接
	for i := uint64(0); i < 1000; i++ {
		hm.UpdateHeartbeat(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hm.CheckTimeouts()
	}
}
