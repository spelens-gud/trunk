package conn

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
)

// mockConn 模拟连接
type mockConn struct {
	data       []byte
	writeCount atomic.Int32
	readCount  atomic.Int32
	closeCount atomic.Int32
	closed     bool
	mu         sync.Mutex
}

// newMockConn 创建模拟连接
func newMockConn() *mockConn {
	return &mockConn{
		data: make([]byte, 0),
	}
}

// TestNewConn 测试创建连接
func TestNewConn(t *testing.T) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id:   1,
		Name: "test",
		OnWrite: func(conn *mockConn, raw []byte) error {
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	if conn == nil {
		t.Fatal("创建连接失败")
	}

	if conn.GetId() != 1 {
		t.Errorf("连接ID错误: 期望=1, 实际=%d", conn.GetId())
	}

	if conn.IsClosed() {
		t.Error("新创建的连接不应该是关闭状态")
	}

	if conn.GetCreateTime().IsZero() {
		t.Error("创建时间不应该为零值")
	}
}

// TestConn_SetId 测试设置连接ID
func TestConn_SetId(t *testing.T) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())

	// 设置新ID
	conn.SetId(100)
	if conn.GetId() != 100 {
		t.Errorf("设置ID失败: 期望=100, 实际=%d", conn.GetId())
	}

	// 关闭连接后设置ID应该无效
	_ = conn.Close()
	conn.SetId(200)
	if conn.GetId() != 100 {
		t.Errorf("关闭后设置ID应该无效: 期望=100, 实际=%d", conn.GetId())
	}
}

// TestConn_Write 测试写数据
func TestConn_Write(t *testing.T) {
	mc := newMockConn()
	var writeData []byte
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			writeData = raw
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
		WriteTimeout: time.Second,
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	// 写入数据
	testData := []byte("test data")
	conn.Write(testData)

	// 等待写入完成
	time.Sleep(time.Millisecond * 100)

	if mc.writeCount.Load() != 1 {
		t.Errorf("写入次数错误: 期望=1, 实际=%d", mc.writeCount.Load())
	}

	if string(writeData) != string(testData) {
		t.Errorf("写入数据错误: 期望=%s, 实际=%s", testData, writeData)
	}

	_ = conn.Close()
}

// TestConn_Close 测试关闭连接
func TestConn_Close(t *testing.T) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
		OnClose: func(conn *mockConn) error {
			conn.closeCount.Add(1)
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	// 关闭连接
	err := conn.Close()
	if err != nil {
		t.Errorf("关闭连接失败: %v", err)
	}

	if !conn.IsClosed() {
		t.Error("连接应该是关闭状态")
	}

	if mc.closeCount.Load() != 1 {
		t.Errorf("OnClose调用次数错误: 期望=1, 实际=%d", mc.closeCount.Load())
	}

	// 重复关闭应该不会报错
	err = conn.Close()
	if err != nil {
		t.Errorf("重复关闭不应该报错: %v", err)
	}

	if mc.closeCount.Load() != 1 {
		t.Errorf("重复关闭不应该再次调用OnClose: 实际调用次数=%d", mc.closeCount.Load())
	}
}

// TestConn_WriteError 测试写数据错误
func TestConn_WriteError(t *testing.T) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			return errors.New("写入错误")
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
		OnClose: func(conn *mockConn) error {
			conn.closeCount.Add(1)
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	// 写入数据
	conn.Write([]byte("test"))

	// 等待写入和关闭
	time.Sleep(time.Millisecond * 200)

	if mc.writeCount.Load() != 1 {
		t.Errorf("写入次数错误: 期望=1, 实际=%d", mc.writeCount.Load())
	}

	if !conn.IsClosed() {
		t.Error("写入错误后连接应该被关闭")
	}
}

