package tcp

import "github.com/yaice-rx/yaice/network"

type Message struct {
	ID   int
	Conn network.IConn
	Data []byte
}

func NewMessage(id int, data []byte, conn network.IConn) network.IMessage {
	return &Message{
		ID:   id,
		Data: data,
		Conn: conn,
	}
}

//获取消息ID
func (this *Message) GetMsgId() int {
	return this.ID
}

//获取消息内容
func (this *Message) GetData() []byte {
	return this.Data
}

func (this *Message) GetConn() network.IConn {
	return this.Conn
}
