package registry

import (
	"context"
)

// RegistryType 注册中心类型
type RegistryType string

const (
	// RegistryTypeEtcd etcd注册中心
	RegistryTypeEtcd RegistryType = "etcd"
	// RegistryTypeNacos nacos注册中心
	RegistryTypeNacos RegistryType = "nacos"
	// RegistryTypeConsul consul注册中心
	RegistryTypeConsul RegistryType = "consul"
)

// Registry 注册中心接口
type Registry interface {
	// New 初始化注册中心客户端
	New() error

	// Publisher 发布/注册服务
	Publisher(value string) error

	// Deregister 注销服务
	Deregister() error

	// GetValue 获取单个值
	GetValue(key string, opts ...any) string

	// GetValues 获取多个值
	GetValues(key string, opts ...any) any

	// Put 创建或更新键值
	Put(ctx context.Context, key string, val string) error

	// Watch 监听键变化
	Watch(ctx context.Context, prefix string) any

	// Close 关闭注册中心连接
	Close() error

	// IsHealthy 健康检查
	IsHealthy() bool

	// Refresh 刷新服务注册
	Refresh() error

	// GetLeaseID 获取租约ID（仅etcd使用）
	GetLeaseID() uint64
}

// Config 注册中心配置接口
type Config interface {
	// Validate 验证配置
	Validate() error

	// GetType 获取注册中心类型
	GetType() RegistryType
}

// Factory 注册中心工厂接口
type Factory interface {
	// CreateRegistry 创建注册中心实例
	CreateRegistry(config Config) (Registry, error)
}
