package network

type IClient interface {
	Connect(IP string, port int) IConn
	GetConns() IConnManager
	Stop()
}
