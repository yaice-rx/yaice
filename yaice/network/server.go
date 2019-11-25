package network


type IServer interface {
	//监听端口
	Listen(port int)error
	//接收数据
	Accept()
	//关闭
	Close()
}
