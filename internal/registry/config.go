package registry

import (
	"errors"
)

var (
	// errEmptyEtcdHosts etcd主机列表为空
	errEmptyEtcdHosts = errors.New("empty registry hosts")
	// errEmptyEtcdKey etcd键值为空
	errEmptyEtcdKey = errors.New("empty registry key")
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
	}
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
		return errEmptyEtcdHosts
	} else if len(c.Key) == 0 {
		return errEmptyEtcdKey
	} else {
		return nil
	}
}
