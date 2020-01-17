package network

type IMessage interface {
	GetMsgId() int32 //获取消息ID
	GetData() []byte //获取消息内容
	GetCount() uint8 //发送计数
	AddCount()       //增加计数
	GetConn() IConn
}
