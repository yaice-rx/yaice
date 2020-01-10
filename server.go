package yaice

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/router"
	"strconv"
)

//服务运行状态
var shutdown = make(chan bool, 1)

type IServer interface {
	AddRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	RegisterNodeData(config config.Config) error
	GetNodeData(path string) []*config.Config
	WatchNodeData(eventHandler func(isAdd mvccpb.Event_EventType, config *config.Config))
	ListenTCP(startPort int, endPort int) int
	ListenKCP(startPort int, endPort int) int
	DialTCP(address string) network.IConn
	DialKCP(address string) network.IConn
	Close()
}

type server struct {
	routerMgr    router.IRouter
	clusterMgr   cluster.IManager
	config       config.Config
	connServices []string
	connEtcds    []string
}

/**
 * @param endpoints 集群管理中心连接节点
 */
func NewServer(endpoints []string) IServer {
	server := &server{
		routerMgr:  router.RouterMgr,
		clusterMgr: cluster.ManagerMgr,
		config:     config.Config{},
		connEtcds:  endpoints,
	}
	server.clusterMgr.Listen(server.connEtcds)
	return server
}

/**
 * @param message 消息传递结构体
 * @param handler func(conn network.IConn, content []byte) 网络调用函数
 */
func (s *server) AddRouter(message proto.Message, handler func(conn network.IConn, content []byte)) {
	s.routerMgr.AddRouter(message, handler)
}

/**
 * @param config 服务参数配置
 */
func (s *server) RegisterNodeData(config config.Config) error {
	return s.clusterMgr.Set(config.ServerGroup+"\\"+config.TypeId+"\\"+strconv.Itoa(config.Pid), config)
}

/**
 * @param path 获取服务的路径
 * @return 返回多个服务配置
 */
func (s *server) GetNodeData(path string) []*config.Config {
	return s.clusterMgr.Get(path)
}

/**
 * @func  监听来自集群服务的通知
 * @param 异步调用 func(回调事件，回调函数)
 */
func (s *server) WatchNodeData(eventHandler func(eventType mvccpb.Event_EventType, config *config.Config)) {
	s.clusterMgr.Watch(eventHandler)
}

/**
 * @func  监听端口
 * @param startPort 监听端口范围开始，endPort 监听端口范围结束
 */
func (s *server) ListenTCP(startPort int, endPort int) int {
	return tcp.ServerMgr.Listen(startPort, endPort)
}

/**
 * 连接网络
 * @param address 地址
 */
func (s *server) DialTCP(address string) network.IConn {
	return tcp.TCPClientMgr.Connect(address)
}

/**
 * 连接网络
 * @param address 地址
 */
func (s *server) DialKCP(address string) network.IConn {
	return tcp.TCPClientMgr.Connect(address)
}

/**
 * @func  监听端口
 * @param startPort 监听端口范围开始，endPort 监听端口范围结束
 */
func (s *server) ListenKCP(startPort int, endPort int) int {
	return tcp.ServerMgr.Listen(startPort, endPort)
}

/**
 * @func  监听端口
 * @param startPort 监听端口范围开始，endPort 监听端口范围结束
 */
func (s *server) ListenKcp(startPort int, endPort int) int {
	return -1
}

/**
 * 关闭集群服务
 */
func (s *server) Close() {
	s.clusterMgr.Close()
}
