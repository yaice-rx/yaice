package network

type IServer interface {
	Listen(packet IPacket, startPort int, endPort int, opt IOptions) int
}
