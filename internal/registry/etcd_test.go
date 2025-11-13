package registry

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// 测试用的 etcd 配置
var testEtcdConfig = &EtcdConfig{
	Hosts:       []string{"192.168.1.150:2379"},
	Key:         "/test/service",
	LeaseTTL:    5,
	DialTimeout: 3,
	User:        "root",
	Pass:        "123456",
}

// 创建测试用的 EtcdRegistry 实例
func newTestEtcdRegistry(t *testing.T) *EtcdRegistry {
	t.Helper()

	// 创建测试用的 logger
	logConfig := logger.DefaultConfig()
	// logConfig.Console = false // 减少测试输出噪音
	// logConfig.File = false
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	assert.SetLogger(log)

	ctx, cancel := context.WithCancel(context.Background())

	registry := &EtcdRegistry{
		cnf:    testEtcdConfig.Copy(),
		log:    log,
		ctx:    ctx,
		cancel: cancel,
	}

	// 尝试连接 etcd
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("跳过测试：无法连接到 etcd，请确保 etcd 服务正在运行: %v", r)
		}
	}()

	registry.New()

	// 检查连接是否成功
	if registry.cli == nil {
		t.Skip("跳过测试：无法连接到 etcd")
	}

	return registry
}

// 清理测试数据
func cleanupTestData(t *testing.T, registry *EtcdRegistry, prefix string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = registry.cli.Delete(ctx, prefix, clientv3.WithPrefix())
}

// TestEtcdRegistry_New 测试创建 etcd 客户端
func TestEtcdRegistry_New(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 验证客户端创建成功
	if registry.cli == nil {
		t.Fatal("etcd 客户端创建失败")
	}

	// 验证配置正确设置
	if registry.key != testEtcdConfig.Key {
		t.Errorf("key 设置错误，期望: %s, 实际: %s", testEtcdConfig.Key, registry.key)
	}

	// 验证上下文和取消函数已初始化
	if registry.ctx == nil {
		t.Error("上下文未初始化")
	}
	if registry.cancel == nil {
		t.Error("取消函数未初始化")
	}
}

// TestEtcdRegistry_GetCacheClient 测试获取缓存客户端
func TestEtcdRegistry_GetCacheClient(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	client := registry.GetCacheClient()
	if client == nil {
		t.Fatal("获取缓存客户端失败")
	}

	if client != registry.cli {
		t.Error("返回的客户端与内部客户端不一致")
	}
}

// TestEtcdRegistry_Publisher 测试服务发布
func TestEtcdRegistry_Publisher(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	testValue := "test-service-value"
	registry.Publisher(testValue)

	// 等待服务注册完成
	time.Sleep(150 * time.Millisecond)

	// 验证服务值设置正确
	if registry.val != testValue {
		t.Errorf("服务值设置错误，期望: %s, 实际: %s", testValue, registry.val)
	}

	// 验证租约ID已创建
	if registry.leaseID == 0 {
		t.Error("租约ID应该不为0")
	}

	// 验证 keepAlive 通道已创建
	if registry.keepAliveChan == nil {
		t.Error("keepAlive 通道应该已初始化")
	}

	// 验证服务已注册到 etcd
	ctx := context.Background()
	key := fmt.Sprintf("%s/%d", registry.key, registry.leaseID)
	resp, err := registry.cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("获取注册的服务失败: %v", err)
	}

	if len(resp.Kvs) == 0 {
		t.Error("服务未成功注册到 etcd")
	} else if string(resp.Kvs[0].Value) != testValue {
		t.Errorf("注册的服务值不正确，期望: %s, 实际: %s", testValue, string(resp.Kvs[0].Value))
	}
}

// TestEtcdRegistry_PutAndGet 测试键值存取
func TestEtcdRegistry_PutAndGet(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/test/key")

	ctx := context.Background()
	testKey := "/test/key1"
	testValue := "test-value-1"

	// 测试 Put
	registry.Put(ctx, testKey, testValue)

	// 等待写入完成
	time.Sleep(50 * time.Millisecond)

	// 测试 GetValue
	value := registry.GetValue(testKey)
	if value != testValue {
		t.Errorf("获取的值不正确，期望: %s, 实际: %s", testValue, value)
	}

	// 测试获取不存在的键
	nonExistentValue := registry.GetValue("/test/nonexistent")
	if nonExistentValue != "" {
		t.Errorf("获取不存在的键应返回空字符串，实际: %s", nonExistentValue)
	}
}

