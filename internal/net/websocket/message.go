package webSocket

import "github.com/gogo/protobuf/proto"

// WebsocketMessage websocket消息
type WebsocketMessage struct {
	Type int           // 消息类型
	Data proto.Message // 消息体
	Msg  []byte        // 原始消息
}
