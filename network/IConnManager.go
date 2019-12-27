package network

type IConnManager interface {
	Add(conn IConn)
	Remove(guid string)
	Len() int64
}
