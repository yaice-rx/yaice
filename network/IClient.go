package network

type IOptions interface {
	GetMax() uint
}

type IClient interface {
	Connect(address string, opt IOptions) IConn
	Close()
}
