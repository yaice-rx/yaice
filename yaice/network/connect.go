package network

type IConnect interface {
	Send(data []byte) error
	Stop()
	GetGuid() string
	Receive()
}
