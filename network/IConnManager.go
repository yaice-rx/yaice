package network

type IConnManager interface {
	Add(conn IConn)
	Remove(guid string)
	Close()
	Len() int64
}
