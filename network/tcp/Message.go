package tcp

type Message struct {
	ID   int32
	Data []byte
}

//获取消息ID
func (this *Message) GetMsgId() int32 {
	return this.ID
}

//获取消息内容
func (this *Message) GetData() []byte {
	return this.Data
}
