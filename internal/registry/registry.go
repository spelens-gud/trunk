package registry

import (
	"context"

	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
)

type IRegistry interface {
	// CreateEtcdRegistry 创建Etcd注册中心实例
	CreateEtcdRegistry(config *EtcdConfig) (Registry, error)
	// CreateNacosRegistry 创建Nacos注册中心实例
	CreateNacosRegistry(config *NacosConfig) (Registry, error)
	// CreateConsulRegistry 创建Consul注册中心实例
	CreateConsulRegistry(config *ConsulConfig) (Registry, error)
}

// Registry 注册中心接口
type Registry interface {
	// New 初始化注册中心客户端
	New()
	// Publisher 发布/注册服务
	Publisher(value string)
	// Deregister 注销服务
	Deregister()
	// GetValue 获取单个值
	GetValue(key string, opts ...any) string
	// GetValues 获取多个值
	GetValues(key string, opts ...any) any
	// Put 创建或更新键值
	Put(ctx context.Context, key string, val string)
	// Watch 监听键变化
	Watch(ctx context.Context, prefix string) any
	// Close 关闭注册中心连接
	Close()
	// IsHealthy 健康检查
	IsHealthy() bool
	// Refresh 刷新服务注册
	Refresh()
	// GetLeaseID 获取租约ID（仅etcd使用）
	GetLeaseID() uint64
}

// GRegistryFactory 注册中心工厂
type GRegistryFactory struct {
	log logger.ILogger
}

// NewRegistryFactory 创建注册中心工厂
func NewRegistryFactory(log logger.ILogger) *GRegistryFactory {
	return &GRegistryFactory{
		log: log,
	}
}

// CreateEtcdRegistry 创建etcd注册中心
func (f *GRegistryFactory) CreateEtcdRegistry(config *EtcdConfig) (Registry, error) {
	// 必须验证配置,否则阻断程序
	assert.MustCall0E(config.Validate)
	registry := &EtcdRegistry{
		log: f.log,
		cnf: config,
	}

	registry.New()

	return registry, nil
}

// CreateNacosRegistry 创建nacos注册中心
func (f *GRegistryFactory) CreateNacosRegistry(config *NacosConfig) (Registry, error) {
	// 必须验证配置,否则阻断程序
	assert.MustCall0E(config.Validate)
	registry := &NacosRegistry{
		log: f.log,
		cnf: config,
	}

	registry.New()

	return registry, nil
}

// CreateConsulRegistry 创建consul注册中心
func (f *GRegistryFactory) CreateConsulRegistry(config *ConsulConfig) (Registry, error) {
	// 必须验证配置,否则阻断程序
	assert.MustCall0E(config.Validate)
	registry := &ConsulRegistry{
		log: f.log,
		cnf: config,
	}

	registry.New()

	return registry, nil
}
