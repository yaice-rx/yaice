package network

type IClient interface {
	Connect(IP string, port int) IConnect
	Stop()
}
