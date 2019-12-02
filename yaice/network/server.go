package network

type IServer interface {
	//监听端口
	Start(port int) error
	//关闭
	Close()
}
