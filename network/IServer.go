package network

type IServer interface {
	Listen(startPort int, endPort int) int
}
