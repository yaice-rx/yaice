package network

type Msg struct {
	ID   int32
	Conn IConnect
	Data []byte
}

func NewMsg(msgId int32, conn IConnect, data []byte) *Msg {
	return &Msg{
		ID:   msgId,
		Conn: conn,
		Data: data,
	}
}
