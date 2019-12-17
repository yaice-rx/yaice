package network

type IConnManager interface {
	Add(typeId string, connect IConn)
	Remove(typeId string, guid string)
	GetConn(typeId string, guid string) IConn
	GetTypeConnMap(typeId string) map[string]IConn
	Len() int
	ClearConn()
}
