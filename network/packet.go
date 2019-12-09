package network

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/utils"
)

type Msg struct {
	ID   int
	Conn IConn
	Data []byte
}

const (
	//协议占位
	ConstProtocolLength = 4
	//数据长度占位
	ConstDataLength = 4
)

func NewMsg(msgId int, conn IConn, data []byte) *Msg {
	return &Msg{
		ID:   msgId,
		Conn: conn,
		Data: data,
	}
}

//封包
func Packet(message proto.Message) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		return nil
	}
	protoNumber := utils.ProtocalNumber(utils.GetProtoName(message))
	return append(append(utils.IntToBytes(protoNumber), utils.IntToBytes(len(data))...), data...)
}

//解包
func UnPacket(conn IConn, buffer []byte, readerChannel chan *Msg) []byte {
	//消息长度
	length := len(buffer)
	var i int
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
		readerChannel <- NewMsg(msgId, conn, data)
		i += ConstProtocolLength + ConstDataLength + messageLength - 1
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}
