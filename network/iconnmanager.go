package network

type IConnManager interface {
	Add(connect IConn)
	Remove(connect IConn)
	Get(guid string) IConn
	Len() int
	ClearConn()
}
