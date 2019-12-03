package cluster

import (
	"encoding/json"
	"sync"
	"yaice/network"
	"yaice/network/tcp"
	"yaice/proto"
	"yaice/resource"
	"yaice/router"
)

/**
 * 集群服务
 */

type IClusterServer interface {
	registerRouter()
}

type clusterServer struct {
	sync.RWMutex
	service network.IServer
	//服务连接集群列表	map[服务类型]map[进程id]连接句柄
	ServiceList map[string]map[string]network.IConnect
}

var ClusterServerMgr = newClusterServer()

/**
 * 初始化集群服务
 */
func newClusterServer() IClusterServer {
	mgr := &clusterServer{
		service:     tcp.TcpServerMgr,
		ServiceList: make(map[string]map[string]network.IConnect),
	}
	//注册服务
	mgr.registerRouter()
	//启动监听service端口
	for i := resource.ServiceResMgr.IntranetPortStart; i <= resource.ServiceResMgr.IntranetPortEnd; i++ {
		if mgr.service.Start(i) != nil {
			clusterConfMgr.InPort = i
		}
	}
	return mgr
}

/**
 * 注册路由方法
 */
func (this *clusterServer) registerRouter() {
	router.RouterMgr.RegisterRouterFunc(&proto.C2SServiceAssociate{}, this.serviceAssociateFunc)
}

func (this *clusterServer) serviceAssociateFunc(conn network.IConnect, content []byte) {
	this.Lock()
	defer this.Unlock()
	var data proto.C2SServiceAssociate
	if json.Unmarshal(content, &data) != nil {
		return
	}
	var connect map[string]network.IConnect
	connect[data.Pid] = conn
	this.ServiceList[data.TypeName] = connect
}
