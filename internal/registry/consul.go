package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/spelens-gud/assert"
	"github.com/spelens-gud/logger"
)

// ConsulRegistry consul注册中心实现
type ConsulRegistry struct {
	client     *api.Client
	cnf        *ConsulConfig
	log        logger.ILogger
	lock       sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	registered bool // 标记服务是否已注册
}

// 确保 ConsulRegistry 实现了 Registry 接口
var _ Registry = (*ConsulRegistry)(nil)

// New 初始化consul客户端
func (c *ConsulRegistry) New() {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	config := api.DefaultConfig()
	config.Address = c.cnf.Address
	config.Scheme = c.cnf.GetScheme()
	config.Datacenter = c.cnf.Datacenter

	assert.MayTrue(c.cnf.HasToken(), func() {
		config.Token = c.cnf.Token
	})

	assert.MayTrue(c.cnf.HasTLS(), func() {
		tlsConfig := &api.TLSConfig{
			CertFile:           c.cnf.TLSConfig.CertFile,
			KeyFile:            c.cnf.TLSConfig.KeyFile,
			CAFile:             c.cnf.TLSConfig.CAFile,
			InsecureSkipVerify: c.cnf.TLSConfig.InsecureSkipVerify,
		}
		config.TLSConfig = *tlsConfig
	})

	c.client = assert.ShouldCall1RE(api.NewClient, config, "创建consul客户端失败")

	c.log.Infof("初始化consul注册中心，地址: %s", c.cnf.Address)
}

// Publisher 注册服务
func (c *ConsulRegistry) Publisher(value string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.log.Infof("注册consul服务: %s", c.cnf.ServiceName)

	registration := &api.AgentServiceRegistration{
		ID:                c.cnf.GetServiceID(),
		Name:              c.cnf.ServiceName,
		Address:           c.cnf.ServiceAddress,
		Port:              c.cnf.ServicePort,
		Tags:              c.cnf.ServiceTags,
		Meta:              c.cnf.ServiceMeta,
		EnableTagOverride: c.cnf.EnableTagOverride,
	}

	// 添加健康检查
	assert.MayTrue(c.cnf.HaHealthCheckPath(), func() {
		check := &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("%s://%s:%d%s", c.cnf.GetScheme(), c.cnf.ServiceAddress, c.cnf.ServicePort, c.cnf.HealthCheckPath),
			Interval:                       c.cnf.GetHealthCheckInterval(),
			Timeout:                        c.cnf.GetHealthCheckTimeout(),
			DeregisterCriticalServiceAfter: c.cnf.GetDeregisterAfter(),
		}
		registration.Check = check
	})

	// 企业版特性
	assert.MayTrue(c.cnf.HasNamespace(), func() {
		registration.Namespace = c.cnf.Namespace
	})
	assert.MayTrue(c.cnf.HasPartition(), func() {
		registration.Partition = c.cnf.Partition
	})

	assert.ShouldCall1E(c.client.Agent().ServiceRegister, registration, "注册consul服务失败")

	c.registered = true
	c.log.Infof("consul服务注册成功")
}

// Deregister 注销服务
func (c *ConsulRegistry) Deregister() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 如果服务未注册，直接返回
	if !c.registered {
		return
	}

	c.log.Infof("注销consul服务: %s", c.cnf.GetServiceID())

	assert.ShouldCall1E(c.client.Agent().ServiceDeregister, c.cnf.GetServiceID(), "注销consul服务失败")

	c.registered = false
	c.log.Infof("consul服务注销成功")
}

// GetValue 获取单个服务实例
func (c *ConsulRegistry) GetValue(key string, opts ...any) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// 默认不要求健康检查通过，因为服务刚注册时健康检查可能还未完成
	passingOnly := false
	if len(opts) > 0 {
		if passing, ok := opts[0].(bool); ok {
			passingOnly = passing
		}
	}

	services, _, err := c.client.Health().Service(key, "", passingOnly, nil)
	if err != nil {
		c.log.Errorf("获取consul服务失败: %v", err)
		return ""
	}
	if len(services) == 0 {
		return ""
	}
	// 返回第一个服务实例
	service := services[0]
	address := service.Service.Address
	if address == "" {
		address = service.Node.Address
	}
	return fmt.Sprintf("%s:%d", address, service.Service.Port)
}

// GetValues 获取所有服务实例
func (c *ConsulRegistry) GetValues(key string, opts ...any) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// 默认不要求健康检查通过，因为服务刚注册时健康检查可能还未完成
	passingOnly := false
	if len(opts) > 0 {
		if passing, ok := opts[0].(bool); ok {
			passingOnly = passing
		}
	}

	services, _, err := c.client.Health().Service(key, "", passingOnly, nil)
	if err != nil {
		c.log.Errorf("获取consul服务列表失败: %v", err)
		return nil
	}
	return services
}

// Put 创建或更新键值
func (c *ConsulRegistry) Put(ctx context.Context, key string, val string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.log.Infof("写入consul KV: %s = %s", key, val)

	kv := &api.KVPair{
		Key:   key,
		Value: []byte(val),
	}
	assert.ShouldCall2RE(c.client.KV().Put, kv, nil, "写入consul KV失败")
}

// Watch 监听服务变化
func (c *ConsulRegistry) Watch(ctx context.Context, prefix string) interface{} {
	c.log.Infof("开始监听consul服务变化: %s", prefix)

	go logger.WithRecover(c.log, func() {
		var lastIndex uint64
		for {
			select {
			case <-ctx.Done():
				c.log.Infof("停止监听consul服务: %s", prefix)
				return
			default:
				queryOpts := &api.QueryOptions{
					WaitIndex: lastIndex,
					WaitTime:  time.Minute,
				}
				services, meta, err := c.client.Health().Service(prefix, "", false, queryOpts)
				if err != nil {
					c.log.Errorf("监听consul服务失败: %v", err)
					time.Sleep(time.Second)
					continue
				}
				lastIndex = meta.LastIndex
				c.log.Infof("consul服务变化: %v", services)
			}
		}
	})

	return nil
}

// Close 关闭consul客户端
func (c *ConsulRegistry) Close() {
	c.log.Infof("关闭consul注册中心")

	assert.MayTrue(c.client != nil, func() {
		c.cancel()
	})

	// 注销服务
	c.Deregister()

	c.log.Infof("consul注册中心关闭成功")
}

// IsHealthy 健康检查
func (c *ConsulRegistry) IsHealthy() bool {
	if c.client == nil {
		return false
	}
	_, err := c.client.Agent().Self()
	return err == nil
}

// Refresh 刷新服务注册
func (c *ConsulRegistry) Refresh() {
	c.log.Infof("刷新consul服务注册")

	// 先注销
	c.Deregister()

	// 重新注册
	c.Publisher("")

	c.log.Infof("consul服务刷新成功")
}

// GetLeaseID 获取租约ID（consul不使用租约）
func (c *ConsulRegistry) GetLeaseID() uint64 {
	return 0
}
