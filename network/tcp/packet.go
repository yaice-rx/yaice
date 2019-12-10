package tcp

import (
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
)

const (
	ConstProtocolLength = 4
	ConstDataLength     = 4
)

type Packet struct {
}

func NewPacket() network.IPacket {
	return &Packet{}
}

//获取包头长度方法
func (dp *Packet) GetHeadLen() int {
	return ConstProtocolLength + ConstDataLength
}

//封包
func (dp *Packet) Pack(msg network.IMessage) []byte {
	data := msg.GetData()
	msgLength := uint32(len(data))
	return append(append(utils.IntToBytes(msg.GetMsgId()), utils.IntToBytes(msgLength)...), data...)
}

//解包
func (dp *Packet) Unpack(conn network.IConn, buffer []byte) []byte {
	//消息长度
	length := uint32(len(buffer))
	var i uint32
	for i = 0; i < length; i = i + 1 {
		if length < i+ConstProtocolLength+ConstDataLength {
			break
		}
		msgId := utils.BytesToInt(buffer[i : i+ConstProtocolLength])
		messageLength := utils.BytesToInt(buffer[i+ConstProtocolLength : i+ConstProtocolLength+ConstDataLength])
		if length < i+ConstDataLength+ConstProtocolLength+messageLength {
			break
		}
		data := buffer[i+ConstProtocolLength+ConstDataLength : i+ConstProtocolLength+ConstDataLength+messageLength]
		router.RouterMgr.SendMsgToReadQueue(NewMessage(msgId, data, conn))
		i += ConstProtocolLength + ConstDataLength + messageLength - 1
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}
