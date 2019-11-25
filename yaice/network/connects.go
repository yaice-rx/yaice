package network

type IConnects interface {
	Add(connect IConnect)
	Remove(connect IConnect)
	Get(guid string) (IConnect)
	Len() int
	RunThread()
}
