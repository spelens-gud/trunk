# Registry 注册中心组件

基于抽象工厂模式实现的多注册中心支持组件，支持 Etcd、Nacos、Consul 三种主流注册中心。

## 目录

- [特性](#特性)
- [快速开始](#快速开始)
- [架构设计](#架构设计)
- [接口说明](#接口说明)
- [配置说明](#配置说明)
- [高级用法](#高级用法)
- [注册中心对比](#注册中心对比)
- [选型建议](#选型建议)
- [实现状态](#实现状态)
- [最佳实践](#最佳实践)

---

## 特性

- 🏭 **抽象工厂模式**：统一的接口，轻松切换不同的注册中心
- 🔌 **多注册中心支持**：支持 Etcd、Nacos、Consul
- 🔒 **线程安全**：使用读写锁保护并发访问
- 🔐 **安全连接**：支持 TLS/SSL 加密连接和认证
- 💓 **健康检查**：内置健康检查机制
- 🔄 **自动续约**：支持服务自动续约（Etcd）
- 👀 **服务监听**：支持监听服务变化
- 📝 **完善日志**：详细的操作日志记录

---

## 快速开始

### 1. Etcd 注册中心

```go
import (
    "github.com/spelens-gud/trunk/internal/logger"
    "github.com/spelens-gud/trunk/internal/registry"
)

// 创建配置
config := &registry.EtcdConfig{
    Hosts:    []string{"127.0.0.1:2379"},
    Key:      "/services/my-service",
    LeaseTTL: 10, // 租约时间（秒）
    User:     "root",     // 可选：用户名
    Pass:     "password", // 可选：密码
}

// 创建注册中心
log := logger.NewLogger()
factory := registry.NewRegistryFactory(log)
reg, err := factory.CreateRegistry(config)
if err != nil {
    log.Fatalf("创建注册中心失败: %v", err)
}
defer reg.Close()

// 注册服务
err = reg.Publisher("192.168.1.100:8080")
if err != nil {
    log.Errorf("注册服务失败: %v", err)
}

// 获取服务
value := reg.GetValue("/services/my-service")
fmt.Printf("服务地址: %s\n", value)
```

### 2. Nacos 注册中心

```go
config := &registry.NacosConfig{
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
    Username:    "nacos",   // 可选：用户名
    Password:    "nacos",   // 可选：密码
}

factory := registry.NewRegistryFactory(log)
reg, err := factory.CreateRegistry(config)
// ... 使用方式同上
```

### 3. Consul 注册中心

```go
config := &registry.ConsulConfig{
    Address:             "127.0.0.1:8500",
    Scheme:              "http",
    ServiceName:         "my-service",
    ServiceAddress:      "192.168.1.100",
    ServicePort:         8080,
    ServiceTags:         []string{"v1", "production"},
    HealthCheckPath:     "/health",
    HealthCheckInterval: "10s",
    HealthCheckTimeout:  "5s",
    Token:               "your-acl-token", // 可选：ACL Token
}

factory := registry.NewRegistryFactory(log)
reg, err := factory.CreateRegistry(config)
// ... 使用方式同上
```

---

## 架构设计

### 设计模式

本组件采用**抽象工厂模式（Abstract Factory Pattern）**实现多注册中心支持。

#### 为什么选择抽象工厂模式？

1. **统一接口**：不同注册中心提供统一的操作接口，业务代码无需关心底层实现
2. **易于扩展**：新增注册中心只需实现 Registry 接口，无需修改现有代码
3. **灵活切换**：通过配置即可切换不同的注册中心，无需修改业务逻辑
4. **解耦合**：业务代码与具体注册中心实现解耦，降低维护成本

### 架构图

```
┌─────────────────────────────────────────────────────────┐
│                    Business Layer                        │
│                   (业务层使用统一接口)                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  RegistryFactory                         │
│              (工厂负责创建具体实例)                       │
└────────┬──────────────┬──────────────┬──────────────────┘
         │              │              │
         ▼              ▼              ▼
┌────────────┐  ┌────────────┐  ┌────────────┐
│   Etcd     │  │   Nacos    │  │  Consul    │
│  Registry  │  │  Registry  │  │  Registry  │
└────────────┘  └────────────┘  └────────────┘
         │              │              │
         ▼              ▼              ▼
┌────────────┐  ┌────────────┐  ┌────────────┐
│   Etcd     │  │   Nacos    │  │  Consul    │
│   Client   │  │   Client   │  │   Client   │
└────────────┘  └────────────┘  └────────────┘
```

### 组件结构

```
Registry (接口)
    ├── EtcdRegistry (Etcd实现)
    ├── NacosRegistry (Nacos实现)
    └── ConsulRegistry (Consul实现)

Config (接口)
    ├── EtcdConfig
    ├── NacosConfig
    └── ConsulConfig

RegistryFactory (工厂)
    └── CreateRegistry(config Config) Registry
```

### SOLID 设计原则

1. **单一职责原则（SRP）**：每个注册中心实现只负责与对应服务的交互
2. **开闭原则（OCP）**：对扩展开放，对修改关闭
3. **里氏替换原则（LSP）**：所有 Registry 实现可以互相替换
4. **接口隔离原则（ISP）**：Registry 接口定义了必要的方法
5. **依赖倒置原则（DIP）**：业务代码依赖 Registry 接口

---

## 接口说明

### Registry 接口

```go
type Registry interface {
    // 初始化注册中心客户端
    New() error

    // 发布/注册服务
    Publisher(value string) error

    // 注销服务
    Deregister() error

    // 获取单个值
    GetValue(key string, opts ...interface{}) string

    // 获取多个值
    GetValues(key string, opts ...interface{}) interface{}

    // 创建或更新键值
    Put(ctx context.Context, key string, val string) error

    // 监听键变化
    Watch(ctx context.Context, prefix string) interface{}

    // 关闭注册中心连接
    Close() error

    // 健康检查
    IsHealthy() bool

    // 刷新服务注册
    Refresh() error

    // 获取租约ID（仅etcd使用）
    GetLeaseID() uint64
}
```

### Config 接口

```go
type Config interface {
    // 验证配置
    Validate() error

    // 获取注册中心类型
    GetType() RegistryType
}
```

---

## 配置说明

### EtcdConfig

| 字段               | 类型     | 说明                | 默认值 |
| ------------------ | -------- | ------------------- | ------ |
| Hosts              | []string | etcd 服务器地址列表 | 必填   |
| Key                | string   | 服务注册的键前缀    | 必填   |
| LeaseTTL           | int64    | 租约时间（秒）      | 6      |
| DialTimeout        | int      | 连接超时（秒）      | 5      |
| User               | string   | 用户名              | 可选   |
| Pass               | string   | 密码                | 可选   |
| CertFile           | string   | 客户端证书文件      | 可选   |
| CertKeyFile        | string   | 客户端密钥文件      | 可选   |
| CACertFile         | string   | CA 证书文件         | 可选   |
| InsecureSkipVerify | bool     | 是否跳过证书验证    | false  |

### NacosConfig

| 字段        | 类型              | 说明             | 默认值        |
| ----------- | ----------------- | ---------------- | ------------- |
| Hosts       | []string          | Nacos 服务器地址 | 必填          |
| Port        | uint64            | Nacos 端口       | 8848          |
| NamespaceId | string            | 命名空间 ID      | 必填          |
| GroupName   | string            | 分组名称         | DEFAULT_GROUP |
| ServiceName | string            | 服务名称         | 必填          |
| IP          | string            | 服务 IP          | 必填          |
| ServicePort | uint64            | 服务端口         | 必填          |
| Weight      | float64           | 权重             | 1.0           |
| Enable      | bool              | 是否启用         | true          |
| Healthy     | bool              | 是否健康         | true          |
| Ephemeral   | bool              | 是否临时实例     | true          |
| Metadata    | map[string]string | 元数据           | 可选          |
| Username    | string            | 用户名           | 可选          |
| Password    | string            | 密码             | 可选          |

### ConsulConfig

| 字段                | 类型              | 说明               | 默认值   |
| ------------------- | ----------------- | ------------------ | -------- |
| Address             | string            | Consul 地址        | 必填     |
| Scheme              | string            | 协议（http/https） | http     |
| Datacenter          | string            | 数据中心           | 可选     |
| Token               | string            | ACL Token          | 可选     |
| ServiceName         | string            | 服务名称           | 必填     |
| ServiceID           | string            | 服务 ID            | 自动生成 |
| ServiceAddress      | string            | 服务地址           | 必填     |
| ServicePort         | int               | 服务端口           | 必填     |
| ServiceTags         | []string          | 服务标签           | 可选     |
| ServiceMeta         | map[string]string | 服务元数据         | 可选     |
| HealthCheckPath     | string            | 健康检查路径       | 可选     |
| HealthCheckInterval | string            | 健康检查间隔       | 10s      |
| HealthCheckTimeout  | string            | 健康检查超时       | 5s       |
| DeregisterAfter     | string            | 注销时间           | 30s      |
| TLSConfig           | \*ConsulTLSConfig | TLS 配置           | 可选     |

---

## 高级用法

### 监听服务变化

```go
ctx := context.Background()
watchChan := reg.Watch(ctx, "/services/")

// Etcd 类型断言
if etcdReg, ok := reg.(*registry.EtcdRegistry); ok {
    watchChan := etcdReg.WatchTyped(ctx, "/services/")
    for watchResp := range watchChan {
        for _, event := range watchResp.Events {
            fmt.Printf("事件类型: %s, Key: %s\n",
                event.Type, string(event.Kv.Key))
        }
    }
}
```

### 多注册中心同时使用

```go
configs := []registry.Config{
    &registry.EtcdConfig{...},
    &registry.NacosConfig{...},
    &registry.ConsulConfig{...},
}

registries := make([]registry.Registry, 0)
for _, config := range configs {
    reg, err := factory.CreateRegistry(config)
    if err != nil {
        continue
    }
    registries = append(registries, reg)
}

// 注册到所有注册中心
for _, reg := range registries {
    reg.Publisher("192.168.1.100:8080")
}
```

### TLS 安全连接

```go
// Etcd TLS
etcdConfig := &registry.EtcdConfig{
    Hosts:              []string{"127.0.0.1:2379"},
    Key:                "/services/my-service",
    CertFile:           "/path/to/client.crt",
    CertKeyFile:        "/path/to/client.key",
    CACertFile:         "/path/to/ca.crt",
    InsecureSkipVerify: false,
}

// Consul TLS
consulConfig := &registry.ConsulConfig{
    Address:     "127.0.0.1:8500",
    Scheme:      "https",
    ServiceName: "my-service",
    TLSConfig: &registry.ConsulTLSConfig{
        CertFile: "/path/to/client.crt",
        KeyFile:  "/path/to/client.key",
        CAFile:   "/path/to/ca.crt",
    },
}
```

### 扩展新的注册中心

只需三步：

```go
// 1. 创建配置，实现 Config 接口
type ZookeeperConfig struct {
    Hosts []string
}

func (c *ZookeeperConfig) Validate() error { ... }
func (c *ZookeeperConfig) GetType() RegistryType {
    return "zookeeper"
}

// 2. 创建实现，实现 Registry 接口
type ZookeeperRegistry struct {
    // ...
}

func (z *ZookeeperRegistry) New() error { ... }
// 实现其他接口方法...

// 3. 在工厂中添加创建逻辑
func (f *RegistryFactory) createZookeeperRegistry(config *ZookeeperConfig) (Registry, error) {
    // ...
}
```

---

## 注册中心对比

### 功能对比

| 特性           | Etcd          | Nacos        | Consul      |
| -------------- | ------------- | ------------ | ----------- |
| **服务注册**   | ✅ KV 存储    | ✅ 原生支持  | ✅ 原生支持 |
| **服务发现**   | ✅ Watch 机制 | ✅ 推送+轮询 | ✅ DNS+HTTP |
| **健康检查**   | ⚠️ 租约机制   | ✅ 心跳检测  | ✅ 多种方式 |
| **配置管理**   | ✅ KV 存储    | ✅ 原生支持  | ✅ KV 存储  |
| **一致性协议** | Raft          | Distro+Raft  | Raft        |
| **CAP 理论**   | CP            | AP+CP        | CP          |
| **多数据中心** | ❌            | ✅           | ✅          |
| **服务网格**   | ❌            | ❌           | ✅ Connect  |
| **UI 界面**    | ❌            | ✅           | ✅          |
| **权限控制**   | ✅ RBAC       | ✅           | ✅ ACL      |

### 性能对比

| 指标           | Etcd     | Nacos  | Consul   |
| -------------- | -------- | ------ | -------- |
| **写入性能**   | 中等     | 高     | 中等     |
| **读取性能**   | 高       | 高     | 高       |
| **集群规模**   | 3-7 节点 | 无限制 | 3-7 节点 |
| **服务实例数** | 数千     | 数十万 | 数千     |
| **推送延迟**   | 毫秒级   | 毫秒级 | 秒级     |
| **资源占用**   | 低       | 中     | 中       |

### 使用场景

#### Etcd 适用场景

✅ **推荐使用**

- Kubernetes 集群（官方默认）
- 配置中心为主
- 强一致性要求
- 小规模服务（< 1000 实例）
- 已有 Kubernetes 环境

❌ **不推荐使用**

- 大规模服务注册（> 10000 实例）
- 需要多数据中心
- 需要丰富的 UI 界面
- 需要服务网格功能

#### Nacos 适用场景

✅ **推荐使用**

- Spring Cloud 微服务
- 阿里云环境
- 大规模服务注册
- 需要配置管理
- 需要友好的 UI
- 多语言环境

❌ **不推荐使用**

- Kubernetes 原生环境
- 需要强一致性
- 资源受限环境

#### Consul 适用场景

✅ **推荐使用**

- 服务网格需求
- 多数据中心
- 需要健康检查
- 需要 DNS 服务发现
- HashiCorp 技术栈

❌ **不推荐使用**

- 大规模配置管理
- 需要实时推送
- 资源受限环境

### 部署复杂度

| 注册中心   | 部署难度 | 集群规模 | 外部依赖    | 运维难度 |
| ---------- | -------- | -------- | ----------- | -------- |
| **Etcd**   | 简单     | 3-5 节点 | 无          | 低       |
| **Nacos**  | 中等     | 可扩展   | MySQL(可选) | 中       |
| **Consul** | 简单     | 3-5 节点 | 无          | 低       |

---

## 选型建议

### 按项目规模选择

| 项目规模              | 首选  | 备选   | 理由                 |
| --------------------- | ----- | ------ | -------------------- |
| 小型（< 100 服务）    | Etcd  | Consul | 部署简单，资源占用少 |
| 中型（100-1000 服务） | Nacos | Consul | 功能完善，性能良好   |
| 大型（> 1000 服务）   | Nacos | -      | 支持大规模，性能优秀 |

### 按使用场景选择

| 场景         | 首选   | 备选   | 理由                    |
| ------------ | ------ | ------ | ----------------------- |
| Kubernetes   | Etcd   | -      | 原生支持，无需额外部署  |
| Spring Cloud | Nacos  | Consul | 生态完善，集成简单      |
| 服务网格     | Consul | -      | Consul Connect 原生支持 |
| 多数据中心   | Consul | Nacos  | WAN 复制，跨区域同步    |
| 配置中心     | Nacos  | Etcd   | 功能丰富，UI 友好       |
| 简单部署     | Etcd   | Consul | 单二进制，无外部依赖    |

### 迁移建议

#### 多注册中心并存（推荐）

使用本组件的抽象工厂模式，可以同时注册到多个注册中心：

```go
// 同时注册到多个注册中心
registries := []Registry{etcdReg, nacosReg, consulReg}
for _, reg := range registries {
    reg.Publisher("192.168.1.100:8080")
}
```

#### 灰度迁移步骤

1. **数据导出**：导出现有服务列表
2. **适配代码**：修改注册逻辑
3. **灰度切换**：逐步迁移服务
4. **验证测试**：确保功能正常
5. **完全切换**：下线旧注册中心

---

## 实现状态

| 注册中心 | 状态      | 说明                                 |
| -------- | --------- | ------------------------------------ |
| Etcd     | ✅ 已完成 | 完整实现，包括租约、续约、监听等功能 |
| Nacos    | 🚧 待完成 | 接口已定义，需引入 nacos-sdk-go      |
| Consul   | 🚧 待完成 | 接口已定义，需引入 consul/api        |

### 依赖

#### 已实现

- `go.etcd.io/etcd/client/v3` - Etcd 客户端

#### 待添加

- `github.com/nacos-group/nacos-sdk-go/v2` - Nacos 客户端
- `github.com/hashicorp/consul/api` - Consul 客户端

---

## 最佳实践

### 1. 使用工厂模式创建实例

```go
// 推荐：通过工厂创建
factory := registry.NewRegistryFactory(log)
reg, err := factory.CreateRegistry(config)

// 不推荐：直接创建具体实现
reg := &registry.EtcdRegistry{...}
```

### 2. 优雅关闭

```go
// 应用启动时注册
reg.Publisher("192.168.1.100:8080")

// 应用关闭时注销
defer func() {
    if err := reg.Close(); err != nil {
        log.Errorf("关闭注册中心失败: %v", err)
    }
}()
```

### 3. 健康检查

```go
// 定期检查连接状态
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for range ticker.C {
    if !reg.IsHealthy() {
        log.Warnf("注册中心连接异常，尝试重新连接")
        reg.Refresh()
    }
}
```

### 4. 监听服务变化

```go
// 使用 Watch 实现动态服务发现
go func() {
    watchChan := reg.Watch(ctx, "/services/")
    // 处理服务变化事件
}()
```

### 5. 合理设置参数

```go
// Etcd 租约时间建议 10-30 秒
config.LeaseTTL = 15

// Consul 健康检查间隔建议 10-30 秒
config.HealthCheckInterval = "15s"

// Nacos 权重建议 1.0-10.0
config.Weight = 5.0
```

### 6. 错误处理

```go
// 注册失败时重试
maxRetries := 3
for i := 0; i < maxRetries; i++ {
    if err := reg.Publisher(addr); err == nil {
        break
    }
    time.Sleep(time.Second * time.Duration(i+1))
}
```

### 7. 生产环境配置

```go
// 启用 TLS
config.CertFile = "/path/to/cert"
config.CertKeyFile = "/path/to/key"
config.CACertFile = "/path/to/ca"

// 启用认证
config.User = "admin"
config.Pass = "password"

// 配置超时
config.DialTimeout = 10
```

---

## 注意事项

1. **Etcd** 使用租约机制，需要定期续约保持服务在线
2. **Nacos** 支持临时实例和持久化实例，临时实例会自动心跳
3. **Consul** 通过健康检查机制维护服务状态
4. 生产环境建议启用 TLS 加密连接
5. 合理设置租约时间和健康检查间隔，避免频繁注册注销
6. 多注册中心场景下，注意数据一致性问题
7. 大规模服务建议使用 Nacos，小规模可选 Etcd 或 Consul

---

## 性能优化建议

### 已实现的优化

1. ✅ 线程安全：使用 `sync.RWMutex` 保护并发访问
2. ✅ 上下文管理：支持优雅关闭和超时控制
3. ✅ 错误处理：完善的错误返回和日志记录
4. ✅ 配置验证：启动前验证配置有效性
5. ✅ 健康检查：定期检查连接状态
6. ✅ TLS 支持：支持安全连接

### 未来可优化

1. 🔄 连接池：复用连接，提高性能
2. 🔄 重试机制：网络故障时自动重试
3. 🔄 熔断器：防止雪崩效应
4. 🔄 指标监控：暴露 Prometheus 指标
5. 🔄 配置热更新：运行时更新配置
6. 🔄 服务降级：主注册中心故障时切换备用

---
**最后更新时间**: 2025-01-11
