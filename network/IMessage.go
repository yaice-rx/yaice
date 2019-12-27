package network

type IMessage interface {
	GetMsgId() int   //获取消息ID
	GetData() []byte //获取消息内容
	GetConn() IConn
}
