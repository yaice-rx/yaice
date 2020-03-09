package tcp

import (
	"github.com/yaice-rx/yaice/network"
)

type Message struct {
	ID    int32
	Data  []byte
	count uint8
}

func NewMessage(id int32, data []byte) network.IMessage {
	return &Message{
		ID:    id,
		Data:  data,
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
