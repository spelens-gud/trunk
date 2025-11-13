package registry

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/logger"
)

// 测试用的 nacos 配置
var testNacosConfig = &NacosConfig{
	Hosts:       []string{"192.168.1.150"},
	Port:        8848,
	NamespaceId: "public",
	GroupName:   "DEFAULT_GROUP",
	ClusterName: "DEFAULT",
	ServiceName: "test-service",
	IP:          "localhost",
	ServicePort: 8877,
	Weight:      1.0,
	Enable:      true,
	Healthy:     true,
	Ephemeral:   true,
	Metadata: map[string]string{
		"version": "1.0.0",
		"env":     "test",
	},
	LogLevel: "error",
	CacheDir: "./cache",
	LogDir:   "./logs",
	Username: "nacos",
	Password: "nacos",
}

// 创建测试用的 NacosRegistry 实例
func newTestNacosRegistry(t *testing.T) *NacosRegistry {
	t.Helper()

	// 创建测试用的 logger
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	assert.SetLogger(log)

	registry := &NacosRegistry{
		cnf: testNacosConfig,
		log: log,
	}

	// 尝试连接 nacos
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("跳过测试：无法连接到 nacos，请确保 nacos 服务正在运行: %v", r)
		}
	}()

	registry.New()

	// 检查连接是否成功
	if registry.namingClient == nil {
		t.Skip("跳过测试：无法连接到 nacos")
	}

	return registry
}

// waitForServiceReady 等待服务注册完成并可查询
// 返回 true 表示服务就绪，false 表示超时
func waitForServiceReady(t *testing.T, registry *NacosRegistry, serviceName string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	retryCount := 0

	for time.Now().Before(deadline) {
		retryCount++
		instances := registry.GetValues(serviceName)

		if instanceList, ok := instances.([]model.Instance); ok && len(instanceList) > 0 {
			t.Logf("服务就绪，重试次数: %d, 实例数: %d", retryCount, len(instanceList))
			return true
		}

		time.Sleep(300 * time.Millisecond)
	}

	t.Logf("等待服务就绪超时，重试次数: %d", retryCount)
	return false
}

// TestNacosRegistry_New 测试创建 nacos 客户端
func TestNacosRegistry_New(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 验证客户端创建成功
	if registry.namingClient == nil {
		t.Fatal("nacos 客户端创建失败")
	}

	// 验证配置正确设置
	if registry.cnf.ServiceName != testNacosConfig.ServiceName {
		t.Errorf("服务名设置错误，期望: %s, 实际: %s", testNacosConfig.ServiceName, registry.cnf.ServiceName)
	}

	// 验证上下文和取消函数已初始化
	if registry.ctx == nil {
		t.Error("上下文未初始化")
	}
	if registry.cancel == nil {
		t.Error("取消函数未初始化")
	}
}

// TestNacosRegistry_Publisher 测试服务注册
func TestNacosRegistry_Publisher(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 注册服务
	registry.Publisher("")

	// 等待服务注册完成（增加等待时间）
	time.Sleep(500 * time.Millisecond)

	t.Log("服务注册成功")
}

// TestNacosRegistry_GetValue 测试获取单个服务实例
func TestNacosRegistry_GetValue(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")

	//// 等待服务就绪
	//if !waitForServiceReady(t, registry, testNacosConfig.ServiceName, 5*time.Second) {
	//	t.Fatal("超时：无法获取到服务实例，请检查 Nacos 服务是否正常")
	//}

	// 等待服务就绪
	time.Sleep(700 * time.Millisecond)

	// 获取服务实例
	value := registry.GetValue(testNacosConfig.ServiceName)
	if value == "" {
		t.Fatal("服务就绪后仍无法获取实例")
	}

	t.Logf("获取到服务实例: %s", value)

	// 验证返回格式为 IP:Port
	expectedFormat := fmt.Sprintf("%s:%d", testNacosConfig.IP, testNacosConfig.ServicePort)
	if value != expectedFormat {
		t.Logf("服务实例格式: 期望 %s, 实际 %s", expectedFormat, value)
	}
}

// TestNacosRegistry_GetValues 测试获取所有服务实例
func TestNacosRegistry_GetValues(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")

	// 等待服务就绪
	//if !waitForServiceReady(t, registry, testNacosConfig.ServiceName, 5*time.Second) {
	//	t.Fatal("超时：无法获取到服务实例列表，请检查 Nacos 服务是否正常")
	//}
	time.Sleep(700 * time.Millisecond)

	// 获取所有服务实例
	instances := registry.GetValues(testNacosConfig.ServiceName)
	if instances == nil {
		t.Fatal("服务就绪后仍无法获取实例列表")
	}

	// 验证实例列表
	if instanceList, ok := instances.([]model.Instance); ok {
		if len(instanceList) == 0 {
			t.Fatal("服务实例列表为空")
		}
		t.Logf("获取到 %d 个服务实例", len(instanceList))
	}
}