// TestEtcdRegistry_GetValueWithOptions 测试带选项的获取值
func TestEtcdRegistry_GetValueWithOptions(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/test/options")

	ctx := context.Background()
	prefix := "/test/options/"

	// 插入多个键值
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("%skey%d", prefix, i)
		value := fmt.Sprintf("value%d", i)
		registry.Put(ctx, key, value)
	}

	time.Sleep(50 * time.Millisecond)

	// 使用 WithPrefix 选项获取第一个值
	value := registry.GetValue(prefix, clientv3.WithPrefix(), clientv3.WithLimit(1))
	if value == "" {
		t.Error("使用 WithPrefix 选项应该能获取到值")
	}
}

// TestEtcdRegistry_GetValues 测试获取多个值
func TestEtcdRegistry_GetValues(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/test/multi/")

	ctx := context.Background()
	prefix := "/test/multi/"

	// 插入多个键值
	testData := map[string]string{
		prefix + "key1": "value1",
		prefix + "key2": "value2",
		prefix + "key3": "value3",
	}

	for k, v := range testData {
		registry.Put(ctx, k, v)
	}

	// 等待写入完成
	time.Sleep(50 * time.Millisecond)

	// 测试 GetValuesTyped
	kvs := registry.GetValuesTyped(prefix, clientv3.WithPrefix())
	if len(kvs) != len(testData) {
		t.Errorf("获取的键值对数量不正确，期望: %d, 实际: %d", len(testData), len(kvs))
	}

	// 验证获取的值正确
	for _, kv := range kvs {
		key := string(kv.Key)
		value := string(kv.Value)
		expectedValue, exists := testData[key]
		if !exists {
			t.Errorf("获取到未预期的键: %s", key)
		} else if value != expectedValue {
			t.Errorf("键 %s 的值不正确，期望: %s, 实际: %s", key, expectedValue, value)
		}
	}

	// 测试 GetValues (返回 any 类型)
	result := registry.GetValues(prefix, clientv3.WithPrefix())
	if result == nil {
		t.Error("GetValues 应该返回非空结果")
	}
}

// TestEtcdRegistry_Watch 测试监听功能
func TestEtcdRegistry_Watch(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/test/watch/")

	watchPrefix := "/test/watch/"
	testKey := watchPrefix + "key1"
	testValue := "watch-value"

	// 创建一个通道来接收事件
	eventReceived := make(chan bool, 1)
	var eventType string

	// 启动监听
	registry.WatchWithCallback(watchPrefix, func(event *clientv3.Event) {
		if string(event.Kv.Key) == testKey {
			eventType = event.Type.String()
			if string(event.Kv.Value) == testValue {
				eventReceived <- true
			}
		}
	})

	// 等待监听启动
	time.Sleep(100 * time.Millisecond)

	// 触发 PUT 事件
	ctx := context.Background()
	registry.Put(ctx, testKey, testValue)

	// 等待事件接收
	select {
	case <-eventReceived:
		t.Logf("成功接收到 Watch 事件，类型: %s", eventType)
	case <-time.After(2 * time.Second):
		t.Error("超时：未接收到 Watch 事件")
	}
}

// TestEtcdRegistry_WatchTyped 测试类型安全的监听
func TestEtcdRegistry_WatchTyped(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/test/watchtyped/")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	watchPrefix := "/test/watchtyped/"
	testKey := watchPrefix + "key1"
	testValue := "typed-watch-value"

	// 启动类型安全的监听
	watchChan := registry.WatchTyped(ctx, watchPrefix)
	if watchChan == nil {
		t.Fatal("WatchTyped 返回 nil")
	}

	// 在另一个 goroutine 中触发事件
	go func() {
		time.Sleep(100 * time.Millisecond)
		registry.Put(context.Background(), testKey, testValue)
	}()

	// 等待事件
	select {
	case watchResp := <-watchChan:
		if watchResp.Err() != nil {
			t.Errorf("Watch 错误: %v", watchResp.Err())
		}
		if len(watchResp.Events) == 0 {
			t.Error("未收到任何事件")
		} else {
			t.Logf("成功接收到 %d 个事件", len(watchResp.Events))
		}
	case <-ctx.Done():
		t.Error("超时：未接收到 Watch 事件")
	}
}

