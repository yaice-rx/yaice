package network

type IOptions interface {
	GetMax() uint32
	SetMax()
}

type IClient interface {
	Connect(address string, opt IOptions) IConn
	Close()
}
