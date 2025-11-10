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

func (s *ServiceCache) Publisher(value string) error {

	s.val = value

	if err := s.putKeyWithLease(6); err != nil {
		return fmt.Errorf(" server put key err:%s", err)
	}

	return nil
}

func (s *ServiceCache) GetCacheClient() *clientv3.Client {
	return s.cli
}

func (s *ServiceCache) Put(ctx context.Context, key string, val string) error {
	s.log.Infof("put key:%s val:%s", key, val)
	_, err := s.cli.Put(ctx, key, val)

	return err
}

// GetValue 获取值
func (s *ServiceCache) GetValue(key string, opts ...clientv3.OpOption) (string, error) {
	resp, err := s.cli.Get(context.Background(), key, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

// GetValues 获取值
func (s *ServiceCache) GetValues(key string, opts ...clientv3.OpOption) ([]*mvccpb.KeyValue, error) {
	resp, err := s.cli.Get(context.Background(), key, opts...)
	if err != nil {
		return nil, err
	}

	return resp.Kvs, nil
}

// 设置租约
func (s *ServiceCache) putKeyWithLease(lease int64) error {
	// 创建一个新的租约，并设置ttl时间
	resp, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}

	// 注册服务并绑定租约
	_, err = s.cli.Put(context.Background(), fmt.Sprintf("%s/%d", s.key, resp.ID), s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// 设置续租 定期发送需求请求
	// KeepAlive使给定的租约永远有效。 如果发布到通道的keepalive响应没有立即被使用，
	// 则租约客户端将至少每秒钟继续向etcd服务器发送保持活动请求，直到获取最新的响应为止。
	// registry client会自动发送ttl到etcd server，从而保证该租约一直有效
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	s.log.Infof("Put key:%s  val:%s  success!", s.key, s.val)

	return nil
}

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
func (s *ServiceCache) Close() error {
	s.log.Infof("close service cache...")
	if s.cli == nil {
		return nil
	}
	// 撤销租约
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	s.log.Infof("撤销租约")
	if err := s.cli.Close(); err != nil {
		return err
	}
	s.log.Infof("ServiceCache[%d] close success!", s.leaseID)

	return nil
}
