package network

type IClient interface {
	Connect(address string) IConn
	Close()
}