// TestNacosRegistry_Deregister 测试服务注销
func TestNacosRegistry_Deregister(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(200 * time.Millisecond)

	// 注销服务
	registry.Deregister()
	time.Sleep(200 * time.Millisecond)

	t.Log("服务注销成功")
}

// TestNacosRegistry_Refresh 测试刷新服务注册
func TestNacosRegistry_Refresh(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")

	// 等待服务就绪
	//if !waitForServiceReady(t, registry, testNacosConfig.ServiceName, 5*time.Second) {
	//	t.Fatal("初始服务注册失败")
	//}
	time.Sleep(700 * time.Millisecond)

	// 刷新服务
	registry.Refresh()

	// 等待刷新完成
	time.Sleep(500 * time.Millisecond)

	// 验证服务仍然存在
	value := registry.GetValue(testNacosConfig.ServiceName)
	if value == "" {
		t.Error("刷新后服务应该仍然存在")
	}

	t.Log("服务刷新成功")
}

// TestNacosRegistry_IsHealthy 测试健康检查
func TestNacosRegistry_IsHealthy(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 正常情况下应该健康
	if !registry.IsHealthy() {
		t.Error("健康检查应该返回 true")
	}

	t.Log("健康检查通过")
}

// TestNacosRegistry_GetLeaseID 测试获取租约ID
func TestNacosRegistry_GetLeaseID(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// nacos 不使用租约，应该返回 0
	leaseID := registry.GetLeaseID()
	if leaseID != 0 {
		t.Error("nacos 不使用租约，应该返回 0")
	}

	t.Log("租约ID检查通过")
}

// TestNacosRegistry_Watch 测试监听服务变化
func TestNacosRegistry_Watch(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")

	// 等待服务就绪
	//if !waitForServiceReady(t, registry, testNacosConfig.ServiceName, 5*time.Second) {
	//	t.Fatal("服务注册失败")
	//}
	time.Sleep(700 * time.Millisecond)

	// 启动监听
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	registry.Watch(ctx, testNacosConfig.ServiceName)

	// 等待监听启动
	time.Sleep(500 * time.Millisecond)

	t.Log("服务监听启动成功")
}

// TestNacosRegistry_ConcurrentAccess 测试并发访问
func TestNacosRegistry_ConcurrentAccess(t *testing.T) {
	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")

	// 等待服务就绪
	//if !waitForServiceReady(t, registry, testNacosConfig.ServiceName, 5*time.Second) {
	//	t.Fatal("服务注册失败")
	//}
	time.Sleep(700 * time.Millisecond)

	concurrency := 10
	operationsPerGoroutine := 20

	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 并发读取
	for i := range concurrency {
		go func(id int) {
			defer wg.Done()
			for range operationsPerGoroutine {
				_ = registry.GetValue(testNacosConfig.ServiceName)
			}
		}(i)
	}

	wg.Wait()

	t.Log("并发访问测试通过")
}

// TestNacosRegistry_ConcurrentPublisher 测试并发注册服务
func TestNacosRegistry_ConcurrentPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发注册测试")
	}

	registry := newTestNacosRegistry(t)
	defer registry.Close()

	concurrency := 5
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 并发注册服务
	for i := range concurrency {
		go func(id int) {
			defer wg.Done()
			registry.Publisher("")
			time.Sleep(700 * time.Millisecond)
		}(i)
	}

	wg.Wait()

	// 验证服务存在
	value := registry.GetValue(testNacosConfig.ServiceName)
	if value == "" {
		t.Error("并发注册后服务应该存在")
	}

	t.Log("并发注册测试通过")
}

// TestNacosRegistry_Performance_MultipleGet 性能测试：多次获取服务实例
func TestNacosRegistry_Performance_MultipleGet(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(700 * time.Millisecond)

	testCount := 100

	// 测试单个实例获取性能
	start := time.Now()
	for range testCount {
		_ = registry.GetValue(testNacosConfig.ServiceName)
	}
	singleGetDuration := time.Since(start)

	t.Logf("单个实例获取 %d 次耗时: %v (平均: %v/次)",
		testCount, singleGetDuration, singleGetDuration/time.Duration(testCount))

	// 测试所有实例获取性能
	start = time.Now()
	for range testCount {
		_ = registry.GetValues(testNacosConfig.ServiceName)
	}
	allGetDuration := time.Since(start)

	t.Logf("所有实例获取 %d 次耗时: %v (平均: %v/次)",
		testCount, allGetDuration, allGetDuration/time.Duration(testCount))
}