// TestEtcdRegistry_Deregister 测试服务注销
func TestEtcdRegistry_Deregister(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 先注册服务
	testValue := "test-service"
	registry.Publisher(testValue)
	time.Sleep(150 * time.Millisecond)

	leaseID := registry.leaseID
	if leaseID == 0 {
		t.Fatal("租约ID不应该为0")
	}

	// 注销服务
	registry.Deregister()

	// 等待注销完成
	time.Sleep(50 * time.Millisecond)

	// 验证服务已被删除
	ctx := context.Background()
	key := fmt.Sprintf("%s/%d", registry.key, leaseID)
	resp, err := registry.cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("获取键失败: %v", err)
	}

	if len(resp.Kvs) != 0 {
		t.Error("服务注销后，键应该被删除")
	}
}

// TestEtcdRegistry_DeregisterWithoutLease 测试未注册时的注销
func TestEtcdRegistry_DeregisterWithoutLease(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 直接注销（未注册服务）
	registry.Deregister()

	// 应该不会崩溃，只是记录警告
	if registry.leaseID != 0 {
		t.Error("未注册服务时，租约ID应该为0")
	}
}

// TestEtcdRegistry_Refresh 测试刷新服务注册
func TestEtcdRegistry_Refresh(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 先注册服务
	testValue := "test-service"
	registry.Publisher(testValue)
	time.Sleep(150 * time.Millisecond)

	oldLeaseID := registry.leaseID
	if oldLeaseID == 0 {
		t.Fatal("初始租约ID不应该为0")
	}

	// 刷新服务
	registry.Refresh()
	time.Sleep(150 * time.Millisecond)

	newLeaseID := registry.leaseID

	if oldLeaseID == newLeaseID {
		t.Error("刷新后租约ID应该改变")
	}

	if newLeaseID == 0 {
		t.Error("刷新后租约ID不应该为0")
	}

	// 验证新服务已注册
	ctx := context.Background()
	newKey := fmt.Sprintf("%s/%d", registry.key, newLeaseID)
	resp, err := registry.cli.Get(ctx, newKey)
	if err != nil {
		t.Fatalf("获取刷新后的服务失败: %v", err)
	}

	if len(resp.Kvs) == 0 {
		t.Error("刷新后服务应该存在")
	} else if string(resp.Kvs[0].Value) != testValue {
		t.Errorf("刷新后服务值不正确，期望: %s, 实际: %s", testValue, string(resp.Kvs[0].Value))
	}

	// 验证旧服务已删除
	oldKey := fmt.Sprintf("%s/%d", registry.key, oldLeaseID)
	oldResp, err := registry.cli.Get(ctx, oldKey)
	if err != nil {
		t.Fatalf("获取旧服务失败: %v", err)
	}

	if len(oldResp.Kvs) != 0 {
		t.Error("刷新后旧服务应该被删除")
	}
}

// TestEtcdRegistry_IsHealthy 测试健康检查
func TestEtcdRegistry_IsHealthy(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 正常情况下应该健康
	if !registry.IsHealthy() {
		t.Error("健康检查应该返回 true")
	}

	// 关闭客户端后应该不健康
	registry.cli.Close()
	if registry.IsHealthy() {
		t.Error("关闭客户端后健康检查应该返回 false")
	}

	// 客户端为 nil 时应该不健康
	registry.cli = nil
	if registry.IsHealthy() {
		t.Error("客户端为 nil 时健康检查应该返回 false")
	}
}

// TestEtcdRegistry_GetLeaseID 测试获取租约ID
func TestEtcdRegistry_GetLeaseID(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 未注册时租约ID应该为0
	initialLeaseID := registry.GetLeaseID()
	if initialLeaseID != 0 {
		t.Error("未注册时租约ID应该为0")
	}

	// 注册服务
	registry.Publisher("test-service")
	time.Sleep(150 * time.Millisecond)

	leaseID := registry.GetLeaseID()
	if leaseID == 0 {
		t.Error("注册后租约ID不应该为0")
	}

	// 验证返回的租约ID与内部租约ID一致
	if leaseID != uint64(registry.leaseID) {
		t.Errorf("返回的租约ID不一致，期望: %d, 实际: %d", registry.leaseID, leaseID)
	}
}

