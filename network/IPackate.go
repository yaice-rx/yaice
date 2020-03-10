package network

type IPacket interface {
	GetHeadLen() uint32
	Pack(TransitData) []byte
	Unpack(binaryData []byte) (IMessage, error)
}
