package core

import "context"

// Message 消息接口（核心定义）
type Message interface {
	GetMsgId() int32
	GetData() []byte
	GetConnection() ConnectionInfo
}

// GlobalMessage 全局消息实现
type GlobalMessage struct {
	MsgId      int32
	Data       []byte
	Connection ConnectionInfo
}

func (m *GlobalMessage) GetMsgId() int32 {
	return m.MsgId
}

func (m *GlobalMessage) GetData() []byte {
	return m.Data
}

func (m *GlobalMessage) GetConnection() ConnectionInfo {
	return m.Connection
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	Handle(ctx context.Context, msg Message) error
}

// HandlerFunc 消息处理器函数类型
type HandlerFunc func(ctx context.Context, msg Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}
