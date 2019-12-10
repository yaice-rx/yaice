package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/proto"
	"github.com/yaice-rx/yaice/resource"
	"github.com/yaice-rx/yaice/router"
	"sync"
	"time"
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
	portChan := make(chan int)
	mgr.network.Start(portChan)
	port := <-portChan
	if port <= 0 {
		return nil
	}
	ClusterConfMgr.InHost = resource.ServiceResMgr.IntranetHost
	ClusterConfMgr.InPort = port
	close(portChan)
	return mgr
}

/**
 * 注册路由方法
 */
func (this *clusterServer) registerRouter() {
	router.RouterMgr.AddRouter(&proto.C2SServiceAssociate{}, this.serviceAssociateFunc)
	router.RouterMgr.AddRouter(&proto.C2SServicePing{}, this.servicePingFunc)
}

func (this *clusterServer) serviceAssociateFunc(conn network.IConn, content []byte) {
	this.Lock()
	defer this.Unlock()
	var data proto.C2SServiceAssociate
	if err := json.Unmarshal(content, &data); err != nil {
		fmt.Println(""+err.Error(), "====", string(content))
		return
	}
	this.network.GetConns().Add(conn)
	conn.SendMsg(&proto.S2CServiceAssociate{})
}

func (this *clusterServer) servicePingFunc(conn network.IConn, content []byte) {
	fmt.Println("ping =================== ", time.Now().String())
}
