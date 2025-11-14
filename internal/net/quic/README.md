# QUIC 网络模块

基于 QUIC 协议（quic-go）的高性能网络通信模块，提供可靠的 UDP 传输。

## 特性

- 基于 QUIC 协议的可靠传输
- 支持多路复用流
- 自动重连机制
- 连接数限制
- TLS 加密支持
- 完整的统计信息
- 低延迟、高吞吐量
- **消息分帧协议**：自动处理消息边界，解决粘包问题
- **Message 模块集成**：支持 JSON、Protobuf、原始字节等多种编解码方式

## 消息协议

本模块使用简单的长度前缀协议来处理消息边界：

```
+----------------+------------------+
| Length (4字节) | Message Content  |
+----------------+------------------+
```

- **Length**：消息长度，大端序（Big Endian），4 字节
- **Message Content**：实际消息内容
- **最大消息大小**：1MB

这个协议确保：

- 每个消息完整传输
- 自动处理 TCP 粘包问题
- 支持任意大小的消息（最大 1MB）

## 使用示例

### 基础使用

#### 服务器端

```go
import (
    "crypto/tls"
    "github.com/spelens-gud/trunk/internal/net/quic"
    "github.com/spelens-gud/logger"
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
        Certificates: []tls.Certificate{cert},
        NextProtos:   []string{"quic-trunk"},
    },
    MaxConnections: 1000,
    IdleTimeout: 30 * time.Second,
    OnData: func(c conn.IConn, data []byte) error {
        // 处理接收到的完整消息
        log.Infof("收到消息: %s", string(data))
        return nil
    },
}

server := &quic.NetQuicServer{
    cnf: config,
    log: log,
}

server.New()
server.Start()
```

#### 客户端

```go
config := &quic.ClientConfig{
    Name: "quic-client",
    Host: "localhost:8443",
    TLSConfig: &tls.Config{
        InsecureSkipVerify: true,
        NextProtos:         []string{"quic-trunk"},
    },
    ReconnectEnabled: true,
    ReconnectDelay: 2 * time.Second,
    MaxReconnect: 5,
    OnData: func(client *quic.NetQuicClient, data []byte) error {
        // 处理接收到的完整消息
        log.Infof("收到响应: %s", string(data))
        return nil
    },
}

client := &quic.NetQuicClient{
    cnf: config,
    log: log,
}

client.New()
client.Start()

// 发送消息（自动添加长度前缀）
client.Write([]byte("Hello QUIC"))
```

### 使用 Message 模块

Message 模块提供了结构化的消息编解码功能，支持消息头（协议号、服务 ID、消息 ID、序列号）和消息体。

#### 定义消息结构

```go
type UserRequest struct {
    UserID   uint64 `json:"user_id"`
    Username string `json:"username"`
    Action   string `json:"action"`
}
```

#### 服务器端使用 Message

```go
import (
    "github.com/spelens-gud/trunk/internal/net/quic"
    "github.com/spelens-gud/trunk/internal/net/message"
)

// 创建服务器
serverConfig := &quic.ServerConfig{
    Name:      "message-server",
    Ip:        "0.0.0.0",
    Port:      8443,
    TLSConfig: tlsConfig,
}

server := &quic.NetQuicServer{
    cnf: serverConfig,
    log: log,
}

// 创建 JSON 编解码器
codec := message.NewJSONCodec[UserRequest]()

// 创建 message 服务器包装器
msgServer := quic.NewMessageServer(server, codec, func(msg *message.Message[UserRequest]) error {
    header := msg.GetHeader()
    body := msg.GetBody()

    log.Infof("收到消息 - 协议:%d, 服务:%d, 消息ID:%d, 序列:%d",
        header.ProtocolID, header.ServiceID, header.MessageID, header.Sequence)
    log.Infof("用户请求 - UserID:%d, Username:%s, Action:%s",
        body.UserID, body.Username, body.Action)

    return nil
})

server.New()
server.Start()
```

#### 客户端使用 Message

