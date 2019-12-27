package network

type IConn interface {
	Start()
	GetGuid() string
	Close()
	GetTimes() int64
	UpdateTime()
}
