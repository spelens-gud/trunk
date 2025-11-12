package registry

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/spelens-gud/trunk/internal/assert"
)

// EtcdConfig etcd配置
type EtcdConfig struct {
	Hosts              []string // etcd主机列表
	Key                string   // etcd键值
	ID                 int64    `yaml:"id"`                 // etcd服务ID
	User               string   `yaml:"user"`               // etcd用户名
	Pass               string   `yaml:"pass"`               // etcd密码
	CertFile           string   `yaml:"certFile"`           // etcd证书文件
	CertKeyFile        string   `yaml:"certKeyFile"`        // etcd密钥文件
	CACertFile         string   `yaml:"caCertFile"`         // etcdCA证书文件
	InsecureSkipVerify bool     `yaml:"insecureSkipVerify"` // 是否跳过证书验证
	LeaseTTL           int64    `yaml:"leaseTTL"`           // 租约TTL（秒），默认6秒
	DialTimeout        int      `yaml:"dialTimeout"`        // 连接超时时间（秒），默认5秒
}

// Copy 复制
func (c *EtcdConfig) Copy() *EtcdConfig {
	hs := make([]string, len(c.Hosts))
	copy(hs, c.Hosts)

	return &EtcdConfig{
		Hosts:              hs,
		Key:                c.Key,
		ID:                 c.ID,
		User:               c.User,
		Pass:               c.Pass,
		CertFile:           c.CertFile,
		CertKeyFile:        c.CertKeyFile,
		CACertFile:         c.CACertFile,
		InsecureSkipVerify: c.InsecureSkipVerify,
		LeaseTTL:           c.LeaseTTL,
		DialTimeout:        c.DialTimeout,
	}
}

// GetLeaseTTL 获取租约TTL，如果未设置则返回默认值6秒
func (c *EtcdConfig) GetLeaseTTL() int64 {
	if c.LeaseTTL <= 0 {
		return 6
	}

	return c.LeaseTTL
}

// GetDialTimeout 获取连接超时时间，如果未设置则返回默认值5秒
func (c *EtcdConfig) GetDialTimeout() time.Duration {
	if c.DialTimeout <= 0 {
		return 5 * time.Second
	}

	return time.Duration(c.DialTimeout) * time.Second
}

// HasAccount 是否有账号
func (c *EtcdConfig) HasAccount() bool {
	return len(c.User) > 0 && len(c.Pass) > 0
}

// HasID 是否有ID
func (c *EtcdConfig) HasID() bool {
	return c.ID > 0
}

// HasTLS 配置了TLS
func (c *EtcdConfig) HasTLS() bool {
	return len(c.CertFile) > 0 && len(c.CertKeyFile) > 0 && len(c.CACertFile) > 0
}

// Validate 验证
func (c *EtcdConfig) Validate() error {
	if len(c.Hosts) == 0 {
		return errors.New("etcd的host为空")
	}
	if len(c.Key) > 0 {
		return errors.New("注册etcd键值不能为空")
	}
	return nil
}

// GetType 获取注册中心类型
func (c *EtcdConfig) GetType() RegistryType {
	return RegistryTypeEtcd
}

// LoadTLSConfig 加载TLS配置
func (c *EtcdConfig) LoadTLSConfig() (config *tls.Config, err error) {
	if !c.HasTLS() {
		return nil, nil
	}

	// 加载客户端证书
	cert := assert.ShouldCall2RE(tls.LoadX509KeyPair, c.CertFile, c.CertKeyFile)

	// 加载CA证书
	caCert := assert.ShouldCall1RE(os.ReadFile, c.CACertFile)

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append CA certificate")
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}, nil
}
