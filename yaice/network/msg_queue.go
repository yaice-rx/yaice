package network

type MsgQueue struct {
	MsgId int32
	Conn   IConnect
	Data   []byte
}

func NewMsgQueue(msgId int32,conn IConnect, data []byte)*MsgQueue{
	return &MsgQueue{
		MsgId:msgId,
		Conn:conn,
		Data:data,
	}
}