// TestEtcdRegistry_ConcurrentAccess 测试并发访问
func TestEtcdRegistry_ConcurrentAccess(t *testing.T) {
	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/test/concurrent/")

	ctx := context.Background()
	prefix := "/test/concurrent/"
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
				_ = registry.GetValue(key)
			}
		}(i)
	}

	wg.Wait()

	// 验证数据
	time.Sleep(200 * time.Millisecond)
	kvs := registry.GetValuesTyped(prefix, clientv3.WithPrefix())
	expectedCount := concurrency * operationsPerGoroutine

	if len(kvs) != expectedCount {
		t.Errorf("并发写入后键值对数量不正确，期望: %d, 实际: %d", expectedCount, len(kvs))
	}
}

// TestEtcdRegistry_ConcurrentPublisher 测试并发发布服务
func TestEtcdRegistry_ConcurrentPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发发布测试")
	}

	registry := newTestEtcdRegistry(t)
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

	// 最后一次发布的值应该被保留
	if registry.val == "" {
		t.Error("并发发布后服务值不应该为空")
	}

	if registry.leaseID == 0 {
		t.Error("并发发布后租约ID不应该为0")
	}
}

// TestEtcdRegistry_Performance_LargeData 性能测试：大数据量操作
func TestEtcdRegistry_Performance_LargeData(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/perf/large/")

	ctx := context.Background()
	prefix := "/perf/large/"
	dataSize := 500 // 减少数据量以加快测试

	// 测试大量写入
	start := time.Now()
	for i := range dataSize {
		key := fmt.Sprintf("%skey-%d", prefix, i)
		value := fmt.Sprintf("value-%d", i)
		registry.Put(ctx, key, value)
	}
	writeDuration := time.Since(start)

	t.Logf("写入 %d 条数据耗时: %v (平均: %v/条)", dataSize, writeDuration, writeDuration/time.Duration(dataSize))

	// 测试批量读取
	time.Sleep(200 * time.Millisecond)
	start = time.Now()
	kvs := registry.GetValuesTyped(prefix, clientv3.WithPrefix())
	readDuration := time.Since(start)

	t.Logf("读取 %d 条数据耗时: %v", len(kvs), readDuration)

	if len(kvs) != dataSize {
		t.Errorf("数据量不匹配，期望: %d, 实际: %d", dataSize, len(kvs))
	}

	// 测试单个键读取性能
	start = time.Now()
	for i := range 100 {
		key := fmt.Sprintf("%skey-%d", prefix, i)
		_ = registry.GetValue(key)
	}
	singleReadDuration := time.Since(start)
	t.Logf("单个键读取 100 次耗时: %v (平均: %v/次)", singleReadDuration, singleReadDuration/100)
}

// TestEtcdRegistry_Performance_HighConcurrency 性能测试：高并发场景
func TestEtcdRegistry_Performance_HighConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/perf/concurrent/")

	ctx := context.Background()
	prefix := "/perf/concurrent/"
	goroutines := 20
	operationsPerGoroutine := 50

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 记录错误
	var errorCount int32
	var mu sync.Mutex

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range operationsPerGoroutine {
				key := fmt.Sprintf("%skey-%d-%d", prefix, id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				registry.Put(ctx, key, value)
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

	mu.Lock()
	if errorCount > 0 {
		t.Logf("错误次数: %d", errorCount)
	}
	mu.Unlock()
}

// TestEtcdRegistry_Performance_LeaseRenewal 性能测试：租约续约
func TestEtcdRegistry_Performance_LeaseRenewal(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	// 注册服务
	registry.Publisher("test-service")
	time.Sleep(150 * time.Millisecond)

	initialLeaseID := registry.leaseID
	if initialLeaseID == 0 {
		t.Fatal("租约ID不应该为0")
	}

	// 启动租约监听
	go registry.ListenLeaseRespChan()

	// 观察租约续约情况
	testDuration := 8 * time.Second
	t.Logf("观察租约续约 %v，租约TTL: %d秒", testDuration, testEtcdConfig.LeaseTTL)

	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	checkCount := 0
	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(start)
			if elapsed >= testDuration {
				t.Logf("租约续约测试完成，检查次数: %d", checkCount)

				// 验证租约仍然有效
				if registry.leaseID == initialLeaseID {
					t.Log("租约ID保持不变，续约正常")
				} else {
					t.Logf("租约ID已更新: %d -> %d", initialLeaseID, registry.leaseID)
				}
				return
			}
			checkCount++

			// 验证服务仍然存在
			ctx := context.Background()
			key := fmt.Sprintf("%s/%d", registry.key, registry.leaseID)
			resp, err := registry.cli.Get(ctx, key)
			if err != nil {
				t.Errorf("获取服务失败: %v", err)
				return
			}

			if len(resp.Kvs) == 0 {
				t.Error("服务不存在，租约可能已过期")
				return
			}

			t.Logf("租约有效，已运行: %v", elapsed.Round(time.Second))

		case <-registry.ctx.Done():
			t.Error("上下文被取消")
			return
		}
	}
}

