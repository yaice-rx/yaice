package network

type IMessage interface {
	GetMsgId() int32 //获取消息ID
	GetData() []byte //获取消息内容
}