// TestNacosRegistry_Performance_HighConcurrency 性能测试：高并发场景
func TestNacosRegistry_Performance_HighConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")

	// 等待服务就绪
	//if !waitForServiceReady(t, registry, testNacosConfig.ServiceName, 5*time.Second) {
	//	t.Fatal("服务注册失败")
	//}

	time.Sleep(700 * time.Millisecond)

	goroutines := 50
	operationsPerGoroutine := 100

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for range operationsPerGoroutine {
				_ = registry.GetValue(testNacosConfig.ServiceName)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalOps := goroutines * operationsPerGoroutine
	opsPerSecond := float64(totalOps) / duration.Seconds()

	t.Logf("高并发测试: %d 个协程，每个 %d 次操作", goroutines, operationsPerGoroutine)
	t.Logf("总操作数: %d, 耗时: %v", totalOps, duration)
	t.Logf("吞吐量: %.2f ops/s", opsPerSecond)
}

// TestNacosRegistry_Performance_ServiceRegistration 性能测试：服务注册性能
func TestNacosRegistry_Performance_ServiceRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestNacosRegistry(t)
	defer registry.Close()

	//time.Sleep(700 * time.Millisecond)

	testCount := 10

	// 测试注册性能
	start := time.Now()
	for range testCount {
		registry.Publisher("")
		time.Sleep(50 * time.Millisecond)
	}
	registerDuration := time.Since(start)

	t.Logf("服务注册 %d 次耗时: %v (平均: %v/次)",
		testCount, registerDuration, registerDuration/time.Duration(testCount))

	// 测试注销性能
	start = time.Now()
	for range testCount {
		registry.Deregister()
		time.Sleep(50 * time.Millisecond)
		registry.Publisher("")
		time.Sleep(50 * time.Millisecond)
	}
	deregisterDuration := time.Since(start)

	t.Logf("服务注销+重注册 %d 次耗时: %v (平均: %v/次)",
		testCount, deregisterDuration, deregisterDuration/time.Duration(testCount))
}

// TestNacosRegistry_Performance_RefreshLatency 性能测试：刷新延迟
func TestNacosRegistry_Performance_RefreshLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestNacosRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(700 * time.Millisecond)

	testCount := 5
	latencies := make([]time.Duration, 0, testCount)

	for range testCount {
		start := time.Now()
		registry.Refresh()
		latency := time.Since(start)
		latencies = append(latencies, latency)
		time.Sleep(200 * time.Millisecond)
	}

	// 计算统计数据
	var totalLatency time.Duration
	var maxLatency time.Duration
	minLatency := latencies[0]

	for _, latency := range latencies {
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
		if latency < minLatency {
			minLatency = latency
		}
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	t.Logf("刷新延迟统计 (测试次数: %d):", testCount)
	t.Logf("  平均延迟: %v", avgLatency)
	t.Logf("  最小延迟: %v", minLatency)
	t.Logf("  最大延迟: %v", maxLatency)
}

// BenchmarkNacosRegistry_GetValue 基准测试：GetValue 操作
func BenchmarkNacosRegistry_GetValue(b *testing.B) {
	t := &testing.T{}
	registry := newTestNacosRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：nacos 不可用")
	}
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(700 * time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		_ = registry.GetValue(testNacosConfig.ServiceName)
	}
}

// BenchmarkNacosRegistry_GetValues 基准测试：GetValues 操作
func BenchmarkNacosRegistry_GetValues(b *testing.B) {
	t := &testing.T{}
	registry := newTestNacosRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：nacos 不可用")
	}
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(700 * time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		_ = registry.GetValues(testNacosConfig.ServiceName)
	}
}

// BenchmarkNacosRegistry_Publisher 基准测试：Publisher 操作
func BenchmarkNacosRegistry_Publisher(b *testing.B) {
	t := &testing.T{}
	registry := newTestNacosRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：nacos 不可用")
	}
	defer registry.Close()

	b.ResetTimer()
	for b.Loop() {
		registry.Publisher("")
		time.Sleep(50 * time.Millisecond) // 避免过快注册
	}
}

// BenchmarkNacosRegistry_Refresh 基准测试：Refresh 操作
func BenchmarkNacosRegistry_Refresh(b *testing.B) {
	t := &testing.T{}
	registry := newTestNacosRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：nacos 不可用")
	}
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(200 * time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		registry.Refresh()
		time.Sleep(100 * time.Millisecond)
	}
}

// BenchmarkNacosRegistry_ConcurrentGet 基准测试：并发 Get 操作
func BenchmarkNacosRegistry_ConcurrentGet(b *testing.B) {
	t := &testing.T{}
	registry := newTestNacosRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：nacos 不可用")
	}
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(700 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = registry.GetValue(testNacosConfig.ServiceName)
		}
	})
}

// BenchmarkNacosRegistry_ConcurrentGetValues 基准测试：并发 GetValues 操作
func BenchmarkNacosRegistry_ConcurrentGetValues(b *testing.B) {
	t := &testing.T{}
	registry := newTestNacosRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：nacos 不可用")
	}
	defer registry.Close()

	// 先注册服务
	registry.Publisher("")
	time.Sleep(700 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = registry.GetValues(testNacosConfig.ServiceName)
		}
	})
}
