package yaice

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/rpc"
	"google.golang.org/grpc"
	"os"
)

type IServer interface {
	AddRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	AddServerRpc(handler func(s *grpc.Server))
	AddClientRpc(handler func(s *grpc.ClientConn))
	MatchNetwork(network string)
	Serve(startPort int, endPort int)
}

type server struct {
	routerMgr  router.IRouter
	rpcMgr     rpc.IRPC
	clusterMgr cluster.IServer
	serverMgr  network.IServer
	config     cluster.Config
	etcdConns  []string
}

func NewServer(typeId string, serverGroup string, etcdConns []string) IServer {
	server := &server{
		routerMgr:  router.RouterMgr,
		rpcMgr:     rpc.RPCMgr,
		clusterMgr: cluster.ServerMgr,
		config:     cluster.Config{},
		etcdConns:  etcdConns,
	}
	server.config.TypeId = typeId
	server.config.ServerGroup = serverGroup
	return server
}

func (s *server) AddRouter(message proto.Message, handler func(conn network.IConn, content []byte)) {
	s.routerMgr.AddRouter(message, handler)
}

func (s *server) AddServerRpc(handler func(s *grpc.Server)) {
	s.rpcMgr.RegisterServerRPC(handler)
}

func (s *server) AddClientRpc(handler func(s *grpc.ClientConn)) {
	s.rpcMgr.RegisterClientRPC(handler)
}

func (s *server) MatchNetwork(network string) {
	switch network {
	case "tcp":
		s.serverMgr = tcp.ServerMgr
		break
	}
}

func (s *server) Serve(startPort int, endPort int) {
	if s.serverMgr == nil {
		s.serverMgr = tcp.ServerMgr
	}
	//开启内网服务
	s.clusterMgr.Start(s.etcdConns)
	//添加配置文件
	s.config.Pid = os.Getpid()
	s.config.InPort = s.clusterMgr.Listen(startPort, endPort)
	s.config.OutPort = s.serverMgr.Listen(startPort, endPort)
	s.clusterMgr.Set(s.config)
}
