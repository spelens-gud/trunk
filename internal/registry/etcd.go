package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	// 默认上下文超时时间
	defaultContextTimeout = 5 * time.Second
)

// EtcdRegistry etcd注册中心实现
type EtcdRegistry struct {
	cli           *clientv3.Client                        // etcd v3客户端
	leaseID       clientv3.LeaseID                        // 租约ID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 租约keepalive响应chan
	key           string                                  // key
	val           string                                  // value
	lock          sync.RWMutex                            // 读写锁
	log           logger.ILogger                          // 日志句柄
	cnf           *EtcdConfig                             // registry 配置
	ctx           context.Context                         // 上下文
	cancel        context.CancelFunc                      // 取消函数
}

// 确保 EtcdRegistry 实现了 Registry 接口
var _ Registry = (*EtcdRegistry)(nil)

// New 创建服务缓存
func (s *EtcdRegistry) New() {
	// 新建etc客户端连接配置
	config := clientv3.Config{
		Endpoints:   s.cnf.Hosts,
		DialTimeout: s.cnf.GetDialTimeout(),
	}

	// 添加认证支持
	assert.Then(s.cnf.HasAccount()).Do(func() {
		config.Username = s.cnf.User
		config.Password = s.cnf.Pass
		s.log.Infof("启用etcd认证，用户: %s", s.cnf.User)
	})

	// 添加TLS支持
	assert.Then(s.cnf.HasTLS()).Do(func() {
		tlsConfig, err := s.cnf.LoadTLSConfig()
		if err != nil {
			s.log.Errorf("TLS加载配置文件错误")
			return
		}
		config.TLS = tlsConfig
		s.log.Infof("启用etcd TLS连接")
	})

	s.key = s.cnf.Key

	s.cli = assert.ShouldCall1RE(clientv3.New, config, "新建etcd客户端失败")
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.log.Infof("etcd客户端创建成功，连接: %v", s.cnf.Hosts)
}

// Publisher 发布服务
func (s *EtcdRegistry) Publisher(value string) {
	s.lock.Lock()
	s.val = value
	s.lock.Unlock()

	// 使用配置的租约TTL
	s.putKeyWithLease(s.cnf.GetLeaseTTL())
}

// GetCacheClient 获取缓存客户端
func (s *EtcdRegistry) GetCacheClient() *clientv3.Client {
	return s.cli
}

// Put 添加服务(KV分布式缓存)
func (s *EtcdRegistry) Put(ctx context.Context, key string, val string) {
	s.log.Infof("put key:%s val:%s", key, val)
	_, err := s.cli.Put(ctx, key, val)
	if err != nil {
		s.log.Errorf(err.Error())
		return
	}
}

// GetValue 获取值,KV分布式缓存
func (s *EtcdRegistry) GetValue(key string, opts ...any) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	// 转换 opts 为 clientv3.OpOption
	etcdOpts := make([]clientv3.OpOption, 0, len(opts))
	for _, opt := range opts {
		if etcdOpt, ok := opt.(clientv3.OpOption); ok {
			etcdOpts = append(etcdOpts, etcdOpt)
		}
	}

	resp, err := s.cli.Get(ctx, key, etcdOpts...)
	if err != nil {
		s.log.Errorf("获取值失败: %v", err)
		return ""
	}

	if len(resp.Kvs) == 0 {
		return ""
	}

	return string(resp.Kvs[0].Value)
}

// GetValues 获取值
func (s *EtcdRegistry) GetValues(key string, opts ...any) any {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	// 转换 opts 为 clientv3.OpOption
	etcdOpts := make([]clientv3.OpOption, 0, len(opts))
	for _, opt := range opts {
		if etcdOpt, ok := opt.(clientv3.OpOption); ok {
			etcdOpts = append(etcdOpts, etcdOpt)
		}
	}

	resp, err := s.cli.Get(ctx, key, etcdOpts...)
	if err != nil {
		s.log.Errorf("获取值失败: %v", err)
		return nil
	}

	return resp.Kvs
}

// GetValuesTyped 获取值(类型安全版本)
func (s *EtcdRegistry) GetValuesTyped(key string, opts ...clientv3.OpOption) []*mvccpb.KeyValue {
	result := s.GetValues(key, convertToInterface(opts)...)
	if kvs, ok := result.([]*mvccpb.KeyValue); ok {
		return kvs
	}
	return nil
}

// convertToInterface 转换为 any 切片
func convertToInterface(opts []clientv3.OpOption) []any {
	result := make([]any, len(opts))
	for i, opt := range opts {
		result[i] = opt
	}
	return result
}

// GetLeaseID 获取租约ID
func (s *EtcdRegistry) GetLeaseID() uint64 {
	return uint64(s.leaseID)
}

// ListenLeaseRespChan 监听续租情况
func (s *EtcdRegistry) ListenLeaseRespChan() {
	s.log.Infof("开始监听租约续约，租约ID: %d", s.leaseID)

	for {
		select {
		case resp, ok := <-s.keepAliveChan:
			if !ok {
				s.log.Errorf("租约续约通道已关闭，租约ID: %d", s.leaseID)
				return
			}
			if resp == nil {
				s.log.Errorf("租约续约失败，可能需要重新注册，租约ID: %d", s.leaseID)

				s.Refresh()
				return
			}
			s.log.Debugf("租约续约成功，租约ID: %d, TTL: %d", resp.ID, resp.TTL)

		case <-s.ctx.Done():
			s.log.Infof("停止监听租约续约，租约ID: %d", s.leaseID)
			return
		}
	}
}

