package network

type IMessage interface {
	GetMsgId() int64 //获取消息ID
	GetData() []byte //获取消息内容
}
