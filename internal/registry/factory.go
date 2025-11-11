package registry

import (
	"fmt"

	"github.com/spelens-gud/trunk/internal/logger"
)

// RegistryFactory 注册中心工厂
type RegistryFactory struct {
	log logger.ILogger
}

// NewRegistryFactory 创建注册中心工厂
func NewRegistryFactory(log logger.ILogger) *RegistryFactory {
	return &RegistryFactory{
		log: log,
	}
}

// CreateRegistry 创建注册中心实例
func (f *RegistryFactory) CreateRegistry(config Config) (Registry, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	registryType := config.GetType()
	f.log.Infof("创建注册中心实例，类型: %s", registryType)

	switch registryType {
	case RegistryTypeEtcd:
		etcdConfig, ok := config.(*EtcdConfig)
		if !ok {
			return nil, fmt.Errorf("无效的etcd配置类型")
		}
		return f.createEtcdRegistry(etcdConfig)

	case RegistryTypeNacos:
		nacosConfig, ok := config.(*NacosConfig)
		if !ok {
			return nil, fmt.Errorf("无效的nacos配置类型")
		}
		return f.createNacosRegistry(nacosConfig)

	case RegistryTypeConsul:
		consulConfig, ok := config.(*ConsulConfig)
		if !ok {
			return nil, fmt.Errorf("无效的consul配置类型")
		}
		return f.createConsulRegistry(consulConfig)

	default:
		return nil, fmt.Errorf("不支持的注册中心类型: %s", registryType)
	}
}

// createEtcdRegistry 创建etcd注册中心
func (f *RegistryFactory) createEtcdRegistry(config *EtcdConfig) (Registry, error) {
	registry := &EtcdRegistry{
		log: f.log,
		cnf: config,
	}

	if err := registry.New(); err != nil {
		return nil, fmt.Errorf("创建etcd注册中心失败: %w", err)
	}

	return registry, nil
}

// createNacosRegistry 创建nacos注册中心
func (f *RegistryFactory) createNacosRegistry(config *NacosConfig) (Registry, error) {
	registry := &NacosRegistry{
		log: f.log,
		cnf: config,
	}

	if err := registry.New(); err != nil {
		return nil, fmt.Errorf("创建nacos注册中心失败: %w", err)
	}

	return registry, nil
}

// createConsulRegistry 创建consul注册中心
func (f *RegistryFactory) createConsulRegistry(config *ConsulConfig) (Registry, error) {
	registry := &ConsulRegistry{
		log: f.log,
		cnf: config,
	}

	if err := registry.New(); err != nil {
		return nil, fmt.Errorf("创建consul注册中心失败: %w", err)
	}

	return registry, nil
}