// Close 注销服务
func (s *EtcdRegistry) Close() {
	// 取消上下文，停止所有监听
	assert.MayTrue(s.cancel != nil, func() {
		s.cancel()
	})

	if s.cli == nil {
		s.log.Errorf("etcd v3客户端为空")
		return
	}

	// 确保租约ID不为0, 调用KV功能,不调用租赁
	assert.MayTrue(s.leaseID != 0, func() {
		s.log.Infof("开始关闭服务缓存，租约ID: %d", s.leaseID)
		// 注销服务
		s.Deregister()

		// 撤销租约(只有当租约ID不为0时才撤销)
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()
		assert.ShouldCall2RE(s.cli.Revoke, ctx, s.leaseID, "撤销租约失败")
	})

	// 关闭客户端
	assert.ShouldCall0E(s.cli.Close, "关闭etcd客户端失败")

	assert.MayTrue(s.leaseID != 0, func() {
		s.log.Infof("服务缓存关闭成功，租约ID: %d", s.leaseID)
	})
}

// putKeyWithLease 设置租约
func (s *EtcdRegistry) putKeyWithLease(lease int64) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	// 创建一个新的租约，并设置ttl时间
	resp := assert.ShouldCall2RE(s.cli.Grant, ctx, lease, "创建租约失败")

	s.lock.RLock()
	val := s.val
	s.lock.RUnlock()

	// 注册服务并绑定租约
	serviceKey := fmt.Sprintf("%s/%d", s.key, resp.ID)
	if _, err := s.cli.Put(ctx, serviceKey, val, clientv3.WithLease(resp.ID)); err != nil {
		s.log.Errorf("注册服务失败: %v", err)
		return
	}

	// 设置续租 定期发送续约请求
	// KeepAlive使给定的租约永远有效。如果发布到通道的keepalive响应没有立即被使用，
	// 则租约客户端将至少每秒钟继续向etcd服务器发送保持活动请求，直到获取最新的响应为止。
	// registry client会自动发送ttl到etcd server，从而保证该租约一直有效
	leaseRespChan := assert.ShouldCall2RE(s.cli.KeepAlive, s.ctx, resp.ID, "启动租约续约失败")

	s.lock.Lock()
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	s.lock.Unlock()

	s.log.Infof("服务注册成功 - Key: %s, Value: %s, 租约ID: %d, TTL: %d秒", serviceKey, val, resp.ID, lease)
}

// Deregister 注销服务
func (s *EtcdRegistry) Deregister() {
	s.lock.RLock()
	leaseID := s.leaseID
	s.lock.RUnlock()

	if leaseID == 0 {
		s.log.Warnf("租约ID为0，无需注销")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	key := fmt.Sprintf("%s/%d", s.key, leaseID)
	if _, err := s.cli.Delete(ctx, key); err != nil {
		s.log.Errorf("注销服务失败，Key: %s, 错误: %v", key, err)
		return
	}

	s.log.Infof("服务注销成功，Key: %s", key)
}

// Watch 监听指定前缀的键变化
func (s *EtcdRegistry) Watch(ctx context.Context, prefix string) any {
	s.log.Infof("开始监听键变化，前缀: %s", prefix)
	return s.cli.Watch(ctx, prefix, clientv3.WithPrefix())
}

// WatchTyped 监听指定前缀的键变化(类型安全版本)
func (s *EtcdRegistry) WatchTyped(ctx context.Context, prefix string) clientv3.WatchChan {
	result := s.Watch(ctx, prefix)
	if watchChan, ok := result.(clientv3.WatchChan); ok {
		return watchChan
	}
	return nil
}

// WatchWithCallback 监听指定前缀的键变化并执行回调
func (s *EtcdRegistry) WatchWithCallback(prefix string, callback func(event *clientv3.Event)) {
	watchChan := s.WatchTyped(s.ctx, prefix)

	go logger.WithRecover(s.log, func() {
		s.log.Infof("启动Watch回调监听，前缀: %s", prefix)
		for {
			select {
			case watchResp, ok := <-watchChan:
				if !ok {
					s.log.Errorf("Watch通道已关闭，前缀: %s", prefix)
					return
				}
				if watchResp.Err() != nil {
					s.log.Errorf("Watch错误，前缀: %s, 错误: %v", prefix, watchResp.Err())
					continue
				}

				for _, event := range watchResp.Events {
					s.log.Debugf("收到事件 - 类型: %s, Key: %s, Value: %s",
						event.Type, string(event.Kv.Key), string(event.Kv.Value))
					assert.MayTrue(callback != nil, func() {
						callback(event)
					})
				}

			case <-s.ctx.Done():
				s.log.Infof("停止Watch监听，前缀: %s", prefix)
				return
			}
		}
	})
}

// Refresh 刷新服务注册(重新注册)
func (s *EtcdRegistry) Refresh() {
	s.log.Infof("开始刷新服务注册")

	// 先注销旧的服务
	assert.MayTrue(s.leaseID != 0, func() {
		s.log.Infof("开始注销服务，租约ID: %d", s.leaseID)
		s.Deregister()
	})

	// 重新注册
	s.putKeyWithLease(s.cnf.GetLeaseTTL())
	s.log.Infof("刷新服务注册成功")
}

// IsHealthy 检查服务是否健康
func (s *EtcdRegistry) IsHealthy() bool {
	if s.cli == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 尝试获取一个键来测试连接
	_, err := s.cli.Get(ctx, s.key, clientv3.WithLimit(1))
	return err == nil
}
