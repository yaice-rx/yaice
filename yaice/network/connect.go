package network

type IConnect interface {
	Send(data []byte) error
	GetGuid()string
	Receive()
}

