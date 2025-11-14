# TARS 网络模块

基于腾讯 TARS 框架的 RPC 通信模块。

## 特性

- 支持 TARS 协议
- 服务注册与发现
- 自动重连机制
- 完整的统计信息
- Servant 管理

## 使用示例

### 服务器端

```go
import (
    "trunk/internal/net/tars"
    "trunk/pkg/logger"
)

log, _ := logger.NewLogger(&logger.Config{
    Level: "info",
    Console: true,
})

config := &tars.ServerConfig{
    Name: "tars-server",
    Ip: "0.0.0.0",
    Port: 10000,
    Protocol: "tcp",
    MaxConnections: 1000,
}

server := &tars.TarsNetServer{
    cnf: config,
    log: log,
}

server.New()
// 添加Servant
server.AddServant("TestApp.TestServer.TestObj", &YourServant{})
server.Start()
```

### 客户端

```go
config := &tars.ClientConfig{
    Name: "tars-client",
    Host: "localhost:10000",
    Obj: "TestApp.TestServer.TestObj",
    ReconnectEnabled: true,
    ReconnectDelay: 2 * time.Second,
    MaxReconnect: 5,
}

client := &tars.TarsNetClient{
    cnf: config,
    log: log,
}

client.New()
client.Start()

// 获取代理对象
var proxy YourProxy
client.StringToProxy(config.Obj, &proxy)
```

## 配置说明

### ServerConfig

- `Name`: 服务器名称
- `Ip`: 监听 IP 地址
- `Port`: 监听端口
- `Protocol`: 协议类型 (tcp/udp)
- `MaxConnections`: 最大连接数

### ClientConfig

- `Host`: 服务器地址
- `Obj`: TARS 对象名
- `ReconnectEnabled`: 是否启用自动重连
- `ReconnectDelay`: 重连延迟
- `MaxReconnect`: 最大重连次数
