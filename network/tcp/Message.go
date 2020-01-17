package tcp

import (
	"github.com/yaice-rx/yaice/network"
)

type Message struct {
	ID    int32
	Conn  network.IConn
	Data  []byte
	count uint8
}

func NewMessage(id int32, data []byte, conn network.IConn) network.IMessage {
	return &Message{
		ID:    id,
		Data:  data,
		Conn:  conn,
		count: 0,
	}
}

//获取消息ID
func (this *Message) GetMsgId() int32 {
	return this.ID
}

//发送计数
func (this *Message) GetCount() uint8 {
	return this.count
}

//增加计数
func (this *Message) AddCount() {
	this.count += 1
}

//获取消息内容
func (this *Message) GetData() []byte {
	return this.Data
}

func (this *Message) GetConn() network.IConn {
	return this.Conn
}
