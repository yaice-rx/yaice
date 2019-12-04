package network

type IServer interface {
	//获取网络连接名称
	GetNetwork() string
	//监听端口
	Start() (int, error)
	//关闭
	Close()
}
