package network

type IConnectList interface {
	Add(connect IConnect)
	Remove(connect IConnect)
	Get(guid string) IConnect
	Len() int
	ClearConn()
}