// TestEtcdRegistry_Performance_WatchLatency 性能测试：Watch 延迟
func TestEtcdRegistry_Performance_WatchLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestEtcdRegistry(t)
	defer registry.Close()
	defer cleanupTestData(t, registry, "/perf/watch/")

	watchPrefix := "/perf/watch/"
	testCount := 50 // 减少测试次数

	latencies := make([]time.Duration, 0, testCount)
	var mu sync.Mutex
	eventChan := make(chan time.Time, testCount)

	// 启动监听
	registry.WatchWithCallback(watchPrefix, func(event *clientv3.Event) {
		eventChan <- time.Now()
	})

	time.Sleep(150 * time.Millisecond)

	ctx := context.Background()

	// 测试多次 Watch 延迟
	successCount := 0
	for i := range testCount {
		key := fmt.Sprintf("%skey-%d", watchPrefix, i)
		value := fmt.Sprintf("value-%d", i)

		sendTime := time.Now()
		registry.Put(ctx, key, value)

		select {
		case receiveTime := <-eventChan:
			latency := receiveTime.Sub(sendTime)
			mu.Lock()
			latencies = append(latencies, latency)
			mu.Unlock()
			successCount++
		case <-time.After(2 * time.Second):
			t.Logf("第 %d 次测试超时", i)
		}
	}

	if len(latencies) == 0 {
		t.Fatal("没有收到任何 Watch 事件")
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

	t.Logf("Watch 延迟统计 (成功: %d/%d):", successCount, testCount)
	t.Logf("  平均延迟: %v", avgLatency)
	t.Logf("  最小延迟: %v", minLatency)
	t.Logf("  最大延迟: %v", maxLatency)
}

// TestEtcdRegistry_Performance_DistributedLock 性能测试：分布式锁
func TestEtcdRegistry_Performance_DistributedLock(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	registry := newTestEtcdRegistry(t)
	defer registry.Close()

	lockKey := "/perf/lock/test"
	goroutines := 5
	operationsPerGoroutine := 5

	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := time.Now()
	lockTimes := make([]time.Duration, 0, goroutines*operationsPerGoroutine)

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()

			for range operationsPerGoroutine {
				lockStart := time.Now()

				// 创建会话
				session, err := concurrency.NewSession(registry.cli)
				if err != nil {
					t.Errorf("创建会话失败: %v", err)
					return
				}

				// 创建互斥锁
				mutex := concurrency.NewMutex(session, lockKey)

				// 获取锁
				if err := mutex.Lock(context.Background()); err != nil {
					t.Errorf("获取锁失败: %v", err)
					session.Close()
					return
				}

				lockDuration := time.Since(lockStart)
				mu.Lock()
				lockTimes = append(lockTimes, lockDuration)
				mu.Unlock()

				// 临界区操作
				counter++
				time.Sleep(5 * time.Millisecond)

				// 释放锁
				if err := mutex.Unlock(context.Background()); err != nil {
					t.Errorf("释放锁失败: %v", err)
				}

				session.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	expectedCounter := goroutines * operationsPerGoroutine
	if counter != expectedCounter {
		t.Errorf("计数器值不正确，期望: %d, 实际: %d", expectedCounter, counter)
	}

	// 计算锁获取时间统计
	var totalLockTime time.Duration
	for _, lt := range lockTimes {
		totalLockTime += lt
	}
	avgLockTime := totalLockTime / time.Duration(len(lockTimes))

	t.Logf("分布式锁测试: %d 个协程，每个 %d 次操作", goroutines, operationsPerGoroutine)
	t.Logf("总操作数: %d, 耗时: %v", expectedCounter, duration)
	t.Logf("平均每次加锁耗时: %v", avgLockTime)
	t.Logf("平均每次操作耗时: %v", duration/time.Duration(expectedCounter))
}

// BenchmarkEtcdRegistry_Put 基准测试：Put 操作
func BenchmarkEtcdRegistry_Put(b *testing.B) {
	t := &testing.T{}
	registry := newTestEtcdRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：etcd 不可用")
	}
	defer registry.Close()
	defer cleanupTestData(t, registry, "/bench/put/")

	ctx := context.Background()
	key := "/bench/put/key"

	b.ResetTimer()
	i := 0
	for b.Loop() {
		value := fmt.Sprintf("value-%d", i)
		registry.Put(ctx, key, value)
		i++
	}
}

