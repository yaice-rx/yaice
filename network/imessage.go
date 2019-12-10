package network

type IMessage interface {
	GetMsgId() uint32 //获取消息ID
	GetData() []byte  //获取消息内容
	GetConn() IConn
}
