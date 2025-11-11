package registry_test

import (
	"context"
	"fmt"
	"time"

	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spelens-gud/trunk/internal/registry"
)

// ExampleEtcdRegistry 演示如何使用etcd注册中心
func ExampleEtcdRegistry() {
	// 创建日志实例
	log := logger.NewLogger()

	// 创建etcd配置
	etcdConfig := &registry.EtcdConfig{
		Hosts:    []string{"127.0.0.1:2379"},
		Key:      "/services/my-service",
		LeaseTTL: 10,
	}

	// 创建工厂
	factory := registry.NewRegistryFactory(log)

	// 创建etcd注册中心
	reg, err := factory.CreateRegistry(etcdConfig)
	if err != nil {
		log.Errorf("创建注册中心失败: %v", err)
		return
	}
	defer reg.Close()

	// 注册服务
	if err := reg.Publisher("192.168.1.100:8080"); err != nil {
		log.Errorf("注册服务失败: %v", err)
		return
	}

	// 获取服务信息
	value := reg.GetValue("/services/my-service")
	fmt.Printf("服务地址: %s\n", value)

	// 健康检查
	if reg.IsHealthy() {
		fmt.Println("注册中心连接正常")
	}
}

// ExampleNacosRegistry 演示如何使用nacos注册中心
func ExampleNacosRegistry() {
	// 创建日志实例
	log := logger.NewLogger()

	// 创建nacos配置
	nacosConfig := &registry.NacosConfig{
		Hosts:       []string{"127.0.0.1"},
		Port:        8848,
		NamespaceId: "public",
		GroupName:   "DEFAULT_GROUP",
		ServiceName: "my-service",
		IP:          "192.168.1.100",
		ServicePort: 8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	}

	// 创建工厂
	factory := registry.NewRegistryFactory(log)

	// 创建nacos注册中心
	reg, err := factory.CreateRegistry(nacosConfig)
	if err != nil {
		log.Errorf("创建注册中心失败: %v", err)
		return
	}
	defer reg.Close()

	// 注册服务
	if err := reg.Publisher(""); err != nil {
		log.Errorf("注册服务失败: %v", err)
		return
	}

	// 获取服务实例
	value := reg.GetValue("my-service")
	fmt.Printf("服务地址: %s\n", value)
}

// ExampleConsulRegistry 演示如何使用consul注册中心
func ExampleConsulRegistry() {
	// 创建日志实例
	log := logger.NewLogger()

	// 创建consul配置
	consulConfig := &registry.ConsulConfig{
		Address:             "127.0.0.1:8500",
		Scheme:              "http",
		ServiceName:         "my-service",
		ServiceAddress:      "192.168.1.100",
		ServicePort:         8080,
		ServiceTags:         []string{"v1", "production"},
		HealthCheckPath:     "/health",
		HealthCheckInterval: "10s",
		HealthCheckTimeout:  "5s",
	}

	// 创建工厂
	factory := registry.NewRegistryFactory(log)

	// 创建consul注册中心
	reg, err := factory.CreateRegistry(consulConfig)
	if err != nil {
		log.Errorf("创建注册中心失败: %v", err)
		return
	}
	defer reg.Close()

	// 注册服务
	if err := reg.Publisher(""); err != nil {
		log.Errorf("注册服务失败: %v", err)
		return
	}

	// 获取服务实例
	value := reg.GetValue("my-service")
	fmt.Printf("服务地址: %s\n", value)
}

// ExampleMultiRegistry 演示如何同时使用多个注册中心
func ExampleMultiRegistry() {
	log := logger.NewLogger()
	factory := registry.NewRegistryFactory(log)

	// 配置列表
	configs := []registry.Config{
		&registry.EtcdConfig{
			Hosts: []string{"127.0.0.1:2379"},
			Key:   "/services/my-service",
		},
		&registry.NacosConfig{
			Hosts:       []string{"127.0.0.1"},
			NamespaceId: "public",
			ServiceName: "my-service",
			IP:          "192.168.1.100",
			ServicePort: 8080,
		},
		&registry.ConsulConfig{
			Address:        "127.0.0.1:8500",
			ServiceName:    "my-service",
			ServiceAddress: "192.168.1.100",
			ServicePort:    8080,
		},
	}

	// 创建多个注册中心实例
	registries := make([]registry.Registry, 0, len(configs))
	for _, config := range configs {
		reg, err := factory.CreateRegistry(config)
		if err != nil {
			log.Errorf("创建注册中心失败: %v", err)
			continue
		}
		registries = append(registries, reg)
	}

	// 注册到所有注册中心
	for _, reg := range registries {
		if err := reg.Publisher("192.168.1.100:8080"); err != nil {
			log.Errorf("注册服务失败: %v", err)
		}
	}

	// 保持服务运行
	time.Sleep(time.Minute)

	// 关闭所有注册中心
	for _, reg := range registries {
		reg.Close()
	}
}

// ExampleWatchService 演示如何监听服务变化
func ExampleWatchService() {
	log := logger.NewLogger()
	factory := registry.NewRegistryFactory(log)

	etcdConfig := &registry.EtcdConfig{
		Hosts: []string{"127.0.0.1:2379"},
		Key:   "/services",
	}

	reg, err := factory.CreateRegistry(etcdConfig)
	if err != nil {
		log.Errorf("创建注册中心失败: %v", err)
		return
	}
	defer reg.Close()

	// 监听服务变化
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	watchChan := reg.Watch(ctx, "/services/")
	fmt.Println("开始监听服务变化...")

	// 处理变化事件
	// 注意：实际使用时需要根据不同注册中心的返回类型进行类型断言
}
