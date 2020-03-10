package tcp

import (
	"bytes"
	"encoding/binary"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
)

const (
	ConstMsgLength = 4 //消息长度
)

type packet struct {
}

func NewPacket() network.IPacket {
	return &packet{}
}

func (dp *packet) GetHeadLen() uint32 {
	//DataLen uint32(4字节)
	return ConstMsgLength
}

//封包
func (dp *packet) Pack(msg network.TransitData) []byte {
	msgLength := int32(len(msg.Data))
	return append(append(utils.IntToBytes(msgLength), utils.IntToBytes(msg.MsgId)...), msg.Data...)
}

//解包
func (dp *packet) Unpack(binaryData []byte) (network.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)
	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}
	//读msgID
	if err := binary.Read(dataBuff, binary.BigEndian, &msg.ID); err != nil {
		return nil, err
	}
	//读msgID
	if err := binary.Read(dataBuff, binary.BigEndian, &msg.Data); err != nil {
		return nil, err
	}
	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
