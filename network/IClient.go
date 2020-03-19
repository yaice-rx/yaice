package network

type IOptions interface {
	GetMax() int32
	SetMax()
}

type IClient interface {
	Connect() IConn
	Close()
}
