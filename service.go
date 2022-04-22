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
	"reflect"
	"strconv"
)

//服务运行状态
var shutdown = make(chan bool, 1)

type IService interface {
	AddRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	RegisterServeNodeData() error
	GetServeNodeData(path string) []config.IConfig
	WatchServeNodeData(eventHandler func(isAdd mvccpb.Event_EventType, key []byte, value config.IConfig))
	Listen(packet network.IPacket, network string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int
	Dial(packet network.IPacket, network string, address string, options network.IOptions) network.IConn
	Close()
}

type service struct {
	cancel     context.CancelFunc
	routerMgr  router.IRouter
	clusterMgr cluster.IManager
	configMgr  config.IConfig
	connEtcds  []string
}

/**
 * @param endpoints 集群管理中心连接节点
 */
func NewService(endpoints []string) IService {
	server := &service{
		routerMgr:  router.RouterMgr,
		clusterMgr: cluster.ManagerMgr,
		configMgr:  config.ConfInstance(),
		connEtcds:  endpoints,
	}
	return server
}

/**
 * @param message 消息传递结构体
 * @param handler func(conn network.IConn, content []byte) 网络调用函数
 */
func (s *service) AddRouter(message proto.Message, handler func(conn network.IConn, content []byte)) {
	s.routerMgr.AddRouter(message, handler)
}

func (s *service) RegisterMQProto(mqProto interface{}, handler func(content []byte)) {
	val := reflect.Indirect(reflect.ValueOf(mqProto))
	s.routerMgr.RegisterMQ(val.Field(0).Type().Name(), handler)
}

/**
 * @param config 服务参数配置
 */
func (s *service) RegisterServeNodeData() error {
	return s.clusterMgr.Set(s.configMgr.GetServerGroup()+"\\"+s.configMgr.GetTypeId()+"\\"+strconv.FormatUint(s.configMgr.GetPid(), 10), s.configMgr)
}

/**
 * @param path 获取服务的路径
 * @return 返回多个服务配置
 */
func (s *service) GetServeNodeData(path string) []config.IConfig {
	return s.clusterMgr.Get(path)
}

/**
 * @func  监听来自集群服务的通知
 * @param 异步调用 func(回调事件，回调函数)
 */
func (s *service) WatchServeNodeData(eventHandler func(eventType mvccpb.Event_EventType, key []byte, value config.IConfig)) {
	go s.clusterMgr.Watch(eventHandler)
}

/**
 * 连接网络
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param network string 网络连接方式
 * @param address string 地址
 * @param options 最大连接次数
 */
func (s *service) Dial(packet network.IPacket, network_ string, address string, options network.IOptions) network.IConn {
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

/**
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param string network 网络连接方式
 * @param int startPort 监听端口范围开始
 * @param int endPort 监听端口范围结束
 * @param func isAllowConnFunc  限制连接数，超过连接数的时候，由上层逻辑通知，底层不予维护
 */
func (s *service) Listen(packet network.IPacket, network_ string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	//启动集群服务
	s.clusterMgr.Listen(s.connEtcds)
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
 * 关闭集群服务
 */
func (s *service) Close() {
	s.clusterMgr.Close()
}
