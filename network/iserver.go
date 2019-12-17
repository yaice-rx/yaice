package network

type IServer interface {
	//获取网络连接名称
	GetNetworkName() string
	//监听端口
	Start(port chan int)
	//获取连接列表
	GetConns() IConnManager
	//关闭
	Close()
}