```go
// 创建客户端
clientConfig := &quic.ClientConfig{
    Name: "message-client",
    Host: "localhost:8443",
    TLSConfig: &tls.Config{
        InsecureSkipVerify: true,
        NextProtos:         []string{"quic-trunk"},
    },
}

client := &quic.NetQuicClient{
    cnf: clientConfig,
    log: log,
}

// 创建 message 客户端包装器
msgClient := quic.NewMessageClient(
    client,
    codec,
    1001, // ProtocolID
    2001, // ServiceID
    3001, // MessageID
    nil,  // 响应回调（可选）
)

client.New()
client.Start()

// 发送结构化消息
request := UserRequest{
    UserID:   12345,
    Username: "alice",
    Action:   "login",
}

msgClient.SendMessage(request, 1) // 序列号为 1
```

#### 支持的编解码器

1. **JSON 编解码器**

```go
codec := message.NewJSONCodec[YourType]()
```

2. **原始字节编解码器**

```go
codec := message.NewRawCodec()
```

3. **Protobuf 编解码器**

```go
codec := message.NewProtobufCodec[YourProtoType]()
```

4. **自定义编解码器**

```go
type CustomCodec struct{}

func (c *CustomCodec) Encode(msg YourType) ([]byte, error) {
    // 自定义编码逻辑
}

func (c *CustomCodec) Decode(data []byte) (YourType, error) {
    // 自定义解码逻辑
}
```

## 配置说明

### ServerConfig

- `Name`: 服务器名称
- `Ip`: 监听 IP 地址
- `Port`: 监听端口
- `TLSConfig`: TLS 配置（必需）
- `MaxConnections`: 最大连接数
- `IdleTimeout`: 空闲超时时间
- `KeepAlivePeriod`: 保活周期
- `OnConnect`: 连接建立回调
- `OnData`: 数据处理回调（接收完整消息）
- `OnClose`: 连接关闭回调

### ClientConfig

- `Name`: 客户端名称
- `Host`: 服务器地址（格式：host:port）
- `TLSConfig`: TLS 配置（必需）
- `ReconnectEnabled`: 是否启用自动重连
- `ReconnectDelay`: 重连延迟
- `MaxReconnect`: 最大重连次数
- `PingTicker`: 心跳间隔
- `PingFunc`: 心跳函数
- `FirstPingFunc`: 首次连接心跳函数
- `OnReconnect`: 重连成功回调
- `OnDisconnect`: 断开连接回调
- `OnData`: 数据处理回调（接收完整消息）

## 注意事项

1. **TLS 配置**：QUIC 协议要求使用 TLS 1.3，必须正确配置 TLS
2. **NextProtos**：客户端和服务器的 NextProtos 必须匹配
3. **单流模式**：当前实现使用单个流进行通信，适合请求-响应模式
4. **并发写入**：Write 方法已加锁保护，支持并发调用
5. **消息分帧**：使用 4 字节长度前缀协议，自动处理消息边界
   - 每个消息前 4 字节为消息长度（大端序）
   - 最大消息大小：1MB
   - 自动处理粘包问题
6. **完整读取**：使用 `io.ReadFull` 确保读取完整的消息
7. **Message 模块**：提供结构化消息支持，包含消息头（20 字节）和消息体

## 测试

运行基础集成测试：

```bash
go test -v ./internal/net/quic/... -run TestIntegration
```

运行 Message 集成测试：

```bash
go test -v ./internal/net/quic/... -run TestIntegration_Message
```

运行所有测试：

```bash
go test -v ./internal/net/quic/...
```

## 性能

在测试环境中：

- 吞吐量：约 50 KB/s（单流）
- 延迟：< 100ms
- 支持多客户端并发连接
- 支持自动重连
- 100% 消息完整性（无丢包）

## 依赖

- github.com/quic-go/quic-go: QUIC 协议实现
- github.com/spelens-gud/logger: 日志库
- github.com/spelens-gud/trunk/internal/net/message: 消息编解码模块

## 参考

- [QUIC-GO 文档](https://github.com/quic-go/quic-go)
- [QUIC 协议规范](https://www.rfc-editor.org/rfc/rfc9000.html)
