package packates

import (
	"github.com/yaice-rx/yaice/conns"
)

type IPacket interface {
	GetHeadLen() uint32
	Pack(data ProtocolContentData, isPos int64) []byte
	Unpack(binaryData []byte) (IProtocolMessage, error, func(c conns.IConn))
}
