package rpc

import (
	"google.golang.org/grpc"
	"sync"
)

type IRPC interface {
	RegisterServerRPC(func_ func(s *grpc.Server))
	CallServerRPCFunc(s *grpc.Server)
	RegisterClientRPC(func_ func(s *grpc.ClientConn))
	CallClientRPCFunc(s *grpc.ClientConn)
}

type rpc struct {
	sync.Mutex
	rpcServerFunc func(s *grpc.Server)
	rpcClientFunc func(s *grpc.ClientConn)
}

var RPCMgr = _NewRPC()

func _NewRPC() IRPC {
	return &rpc{}
}

func (r *rpc) RegisterServerRPC(func_ func(s *grpc.Server)) {
	r.Lock()
	defer r.Unlock()
	r.rpcServerFunc = func_
}

func (r *rpc) CallServerRPCFunc(s *grpc.Server) {
	r.rpcServerFunc(s)
}

func (r *rpc) RegisterClientRPC(func_ func(s *grpc.ClientConn)) {
	r.Lock()
	defer r.Unlock()
	r.rpcClientFunc = func_
}

func (r *rpc) CallClientRPCFunc(c *grpc.ClientConn) {
	r.rpcClientFunc(c)
}
