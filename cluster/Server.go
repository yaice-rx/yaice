package cluster

import (
	"github.com/yaice-rx/yaice/cluster/etcd_server"
	"github.com/yaice-rx/yaice/rpc"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

type IServer interface {
	Start(conns []string) error
	Set(config Config)
	Listen(startPort int, endPort int) int
	Close()
}

type Server struct {
	etcdManger etcd_server.IEtcdManager
	server     *grpc.Server
}

var ServerMgr = _NewServerMgr()

func _NewServerMgr() IServer {
	return &Server{
		etcdManger: etcd_server.ServerMgr,
	}
}

func (s *Server) Start(conns []string) error {
	return s.etcdManger.Listen(conns)
}

func (s *Server) Set(config Config) {
	s.etcdManger.Set(config.ServerGroup+"+"+config.TypeId+"/"+strconv.Itoa(config.Pid), config)
}

func (s *Server) Listen(startPort int, endPort int) int {
	port := make(chan int)
	for i := startPort; i <= endPort; i++ {
		go func() {
			lis, err := net.Listen("tcp", ":"+strconv.Itoa(i))
			if err != nil {
				return
			}
			port <- i
			server := grpc.NewServer()
			s.server = server
			rpc.RPCMgr.CallServerRPCFunc(server)
			server.Serve(lis)
		}()
	}
	data := <-port
	close(port)
	return data
}

func (s *Server) Close() {
	s.server.Stop()
	s.etcdManger.Close()
}
