package network

type IServer interface {
	Listen(packet IPacket, startPort int, endPort int, noticeHandler func(conn IConn)) int
}
