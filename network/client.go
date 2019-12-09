package network

type IClient interface {
	Connect(IP string, port int) IConn
	Run()
	GetConns() IConnManager
	Stop()
}
