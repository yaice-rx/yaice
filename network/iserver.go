package network

type IServer interface {
	//获取网络连接名称
	GetNetworkName() string
	//监听端口
	Start(port chan int)
	//关闭
	Close()
}
