package cluster

import (
	"encoding/json"
	"yaice/config"
	"yaice/network"
	"yaice/network/tcp"
	"yaice/proto"
	"yaice/resource"
	"yaice/router"
)

type IClusterServer interface {}


type ClusterServer struct {
	service network.IServer
	//服务连接集群列表	map[服务类型]map[进程id]连接句柄
	ModuleServiceMap map[string]map[string]network.IConnect
}

var ClusterServerMgr = newClusterServer()

/**
 * 初始化集群服务
 */
func newClusterServer()IClusterServer{
	mgr :=  &ClusterServer{
		service:tcp.TcpServerMgr,
		ModuleServiceMap:make(map[string]map[string]network.IConnect),
	}
	//注册服务
	router.RouterMgr.RegisterInternalRouterFunc(&proto.C2SServiceAssociate{}, func(conn network.IConnect,content []byte) {
		var data proto.C2SServiceAssociate
		if json.Unmarshal(content,&data) != nil{
			return
		}
		var connect map[string]network.IConnect
		connect[data.Pid] = conn
		mgr.ModuleServiceMap[data.TypeName] = connect
		return
	})
	//启动监听service端口
	for i := resource.ResourceConfigMgr.IntranetPortStart ; i<= resource.ResourceConfigMgr.IntranetPortEnd ; i++ {
		if mgr.service.Listen(i) != nil{
			config.ModuleConfigMgr.InPort = i
		}
	}
	mgr.service.Accept()
	return mgr
}




