package registry

import (
	"context"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"sync"

	"github.com/spelens-gud/trunk/internal/logger"
)

// NacosRegistry nacos注册中心实现
type NacosRegistry struct {
	namingClient naming_client.INamingClient
	cnf          *NacosConfig
	log          logger.ILogger
	lock         sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// 确保 NacosRegistry 实现了 Registry 接口
var _ Registry = (*NacosRegistry)(nil)

// New 初始化nacos客户端
func (n *NacosRegistry) New() {
	n.ctx, n.cancel = context.WithCancel(context.Background())
	n.log.Infof("初始化nacos注册中心，服务器: %v", n.cnf.Hosts)

	serverConfigs := make([]constant.ServerConfig, 0, len(n.cnf.Hosts))
	for _, host := range n.cnf.Hosts {
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: host,
			Port:   n.cnf.GetPort(),
		})
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         n.cnf.NamespaceId,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              n.cnf.LogDir,
		CacheDir:            n.cnf.CacheDir,
		LogLevel:            n.cnf.LogLevel,
	}

	if n.cnf.HasAuth() {
		clientConfig.Username = n.cnf.Username
		clientConfig.Password = n.cnf.Password
	}

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		n.log.Errorf("创建nacos客户端失败: %v", err)
		return fmt.Errorf("创建nacos客户端失败: %w", err)
	}

	n.namingClient = namingClient

	n.log.Warnf("nacos注册中心实现待完成，请引入 github.com/nacos-group/nacos-sdk-go/v2")
	return nil
}

// Publisher 注册服务
func (n *NacosRegistry) Publisher(value string) {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.log.Infof("注册nacos服务: %s", n.cnf.ServiceName)

	// TODO: 实现服务注册
	success, err := n.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          n.cnf.IP,
		Port:        n.cnf.ServicePort,
		ServiceName: n.cnf.ServiceName,
		GroupName:   n.cnf.GetGroupName(),
		ClusterName: n.cnf.GetClusterName(),
		Weight:      n.cnf.GetWeight(),
		Enable:      n.cnf.IsEnable(),
		Healthy:     n.cnf.IsHealthy(),
		Ephemeral:   n.cnf.IsEphemeral(),
		Metadata:    n.cnf.Metadata,
	})
	if err != nil || !success {
		n.log.Errorf("注册nacos服务失败: %v", err)
		return
	}

	n.log.Infof("nacos服务注册成功")
}

// Deregister 注销服务
func (n *NacosRegistry) Deregister() {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.log.Infof("注销nacos服务: %s", n.cnf.ServiceName)

	success, err := n.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          n.cnf.IP,
		Port:        n.cnf.ServicePort,
		ServiceName: n.cnf.ServiceName,
		GroupName:   n.cnf.GetGroupName(),
		Cluster:     n.cnf.GetClusterName(),
		Ephemeral:   n.cnf.IsEphemeral(),
	})
	if err != nil || !success {
		n.log.Errorf("注销nacos服务失败: %v", err)
		return
	}

	n.log.Infof("nacos服务注销成功")
}

// GetValue 获取单个服务实例
func (n *NacosRegistry) GetValue(key string, opts ...interface{}) string {
	n.lock.RLock()
	defer n.lock.RUnlock()

	// TODO: 实现获取服务实例
	// 示例代码:
	/*
		instance, err := n.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
			ServiceName: key,
			GroupName:   n.cnf.GetGroupName(),
			Clusters:    []string{n.cnf.GetClusterName()},
		})
		if err != nil {
			n.log.Errorf("获取nacos服务实例失败: %v", err)
			return ""
		}
		return fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	*/

	return ""
}

// GetValues 获取所有服务实例
func (n *NacosRegistry) GetValues(key string, opts ...interface{}) interface{} {
	n.lock.RLock()
	defer n.lock.RUnlock()

	// TODO: 实现获取所有服务实例
	// 示例代码:
	/*
		instances, err := n.namingClient.SelectInstances(vo.SelectInstancesParam{
			ServiceName: key,
			GroupName:   n.cnf.GetGroupName(),
			Clusters:    []string{n.cnf.GetClusterName()},
			HealthyOnly: true,
		})
		if err != nil {
			n.log.Errorf("获取nacos服务实例列表失败: %v", err)
			return nil
		}
		return instances
	*/

	return nil
}

// Put 创建或更新服务
func (n *NacosRegistry) Put(ctx context.Context, key string, val string) {
	// Nacos 使用服务注册而不是键值存储
	return fmt.Errorf("nacos不支持Put操作，请使用Publisher注册服务")
}

// Watch 监听服务变化
func (n *NacosRegistry) Watch(ctx context.Context, prefix string) interface{} {
	n.log.Infof("开始监听nacos服务变化: %s", prefix)

	// TODO: 实现服务监听
	// 示例代码:
	/*
		err := n.namingClient.Subscribe(&vo.SubscribeParam{
			ServiceName: prefix,
			GroupName:   n.cnf.GetGroupName(),
			Clusters:    []string{n.cnf.GetClusterName()},
			SubscribeCallback: func(services []model.Instance, err error) {
				if err != nil {
					n.log.Errorf("nacos服务变化回调错误: %v", err)
					return
				}
				n.log.Infof("nacos服务变化: %v", services)
			},
		})
		if err != nil {
			n.log.Errorf("订阅nacos服务失败: %v", err)
		}
	*/

	return nil
}

// Close 关闭nacos客户端
func (n *NacosRegistry) Close() {
	n.log.Infof("关闭nacos注册中心")

	if n.cancel != nil {
		n.cancel()
	}

	// 注销服务
	if err := n.Deregister(); err != nil {
		n.log.Errorf("注销nacos服务失败: %v", err)
	}

	n.log.Infof("nacos注册中心关闭成功")
	return nil
}

// IsHealthy 健康检查
func (n *NacosRegistry) IsHealthy() bool {
	// TODO: 实现健康检查
	// 可以尝试获取服务列表来验证连接
	return true
}

// Refresh 刷新服务注册
func (n *NacosRegistry) Refresh() {
	n.log.Infof("刷新nacos服务注册")

	// 先注销
	n.Deregister()

	// 重新注册
	if err := n.Publisher(""); err != nil {
		n.log.Errorf("重新注册服务失败: %v", err)
		return err
	}

	n.log.Infof("nacos服务刷新成功")
}

// GetLeaseID 获取租约ID(nacos不使用租约)
func (n *NacosRegistry) GetLeaseID() uint64 {
	return 0
}
