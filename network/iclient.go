package network

type IClient interface {
	//param : 1、地址 2、端口 3、连接服务类型
	Connect(IP string, port int) IConn
	//停止连接
	Stop()
}