// BenchmarkEtcdRegistry_GetValue 基准测试：GetValue 操作
func BenchmarkEtcdRegistry_GetValue(b *testing.B) {
	t := &testing.T{}
	registry := newTestEtcdRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：etcd 不可用")
	}
	defer registry.Close()
	defer cleanupTestData(t, registry, "/bench/get/")

	ctx := context.Background()
	key := "/bench/get/key"
	value := "bench-value"
	registry.Put(ctx, key, value)
	time.Sleep(50 * time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		_ = registry.GetValue(key)
	}
}

// BenchmarkEtcdRegistry_GetValues 基准测试：GetValues 操作
func BenchmarkEtcdRegistry_GetValues(b *testing.B) {
	t := &testing.T{}
	registry := newTestEtcdRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：etcd 不可用")
	}
	defer registry.Close()
	defer cleanupTestData(t, registry, "/bench/getvalues/")

	ctx := context.Background()
	prefix := "/bench/getvalues/"

	// 准备测试数据
	for i := range 100 {
		key := fmt.Sprintf("%skey-%d", prefix, i)
		value := fmt.Sprintf("value-%d", i)
		registry.Put(ctx, key, value)
	}

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for b.Loop() {
		_ = registry.GetValuesTyped(prefix, clientv3.WithPrefix())
	}
}

// BenchmarkEtcdRegistry_Publisher 基准测试：Publisher 操作
func BenchmarkEtcdRegistry_Publisher(b *testing.B) {
	t := &testing.T{}
	registry := newTestEtcdRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：etcd 不可用")
	}
	defer registry.Close()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		value := fmt.Sprintf("service-%d", i)
		registry.Publisher(value)
		time.Sleep(20 * time.Millisecond) // 避免过快创建租约
		i++
	}
}

// BenchmarkEtcdRegistry_ConcurrentPut 基准测试：并发 Put 操作
func BenchmarkEtcdRegistry_ConcurrentPut(b *testing.B) {
	t := &testing.T{}
	registry := newTestEtcdRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：etcd 不可用")
	}
	defer registry.Close()
	defer cleanupTestData(t, registry, "/bench/concurrent/")

	ctx := context.Background()
	prefix := "/bench/concurrent/"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("%skey-%d", prefix, i)
			value := fmt.Sprintf("value-%d", i)
			registry.Put(ctx, key, value)
			i++
		}
	})
}

// BenchmarkEtcdRegistry_ConcurrentGet 基准测试：并发 Get 操作
func BenchmarkEtcdRegistry_ConcurrentGet(b *testing.B) {
	t := &testing.T{}
	registry := newTestEtcdRegistry(t)
	if t.Skipped() {
		b.Skip("跳过基准测试：etcd 不可用")
	}
	defer registry.Close()
	defer cleanupTestData(t, registry, "/bench/concurrent/")

	ctx := context.Background()
	key := "/bench/concurrent/get"
	registry.Put(ctx, key, "test-value")
	time.Sleep(50 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = registry.GetValue(key)
		}
	})
}
