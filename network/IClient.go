package network

type IOptions interface {
	GetMax() uint32
	SetMax()
}

type IClient interface {
	Connect(packet IPacket, address string, opt IOptions) IConn
	Close()
}
