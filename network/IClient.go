package network

type IClient interface {
	Connect() IClient
	Close()
}
