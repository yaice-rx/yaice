package network

type IServer interface {
	//添加路由
	AddRouter()
	//监听端口
	Start(port int) error
	//关闭
	Close()
}
