# gRPC 网络模块

基于 gRPC 框架的高性能 RPC 通信模块。

## 特性

- 基于 HTTP/2 协议
- 支持流式 RPC
- 自动重连机制
- 连接状态监控
- KeepAlive 支持
- 完整的统计信息

## 使用示例

### 服务器端

```go
import (
    "trunk/internal/net/grpc"
    "trunk/pkg/logger"
)

log, _ := logger.NewLogger(&logger.Config{
    Level: "info",
    Console: true,
})

config := &grpc.ServerConfig{
    Name: "grpc-server",
    Ip: "0.0.0.0",
    Port: 50051,
    MaxConnections: 1000,
    MaxConcurrentStreams: 100,
    KeepAliveTime: 10 * time.Second,
    KeepAliveTimeout: 3 * time.Second,
}

server := &grpc.GrpcNetServer{
    cnf: config,
    log: log,
}

server.New()
// 注册服务
server.RegisterService(&YourServiceDesc, &YourServiceImpl{})
server.Start()
```

### 客户端

```go
config := &grpc.ClientConfig{
    Name: "grpc-client",
    Host: "localhost:50051",
    KeepAliveTime: 10 * time.Second,
    KeepAliveTimeout: 3 * time.Second,
    ReconnectEnabled: true,
    ReconnectDelay: 2 * time.Second,
    MaxReconnect: 5,
}

client := &grpc.GrpcNetClient{
    cnf: config,
    log: log,
}

client.New()
client.Start()

// 调用RPC方法
err := client.Invoke(ctx, "/service/method", req, resp)
```

## 配置说明

### ServerConfig

- `Name`: 服务器名称
- `Ip`: 监听 IP 地址
- `Port`: 监听端口
- `MaxConnections`: 最大连接数
- `MaxConcurrentStreams`: 最大并发流数
- `KeepAliveTime`: KeepAlive 时间间隔

### ClientConfig

- `Host`: 服务器地址
- `KeepAliveTime`: KeepAlive 时间间隔
- `KeepAliveTimeout`: KeepAlive 超时时间
- `ReconnectEnabled`: 是否启用自动重连
