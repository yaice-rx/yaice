package tcp

import (
	"errors"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
)

const (
	ConstMsgIdLength  = 4 //消息Id
	ConstMsgLength    = 4 //消息长度
	ConstHeaderLength = 4 //头部长度
)

const Header = "yaice$#"

type Packet struct {
}

func NewPacket() network.IPacket {
	return &Packet{}
}

//获取包头长度方法
func (dp *Packet) GetHeadLen() int {
	return ConstMsgIdLength + ConstHeaderLength
}

//封包
func (dp *Packet) Pack(msg network.IMessage) []byte {
	data := msg.GetData()
	msgLength := len(data)
	return append(append(utils.IntToBytes(msg.GetMsgId()), utils.IntToBytes(msgLength)...), data...)
}

//解包
func (dp *Packet) Unpack(buff []byte) ([]byte, []byte, int, error) {
	length := len(buff)
	//如果包长小于header 就直接返回 因为接收的数据不完整
	if length < ConstMsgIdLength+ConstHeaderLength+ConstMsgLength {
		return buff, nil, 0, nil
	}
	if string(buff[:ConstHeaderLength]) != Header {
		return []byte{}, nil, 0, errors.New("header is not safe")
	}
	msgId := utils.BytesToInt(buff[ConstHeaderLength : ConstMsgIdLength+ConstHeaderLength])

	msgLength := utils.BytesToInt(buff[ConstMsgIdLength : ConstMsgIdLength+ConstMsgLength+ConstHeaderLength])

	if length < ConstMsgIdLength+ConstHeaderLength+ConstMsgLength+msgLength {
		return buff, nil, 0, nil
	}

	data := buff[ConstMsgIdLength+ConstHeaderLength+ConstMsgLength : ConstMsgIdLength+ConstHeaderLength+ConstMsgLength+msgLength]

	buffs := buff[ConstMsgIdLength+ConstHeaderLength+msgLength:]

	return buffs, data, msgId, nil
}
