package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/proto"
	"github.com/yaice-rx/yaice/router"
	"sync"
)

// 集群服务
type IClusterServer interface {
	registerRouter()
}

type clusterServer struct {
	sync.RWMutex
	network network.IServer
}

var ClusterServerMgr = newClusterServer()

/**
 * 初始化集群服务
 */
func newClusterServer() IClusterServer {
	mgr := &clusterServer{
		network: tcp.TcpServerMgr,
	}
	//注册服务
	mgr.registerRouter()
	//启动监听service端口
	if port, err := mgr.network.Start(); err == nil {
		ClusterConfMgr.InPort = port
	}
	mgr.network.Run()
	return mgr
}

/**
 * 注册路由方法
 */
func (this *clusterServer) registerRouter() {
	router.RouterMgr.RegisterRouterFunc(&proto.C2SServiceAssociate{}, this.serviceAssociateFunc)
	router.RouterMgr.RegisterRouterFunc(&proto.C2SServicePing{}, this.servicePingFunc)
}

func (this *clusterServer) serviceAssociateFunc(conn network.IConn, content []byte) {
	this.Lock()
	defer this.Unlock()
	var data proto.C2SServiceAssociate
	if json.Unmarshal(content, &data) != nil {
		return
	}
	this.network.GetConns().Add(conn)
	conn.Send(&proto.S2CServiceAssociate{})
}

func (this *clusterServer) servicePingFunc(conn network.IConn, content []byte) {
	fmt.Println("ping =================== ")
}
