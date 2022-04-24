package network

type IClient interface {
	Connect() IConn
	Close()
}
