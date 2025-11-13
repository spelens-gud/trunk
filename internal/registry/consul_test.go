package registry

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/logger"
)

// 测试用的 consul 配置
var testConsulConfig = &ConsulConfig{
	Address:             "192.168.1.150:8500",
	Scheme:              "http",
	Datacenter:          "dc1",
	ServiceName:         "test-service",
	ServiceAddress:      "localhost",
	ServicePort:         8080,
	ServiceTags:         []string{"test", "v1"},
	ServiceMeta:         map[string]string{"version": "1.0.0"},
	HealthCheckPath:     "/health",
	HealthCheckInterval: "10s",
	HealthCheckTimeout:  "5s",
	DeregisterAfter:     "30s",
	EnableTagOverride:   false,
}

// 创建测试用的 ConsulRegistry 实例
func newTestConsulRegistry(t *testing.T) *ConsulRegistry {
	t.Helper()

	// 创建测试用的 logger
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	assert.SetLogger(log)

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	// 尝试连接 consul
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("跳过测试：无法连接到 consul，请确保 consul 服务正在运行: %v", r)
		}
	}()

	registry.New()

	// 检查连接是否成功
	if registry.client == nil {
		t.Skip("跳过测试：无法连接到 consul")
	}

	return registry
}

// 清理测试数据
func cleanupConsulTestData(t *testing.T, registry *ConsulRegistry, prefix string) {
	t.Helper()
	_, _ = registry.client.KV().DeleteTree(prefix, nil)
}

// TestConsulRegistry_New 测试创建 consul 客户端
func TestConsulRegistry_New(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// 验证客户端创建成功
	if registry.client == nil {
		t.Fatal("consul 客户端创建失败")
	}

	// 验证配置正确设置
	if registry.cnf.Address != testConsulConfig.Address {
		t.Errorf("地址设置错误，期望: %s, 实际: %s", testConsulConfig.Address, registry.cnf.Address)
	}

	// 验证上下文和取消函数已初始化
	if registry.ctx == nil {
		t.Error("上下文未初始化")
	}
	if registry.cancel == nil {
		t.Error("取消函数未初始化")
	}
}

// TestConsulRegistry_Publisher 测试服务发布
func TestConsulRegistry_Publisher(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	testValue := "test-service-value"
	registry.Publisher(testValue)

	// 等待服务注册完成
	time.Sleep(150 * time.Millisecond)

	// 验证服务已注册到 consul
	services, err := registry.client.Agent().Services()
	if err != nil {
		t.Fatalf("获取注册的服务失败: %v", err)
	}

	serviceID := registry.cnf.GetServiceID()
	service, exists := services[serviceID]
	if !exists {
		t.Error("服务未成功注册到 consul")
	} else {
		if service.Service != registry.cnf.ServiceName {
			t.Errorf("注册的服务名不正确，期望: %s, 实际: %s", registry.cnf.ServiceName, service.Service)
		}
		if service.Port != registry.cnf.ServicePort {
			t.Errorf("注册的服务端口不正确，期望: %d, 实际: %d", registry.cnf.ServicePort, service.Port)
		}
	}
}

// TestConsulRegistry_Deregister 测试服务注销
func TestConsulRegistry_Deregister(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// 先注册服务
	testValue := "test-service"
	registry.Publisher(testValue)
	time.Sleep(150 * time.Millisecond)

	serviceID := registry.cnf.GetServiceID()

	// 验证服务已注册
	services, err := registry.client.Agent().Services()
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}
	if _, exists := services[serviceID]; !exists {
		t.Fatal("服务应该已注册")
	}

	// 注销服务
	registry.Deregister()
	time.Sleep(50 * time.Millisecond)

	// 验证服务已被删除
	services, err = registry.client.Agent().Services()
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if _, exists := services[serviceID]; exists {
		t.Error("服务注销后应该被删除")
	}
}

// TestConsulRegistry_PutAndGet 测试键值存取
func TestConsulRegistry_PutAndGet(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()
	defer cleanupConsulTestData(t, registry, "test/key")

	ctx := context.Background()
	testKey := "test/key1"
	testValue := "test-value-1"

	// 测试 Put
	registry.Put(ctx, testKey, testValue)

	// 等待写入完成
	time.Sleep(50 * time.Millisecond)

	// 验证数据已写入
	pair, _, err := registry.client.KV().Get(testKey, nil)
	if err != nil {
		t.Fatalf("获取键值失败: %v", err)
	}

	if pair == nil {
		t.Fatal("键值不存在")
	}

	if string(pair.Value) != testValue {
		t.Errorf("获取的值不正确，期望: %s, 实际: %s", testValue, string(pair.Value))
	}
}

