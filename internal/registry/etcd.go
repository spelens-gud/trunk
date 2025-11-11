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

// ServiceCache 服务缓存
type ServiceCache struct {
	cli           *clientv3.Client                        // etcd v3客户端
	leaseID       clientv3.LeaseID                        // 租约ID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 租约keepalive相应chan
	key           string                                  // key
	val           string                                  // value
	lock          sync.RWMutex                            // 读写锁
	log           logger.ILogger                          // 日志句柄
	cnf           *EtcdConfig                             // registry 配置
}

// New 创建服务缓存
func (s *ServiceCache) New() {
	cli := assert.MustValue(clientv3.New(clientv3.Config{
		Endpoints:   s.cnf.Hosts,
		DialTimeout: time.Duration(5) * time.Second,
	}))
	s.cli = cli
	s.key = s.cnf.Key
}

// Publisher 发布服务
func (s *ServiceCache) Publisher(value string) {

	s.val = value

	// 错误处理全在调用链最底层, 不讲错误外抛
	s.putKeyWithLease(6)
}

func (s *ServiceCache) GetCacheClient() *clientv3.Client {
	return s.cli
}

// Put 创建服务
func (s *ServiceCache) Put(ctx context.Context, key string, val string) {
	s.log.Infof("put key:%s val:%s", key, val)
	assert.ShouldValue(s.cli.Put(ctx, key, val))
}

// GetValue 获取值
func (s *ServiceCache) GetValue(key string, opts ...clientv3.OpOption) string {
	resp := assert.ShouldValue(s.cli.Get(context.Background(), key, opts...))
	if len(resp.Kvs) == 0 {
		return ""
	}

	return string(resp.Kvs[0].Value)
}

// GetValues 获取值
func (s *ServiceCache) GetValues(key string, opts ...clientv3.OpOption) []*mvccpb.KeyValue {
	resp := assert.ShouldValue(s.cli.Get(context.Background(), key, opts...))

	return resp.Kvs
}

// GetLeaseID 获取租约ID
func (s *ServiceCache) GetLeaseID() uint64 {
	return uint64(s.leaseID)
}

// ListenLeaseRespChan 监听 续租情况
func (s *ServiceCache) ListenLeaseRespChan() {

	for {
		select {
		case _ = <-s.keepAliveChan:

		}
	}
}

// Close 注销服务
func (s *ServiceCache) Close() {
	s.log.Infof("close service cache...")
	if s.cli == nil {
		s.log.Errorf("etcd v3客户端为空")
	}
	// 撤销租约
	assert.ShouldValue(s.cli.Revoke(context.Background(), s.leaseID))
	s.log.Infof("撤销租约")

	assert.ShouldFunc(s.cli.Close)
	s.log.Infof("ServiceCache[%d] close success!", s.leaseID)
}

// putKeyWithLease 设置租约
func (s *ServiceCache) putKeyWithLease(lease int64) {
	// 创建一个新的租约，并设置ttl时间
	resp := assert.ShouldValue(s.cli.Grant(context.Background(), lease))

	// 注册服务并绑定租约
	assert.ShouldValue(s.cli.Put(context.Background(), fmt.Sprintf("%s/%d", s.key, resp.ID), s.val, clientv3.WithLease(resp.ID)))
	// 设置续租 定期发送需求请求
	// KeepAlive使给定的租约永远有效。 如果发布到通道的keepalive响应没有立即被使用，
	// 则租约客户端将至少每秒钟继续向etcd服务器发送保持活动请求，直到获取最新的响应为止。
	// registry client会自动发送ttl到etcd server，从而保证该租约一直有效
	leaseRespChan := assert.ShouldValue(s.cli.KeepAlive(context.Background(), resp.ID))

	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	s.log.Infof("Put key:%s  val:%s  success!", s.key, s.val)
}
