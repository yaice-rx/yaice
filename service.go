package yaice

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/router"
	"reflect"
)

//服务运行状态
var shutdown = make(chan bool, 1)

type IService interface {
	AddRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	Listen(packet network.IPacket, network string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int
	Dial(packet network.IPacket, network string, address string, options network.IOptions) network.IClient
	Close()
}

type service struct {
	cancel     context.CancelFunc
	routerMgr  router.IRouter
	configMgr  config.IConfig
}

/**
 * @param endpoints 集群管理中心连接节点
 */
func NewService() IService {
	server := &service{
		routerMgr:  router.RouterMgr,
		configMgr:  config.ConfInstance(),
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
 * 连接网络
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param network string 网络连接方式
 * @param address string 地址
 * @param options 最大连接次数
 */
func (s *service) Dial(packet network.IPacket, network_ string, address string, options network.IOptions) network.IClient {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "kcp":
		break
	case "tcp", "tcp4", "tcp6":
		return tcp.NewClient(packet, address, options)
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

}
