package tcp

type Message struct {
	ID   int64
	Data []byte
}

type MsgQueueData struct {
	Queue chan Message
}

//获取消息ID
func (this *Message) GetMsgId() int64 {
	return this.ID
}

//获取消息内容
func (this *Message) GetData() []byte {
	return this.Data
}
