package registry

import (
	"errors"
	"time"
)

var (
	// errEmptyConsulAddress consul地址为空
	errEmptyConsulAddress = errors.New("empty consul address")
	// errEmptyConsulServiceName consul服务名为空
	errEmptyConsulServiceName = errors.New("empty consul service name")
)

// ConsulConfig consul配置
type ConsulConfig struct {
	Address             string            `yaml:"address"`             // consul地址，如 "127.0.0.1:8500"
	Scheme              string            `yaml:"scheme"`              // 协议，http或https
	Datacenter          string            `yaml:"datacenter"`          // 数据中心
	Token               string            `yaml:"token"`               // ACL Token
	ServiceName         string            `yaml:"serviceName"`         // 服务名称
	ServiceID           string            `yaml:"serviceId"`           // 服务ID，默认为 ServiceName-IP-Port
	ServiceAddress      string            `yaml:"serviceAddress"`      // 服务地址
	ServicePort         int               `yaml:"servicePort"`         // 服务端口
	ServiceTags         []string          `yaml:"serviceTags"`         // 服务标签
	ServiceMeta         map[string]string `yaml:"serviceMeta"`         // 服务元数据
	HealthCheckPath     string            `yaml:"healthCheckPath"`     // 健康检查路径
	HealthCheckInterval string            `yaml:"healthCheckInterval"` // 健康检查间隔，如 "10s"
	HealthCheckTimeout  string            `yaml:"healthCheckTimeout"`  // 健康检查超时，如 "5s"
	DeregisterAfter     string            `yaml:"deregisterAfter"`     // 注销时间，如 "30s"
	EnableTagOverride   bool              `yaml:"enableTagOverride"`   // 是否允许标签覆盖
	Namespace           string            `yaml:"namespace"`           // 命名空间（企业版）
	Partition           string            `yaml:"partition"`           // 分区（企业版）
	TLSConfig           *ConsulTLSConfig  `yaml:"tlsConfig"`           // TLS配置
}

// ConsulTLSConfig consul TLS配置
type ConsulTLSConfig struct {
	CertFile           string `yaml:"certFile"`           // 证书文件
	KeyFile            string `yaml:"keyFile"`            // 密钥文件
	CAFile             string `yaml:"caFile"`             // CA证书文件
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"` // 是否跳过证书验证
}

// Validate 验证配置
func (c *ConsulConfig) Validate() error {
	if c.Address == "" {
		return errEmptyConsulAddress
	}
	if c.ServiceName == "" {
		return errEmptyConsulServiceName
	}
	return nil
}

// GetType 获取注册中心类型
func (c *ConsulConfig) GetType() RegistryType {
	return RegistryTypeConsul
}

// GetScheme 获取协议，默认http
func (c *ConsulConfig) GetScheme() string {
	if c.Scheme == "" {
		return "http"
	}
	return c.Scheme
}

// GetServiceID 获取服务ID
func (c *ConsulConfig) GetServiceID() string {
	if c.ServiceID == "" {
		return c.ServiceName
	}
	return c.ServiceID
}

// GetHealthCheckInterval 获取健康检查间隔，默认10秒
func (c *ConsulConfig) GetHealthCheckInterval() string {
	if c.HealthCheckInterval == "" {
		return "10s"
	}
	return c.HealthCheckInterval
}

// GetHealthCheckTimeout 获取健康检查超时，默认5秒
func (c *ConsulConfig) GetHealthCheckTimeout() string {
	if c.HealthCheckTimeout == "" {
		return "5s"
	}
	return c.HealthCheckTimeout
}

// GetDeregisterAfter 获取注销时间，默认30秒
func (c *ConsulConfig) GetDeregisterAfter() string {
	if c.DeregisterAfter == "" {
		return "30s"
	}
	return c.DeregisterAfter
}

// HasTLS 是否配置了TLS
func (c *ConsulConfig) HasTLS() bool {
	return c.TLSConfig != nil &&
		c.TLSConfig.CertFile != "" &&
		c.TLSConfig.KeyFile != "" &&
		c.TLSConfig.CAFile != ""
}

// HasToken 是否有Token
func (c *ConsulConfig) HasToken() bool {
	return c.Token != ""
}

// GetHealthCheckIntervalDuration 获取健康检查间隔时间
func (c *ConsulConfig) GetHealthCheckIntervalDuration() time.Duration {
	duration, err := time.ParseDuration(c.GetHealthCheckInterval())
	if err != nil {
		return 10 * time.Second
	}
	return duration
}

// GetHealthCheckTimeoutDuration 获取健康检查超时时间
func (c *ConsulConfig) GetHealthCheckTimeoutDuration() time.Duration {
	duration, err := time.ParseDuration(c.GetHealthCheckTimeout())
	if err != nil {
		return 5 * time.Second
	}
	return duration
}
