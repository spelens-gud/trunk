package registry

import (
	"errors"
)

var (
	// errEmptyNacosHosts nacos主机列表为空
	errEmptyNacosHosts = errors.New("empty nacos hosts")
	// errEmptyNacosNamespace nacos命名空间为空
	errEmptyNacosNamespace = errors.New("empty nacos namespace")
)

// NacosConfig nacos配置
type NacosConfig struct {
	Hosts       []string          `yaml:"hosts"`       // nacos服务器地址列表
	Port        uint64            `yaml:"port"`        // nacos端口，默认8848
	NamespaceId string            `yaml:"namespaceId"` // 命名空间ID
	GroupName   string            `yaml:"groupName"`   // 分组名称，默认DEFAULT_GROUP
	ClusterName string            `yaml:"clusterName"` // 集群名称，默认DEFAULT
	ServiceName string            `yaml:"serviceName"` // 服务名称
	IP          string            `yaml:"ip"`          // 服务IP
	ServicePort uint64            `yaml:"servicePort"` // 服务端口
	Weight      float64           `yaml:"weight"`      // 权重，默认1.0
	Enable      bool              `yaml:"enable"`      // 是否启用，默认true
	Healthy     bool              `yaml:"healthy"`     // 是否健康，默认true
	Ephemeral   bool              `yaml:"ephemeral"`   // 是否临时实例，默认true
	Metadata    map[string]string `yaml:"metadata"`    // 元数据
	Username    string            `yaml:"username"`    // 用户名
	Password    string            `yaml:"password"`    // 密码
	LogLevel    string            `yaml:"logLevel"`    // 日志级别
	CacheDir    string            `yaml:"cacheDir"`    // 缓存目录
	LogDir      string            `yaml:"logDir"`      // 日志目录
}

// Validate 验证配置
func (c *NacosConfig) Validate() error {
	if len(c.Hosts) == 0 {
		return errEmptyNacosHosts
	}
	if len(c.NamespaceId) == 0 {
		return errEmptyNacosNamespace
	}
	return nil
}

// GetType 获取注册中心类型
func (c *NacosConfig) GetType() RegistryType {
	return RegistryTypeNacos
}

// GetPort 获取端口，默认8848
func (c *NacosConfig) GetPort() uint64 {
	if c.Port == 0 {
		return 8848
	}
	return c.Port
}

// GetGroupName 获取分组名称，默认DEFAULT_GROUP
func (c *NacosConfig) GetGroupName() string {
	if c.GroupName == "" {
		return "DEFAULT_GROUP"
	}
	return c.GroupName
}

// GetClusterName 获取集群名称，默认DEFAULT
func (c *NacosConfig) GetClusterName() string {
	if c.ClusterName == "" {
		return "DEFAULT"
	}
	return c.ClusterName
}

// GetWeight 获取权重，默认1.0
func (c *NacosConfig) GetWeight() float64 {
	if c.Weight <= 0 {
		return 1.0
	}
	return c.Weight
}

// IsEnable 是否启用
func (c *NacosConfig) IsEnable() bool {
	return c.Enable
}

// IsHealthy 是否健康
func (c *NacosConfig) IsHealthy() bool {
	return c.Healthy
}

// IsEphemeral 是否临时实例
func (c *NacosConfig) IsEphemeral() bool {
	return c.Ephemeral
}

// HasAuth 是否有认证信息
func (c *NacosConfig) HasAuth() bool {
	return c.Username != "" && c.Password != ""
}
