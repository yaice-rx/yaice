package packates

type IProtocolMessage interface {
	GetMsgId() int32 //获取消息ID
	GetData() []byte //获取消息内容
	GetIsPos() int64
}

type ProtocolContentData struct {
	MsgId int32
	Pos   int64
	Data  []byte
}

func (pcd *ProtocolContentData) GetMsgId() int32 {
	return pcd.MsgId
}

func (pcd *ProtocolContentData) GetData() []byte {
	return pcd.Data
}
func (pcd *ProtocolContentData) GetIsPos() int64 {
	return pcd.Pos
}