// TestConsulRegistry_GetValue 测试获取单个服务实例
func TestConsulRegistry_GetValue(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("test-service")
	time.Sleep(150 * time.Millisecond)

	// 获取服务实例
	value := registry.GetValue(registry.cnf.ServiceName)
	if value == "" {
		t.Error("应该能获取到服务实例")
	}

	expectedValue := fmt.Sprintf("%s:%d", registry.cnf.ServiceAddress, registry.cnf.ServicePort)
	if value != expectedValue {
		t.Errorf("获取的服务地址不正确，期望: %s, 实际: %s", expectedValue, value)
	}

	// 测试获取不存在的服务
	nonExistentValue := registry.GetValue("non-existent-service")
	if nonExistentValue != "" {
		t.Errorf("获取不存在的服务应返回空字符串，实际: %s", nonExistentValue)
	}
}

// TestConsulRegistry_GetValues 测试获取所有服务实例
func TestConsulRegistry_GetValues(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// 先注册服务
	registry.Publisher("test-service")
	time.Sleep(150 * time.Millisecond)

	// 获取所有服务实例
	result := registry.GetValues(registry.cnf.ServiceName)
	if result == nil {
		t.Error("GetValues 应该返回非空结果")
	}

	services, ok := result.([]*api.ServiceEntry)
	if !ok {
		t.Fatal("返回类型不正确")
	}

	if len(services) == 0 {
		t.Error("应该至少有一个服务实例")
	}
}

// TestConsulRegistry_Watch 测试监听服务变化
func TestConsulRegistry_Watch(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	serviceName := "watch-test-service"

	// 启动监听
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go registry.Watch(ctx, serviceName)

	// 等待监听启动
	time.Sleep(200 * time.Millisecond)

	// 注册一个新服务触发变化
	testConfig := *testConsulConfig
	testConfig.ServiceName = serviceName
	testConfig.ServiceID = serviceName + "-1"

	testRegistry := &ConsulRegistry{
		cnf: &testConfig,
		log: registry.log,
	}
	testRegistry.New()
	testRegistry.Publisher("test")

	// 等待事件处理
	time.Sleep(500 * time.Millisecond)

	// 清理
	testRegistry.Close()
}

// TestConsulRegistry_IsHealthy 测试健康检查
func TestConsulRegistry_IsHealthy(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// 正常情况下应该健康
	if !registry.IsHealthy() {
		t.Error("健康检查应该返回 true")
	}

	// 客户端为 nil 时应该不健康
	registry.client = nil
	if registry.IsHealthy() {
		t.Error("客户端为 nil 时健康检查应该返回 false")
	}
}

// TestConsulRegistry_Refresh 测试刷新服务注册
func TestConsulRegistry_Refresh(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// 先注册服务
	testValue := "test-service"
	registry.Publisher(testValue)
	time.Sleep(150 * time.Millisecond)

	serviceID := registry.cnf.GetServiceID()

	// 验证服务已注册
	services, err := registry.client.Agent().Services()
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}
	if _, exists := services[serviceID]; !exists {
		t.Fatal("服务应该已注册")
	}

	// 刷新服务
	registry.Refresh()
	time.Sleep(150 * time.Millisecond)

	// 验证服务仍然存在
	services, err = registry.client.Agent().Services()
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if _, exists := services[serviceID]; !exists {
		t.Error("刷新后服务应该存在")
	}
}

// TestConsulRegistry_GetLeaseID 测试获取租约ID
func TestConsulRegistry_GetLeaseID(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()

	// consul 不使用租约，应该返回 0
	leaseID := registry.GetLeaseID()
	if leaseID != 0 {
		t.Error("consul 不使用租约，应该返回 0")
	}
}

