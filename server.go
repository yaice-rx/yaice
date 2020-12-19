package yaice

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/queue"
	"github.com/yaice-rx/yaice/router"
)

//服务运行状态
var shutdown = make(chan bool, 1)

type IServer interface {
	//注册客户端路由
	RegisterClientRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	//监听网络
	NetworkListen(network string, startPort int, endPort int,packet network.IPacket, isAllowConnFunc func(conn interface{}) bool) int
	//连接网络
	NetworkDial(network string, address string,packet network.IPacket, options network.IOptions) network.IConn
	//关闭
	Close()
}

type Server struct {
	routerMgr 	router.IRouter
	queueMgr	*queue.MQ
	conf		config.Config
}

/**
 * @Author Yaice
 * @Date 2020/12/15 15:16
 * @param  conf  配置文件
 * @return
 */
func Run(conf  config.Config) IServer {
	server := &Server{
		conf		:conf,
		routerMgr	:router.RouterMgr,
		queueMgr	:queue.New(conf.MsgQueueConnectURL),
	}
	server.queueMgr.Open()
	return server
}

/**
 * @param message 消息传递结构体
 * @param handler func(conn network.IConn, content []byte) 网络调用函数
 */
func (s *Server) RegisterClientRouter(message proto.Message, handler func(conn network.IConn, content []byte)) {
	s.routerMgr.AddRouter(message, handler)
}

/**
 * @param string network 网络连接方式
 * @param int startPort 监听端口范围开始
 * @param int endPort 监听端口范围结束
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param func isAllowConnFunc  是否允许该连接是否连接，由上层逻辑通知，底层不予维护
 * @return int   < 0  代表启动服务出错。
 */
func (s *Server) NetworkListen(network_ string, startPort int, endPort int,packet network.IPacket, isAllowConnFunc func(conn interface{}) bool) int {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "kcp":
		break
	case "tcp", "tcp4", "tcp6":
		serverMgr := tcp.NewServer()
		return serverMgr.Listen(packet, startPort, endPort, isAllowConnFunc)
	}
	return 0
}

/**
 * 连接网络
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param network string 网络连接方式
 * @param address string 地址
 * @param options 最大连接次数
 */
func (s *Server) NetworkDial(network_ string, address string,packet network.IPacket, options network.IOptions) network.IConn {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "kcp":
		break
	case "tcp", "tcp4", "tcp6":
		clientMgr := tcp.NewClient(packet, address, options)
		return clientMgr.Connect()
	}
	return nil
}

func (s *Server)MsgQueueProducer(name string) *queue.Producer{
	mq,err := s.queueMgr.Producer(name)
	if err != nil{
		return nil
	}
	return mq
}

func (s *Server)MsgQueueConsumer(name string) *queue.Consumer{
	mq,err := s.queueMgr.Consumer(name)
	if err != nil{
		return nil
	}
	return mq
}

/**
 * 关闭集群服务
 */
func (s *Server) Close() {
}
