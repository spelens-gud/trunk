# QUIC 网络模块

基于 QUIC 协议的高性能网络通信模块，提供可靠的 UDP 传输。

## 特性

- 基于 QUIC 协议的可靠传输
- 支持多路复用流
- 自动重连机制
- 连接数限制
- TLS 加密支持
- 完整的统计信息

## 使用示例

### 服务器端

```go
import (
    "crypto/tls"
    "trunk/internal/net/quic"
    "trunk/pkg/logger"
)

log, _ := logger.NewLogger(&logger.Config{
    Level: "info",
    Console: true,
})

config := &quic.ServerConfig{
    Name: "quic-server",
    Ip: "0.0.0.0",
    Port: 8443,
    TLSConfig: &tls.Config{
        // TLS配置
    },
    MaxConnections: 1000,
    IdleTimeout: 30 * time.Second,
    OnData: func(c conn.IConn, data []byte) error {
        // 处理数据
        return nil
    },
}

server := &quic.QuicNetServer{
    cnf: config,
    log: log,
}

server.New()
server.Start()
```

### 客户端

```go
config := &quic.ClientConfig{
    NetConfig: conn.NetConfig[quic.Connection]{
        Name: "quic-client",
        Host: "localhost:8443",
    },
    TLSConfig: &tls.Config{
        InsecureSkipVerify: true,
    },
    ReconnectEnabled: true,
    ReconnectDelay: 2 * time.Second,
    MaxReconnect: 5,
}

client := &quic.QuicNetClient{
    cnf: config,
    log: log,
}

client.New()
client.Start()
```

## 配置说明

### ServerConfig

- `Name`: 服务器名称
- `Ip`: 监听 IP 地址
- `Port`: 监听端口
- `TLSConfig`: TLS 配置
- `MaxConnections`: 最大连接数
- `IdleTimeout`: 空闲超时时间
- `MaxStreamCount`: 最大流数量

### ClientConfig

- `Host`: 服务器地址
- `TLSConfig`: TLS 配置
- `ReconnectEnabled`: 是否启用自动重连
- `ReconnectDelay`: 重连延迟
- `MaxReconnect`: 最大重连次数