// TestConsulRegistry_ConcurrentAccess 测试并发访问
func TestConsulRegistry_ConcurrentAccess(t *testing.T) {
	registry := newTestConsulRegistry(t)
	defer registry.Close()
	defer cleanupConsulTestData(t, registry, "test/concurrent/")

	ctx := context.Background()
	prefix := "test/concurrent/"
	concurrency := 10
	operationsPerGoroutine := 50

	var wg sync.WaitGroup
	wg.Add(concurrency * 2) // 写入和读取各一半

	// 并发写入
	for i := range concurrency {
		go func(id int) {
			defer wg.Done()
			for j := range operationsPerGoroutine {
				key := fmt.Sprintf("%skey-%d-%d", prefix, id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				registry.Put(ctx, key, value)
			}
		}(i)
	}

	// 并发读取
	for i := range concurrency {
		go func(id int) {
			defer wg.Done()
			for j := range operationsPerGoroutine {
				key := fmt.Sprintf("%skey-%d-%d", prefix, id, j)
				pair, _, _ := registry.client.KV().Get(key, nil)
				_ = pair
			}
		}(i)
	}

	wg.Wait()

	// 验证数据
	time.Sleep(200 * time.Millisecond)
	pairs, _, err := registry.client.KV().List(prefix, nil)
	if err != nil {
		t.Fatalf("获取键值列表失败: %v", err)
	}

	expectedCount := concurrency * operationsPerGoroutine
	if len(pairs) != expectedCount {
		t.Errorf("并发写入后键值对数量不正确，期望: %d, 实际: %d", expectedCount, len(pairs))
	}
}

// TestConsulRegistry_ConcurrentPublisher 测试并发发布服务
func TestConsulRegistry_ConcurrentPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发发布测试")
	}

	registry := newTestConsulRegistry(t)
	defer registry.Close()

	concurrency := 5
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 并发发布服务
	for i := range concurrency {
		go func(id int) {
			defer wg.Done()
			value := fmt.Sprintf("service-%d", id)
			registry.Publisher(value)
			time.Sleep(50 * time.Millisecond)
		}(i)
	}

	wg.Wait()

	// 验证服务已注册
	services, err := registry.client.Agent().Services()
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	serviceID := registry.cnf.GetServiceID()
	if _, exists := services[serviceID]; !exists {
		t.Error("并发发布后服务应该存在")
	}
}

// ============================================================================
// 基准测试 (Benchmark Tests)
// ============================================================================

// BenchmarkConsulRegistry_Publisher 基准测试服务发布性能
func BenchmarkConsulRegistry_Publisher(b *testing.B) {
	// 创建测试用的 logger
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Publisher(fmt.Sprintf("bench-service-%d", i))
	}
}

// BenchmarkConsulRegistry_GetValue 基准测试获取单个服务实例性能
func BenchmarkConsulRegistry_GetValue(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	// 先注册一个服务
	registry.Publisher("bench-service")
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetValue(registry.cnf.ServiceName)
	}
}

// BenchmarkConsulRegistry_GetValues 基准测试获取所有服务实例性能
func BenchmarkConsulRegistry_GetValues(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	// 先注册一个服务
	registry.Publisher("bench-service")
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetValues(registry.cnf.ServiceName)
	}
}

// BenchmarkConsulRegistry_Put 基准测试键值写入性能
func BenchmarkConsulRegistry_Put(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()
	defer cleanupConsulTestData(&testing.T{}, registry, "bench/")

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench/key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		registry.Put(ctx, key, value)
	}
}

// BenchmarkConsulRegistry_Deregister 基准测试服务注销性能
func BenchmarkConsulRegistry_Deregister(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		registry := &ConsulRegistry{
			cnf: testConsulConfig,
			log: log,
		}
		registry.New()
		registry.Publisher("bench-service")
		time.Sleep(50 * time.Millisecond)
		b.StartTimer()

		registry.Deregister()
	}
}

// BenchmarkConsulRegistry_IsHealthy 基准测试健康检查性能
func BenchmarkConsulRegistry_IsHealthy(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.IsHealthy()
	}
}

// BenchmarkConsulRegistry_ConcurrentPut 基准测试并发键值写入性能
func BenchmarkConsulRegistry_ConcurrentPut(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()
	defer cleanupConsulTestData(&testing.T{}, registry, "bench/concurrent/")

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench/concurrent/key-%d", i)
			value := fmt.Sprintf("value-%d", i)
			registry.Put(ctx, key, value)
			i++
		}
	})
}

// BenchmarkConsulRegistry_ConcurrentGetValue 基准测试并发获取服务实例性能
func BenchmarkConsulRegistry_ConcurrentGetValue(b *testing.B) {
	logConfig := logger.DefaultConfig()
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		b.Fatalf("创建 logger 失败: %v", err)
	}

	registry := &ConsulRegistry{
		cnf: testConsulConfig,
		log: log,
	}

	defer func() {
		if r := recover(); r != nil {
			b.Skipf("跳过基准测试：无法连接到 consul: %v", r)
		}
	}()

	registry.New()
	defer registry.Close()

	if registry.client == nil {
		b.Skip("跳过基准测试：无法连接到 consul")
	}

	// 先注册一个服务
	registry.Publisher("bench-service")
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = registry.GetValue(registry.cnf.ServiceName)
		}
	})
}
