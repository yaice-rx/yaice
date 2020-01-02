package GRPCManager

import "google.golang.org/grpc"

type IGRPCConn interface {
	GetTypeId() string
	GetConn() *grpc.ClientConn
}

type GRPCConn struct {
	conn        *grpc.ClientConn
	Pid         int    //服务进程编号
	TypeId      string //服务类型
	ServerGroup string //服务分组
}

func NewGRPCConn(conn *grpc.ClientConn) IGRPCConn {
	return &GRPCConn{
		conn: conn,
	}
}

func (g *GRPCConn) GetTypeId() string {
	return g.TypeId
}

func (g *GRPCConn) GetConn() *grpc.ClientConn {
	return g.conn
}
