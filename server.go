package yaice

import (
	"context"
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
	RegisterServeNodeData(config config.Config) error
	GetServeNodeData(path string) []*config.Config
	WatchServeNodeData(eventHandler func(isAdd mvccpb.Event_EventType, config *config.Config))
	Listen(packet network.IPacket, network string, startPort int, endPort int) int
	Dial(packet network.IPacket, network string, address string) network.IConn
	Close()
}

type server struct {
	cancel       context.CancelFunc
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
	//启动集群服务
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
func (s *server) RegisterServeNodeData(config config.Config) error {
	return s.clusterMgr.Set(config.ServerGroup+"\\"+config.TypeId+"\\"+strconv.Itoa(config.Pid), config)
}

/**
 * @param path 获取服务的路径
 * @return 返回多个服务配置
 */
func (s *server) GetServeNodeData(path string) []*config.Config {
	return s.clusterMgr.Get(path)
}

/**
 * @func  监听来自集群服务的通知
 * @param 异步调用 func(回调事件，回调函数)
 */
func (s *server) WatchServeNodeData(eventHandler func(eventType mvccpb.Event_EventType, config *config.Config)) {
	go s.clusterMgr.Watch(eventHandler)
}

/**
 * 连接网络
 * @param network 网络连接方式,address 地址
 */
func (s *server) Dial(packet network.IPacket, network_ string, address string) network.IConn {
	if packet == nil {
		packet = network.NewPacket()
	}
	switch network_ {
	case "kcp":
		break
	case "tcp", "tcp4", "tcp6":
		return tcp.TCPClientMgr.Connect(packet, address, tcp.WithMax(3))
		break
	}
	return nil
}

/**
 * @func  监听端口
 * @param network 网络连接方式,startPort 监听端口范围开始，endPort 监听端口范围结束
 */
func (s *server) Listen(packet network.IPacket, network_ string, startPort int, endPort int) int {
	if packet == nil {
		packet = network.NewPacket()
	}
	switch network_ {
	case "kcp":
		break
	case "tcp", "tcp4", "tcp6":
		return tcp.ServerMgr.Listen(packet, startPort, endPort)
		break
	}
	return 0
}

/**
 * 关闭集群服务
 */
func (s *server) Close() {
	s.clusterMgr.Close()
}