// TestConn_ConcurrentWrite 测试并发写入
func TestConn_ConcurrentWrite(t *testing.T) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			time.Sleep(time.Millisecond * 10)
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	// 并发写入
	goroutines := 10
	wg := sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			conn.Write([]byte("test"))
		}(i)
	}

	wg.Wait()
	time.Sleep(time.Millisecond * 500)

	if mc.writeCount.Load() != int32(goroutines) {
		t.Errorf("并发写入次数错误: 期望=%d, 实际=%d", goroutines, mc.writeCount.Load())
	}

	_ = conn.Close()
}

// BenchmarkConn_Write 基准测试：写数据
func BenchmarkConn_Write(b *testing.B) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.Write(testData)
	}
	b.StopTimer()

	_ = conn.Close()
}

// BenchmarkConn_WriteParallel 基准测试：并发写数据
func BenchmarkConn_WriteParallel(b *testing.B) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	testData := []byte("benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn.Write(testData)
		}
	})
	b.StopTimer()

	_ = conn.Close()
}

// BenchmarkConn_NewConn 基准测试：创建连接
func BenchmarkConn_NewConn(b *testing.B) {
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc := newMockConn()
		conn := NewConn(mc, cfg)
		_ = conn
	}
}

// BenchmarkConn_GetId 基准测试：获取连接ID
func BenchmarkConn_GetId(b *testing.B) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conn.GetId()
	}
}

// BenchmarkConn_IsClosed 基准测试：检查连接状态
func BenchmarkConn_IsClosed(b *testing.B) {
	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conn.IsClosed()
	}
}

// TestConn_Performance_HighThroughput 性能测试：高吞吐量写入
func TestConn_Performance_HighThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	// 高吞吐量测试：10000次写入
	count := 10000
	start := time.Now()

	for i := 0; i < count; i++ {
		conn.Write([]byte("performance test"))
	}

	// 等待所有写入完成
	time.Sleep(time.Second * 2)

	elapsed := time.Since(start)
	throughput := float64(mc.writeCount.Load()) / elapsed.Seconds()

	t.Logf("写入次数: %d", mc.writeCount.Load())
	t.Logf("耗时: %v", elapsed)
	t.Logf("吞吐量: %.2f ops/s", throughput)

	expectedMin := int32(float64(count) * 0.9)
	if mc.writeCount.Load() < expectedMin {
		t.Errorf("写入成功率过低: 期望>=%d, 实际=%d", expectedMin, mc.writeCount.Load())
	}

	_ = conn.Close()
}

// TestConn_Performance_ConcurrentLoad 性能测试：并发负载
func TestConn_Performance_ConcurrentLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	mc := newMockConn()
	cfg := NetConfig[*mockConn]{
		Id: 1,
		OnWrite: func(conn *mockConn, raw []byte) error {
			conn.writeCount.Add(1)
			time.Sleep(time.Microsecond * 100) // 模拟网络延迟
			return nil
		},
		OnRead: func(conn *mockConn) (int, []byte, error) {
			time.Sleep(time.Millisecond * 100)
			return 0, nil, nil
		},
		OnData: func(conn IConn, raw []byte) error {
			return nil
		},
	}

	conn := NewConn(mc, cfg)
	conn.SetLogger(logger.GetDefault())
	conn.Start()

	// 并发负载测试
	goroutines := 100
	writesPerGoroutine := 100
	start := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				conn.Write([]byte("concurrent test"))
			}
		}()
	}

	wg.Wait()
	time.Sleep(time.Second * 3)

	elapsed := time.Since(start)
	totalWrites := goroutines * writesPerGoroutine
	throughput := float64(mc.writeCount.Load()) / elapsed.Seconds()

	t.Logf("并发数: %d", goroutines)
	t.Logf("总写入次数: %d", totalWrites)
	t.Logf("实际写入次数: %d", mc.writeCount.Load())
	t.Logf("耗时: %v", elapsed)
	t.Logf("吞吐量: %.2f ops/s", throughput)

	expectedMin := int32(float64(totalWrites) * 0.9)
	if mc.writeCount.Load() < expectedMin {
		t.Errorf("并发写入成功率过低: 期望>=%d, 实际=%d", expectedMin, mc.writeCount.Load())
	}

	_ = conn.Close()
}
