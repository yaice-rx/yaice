package yaice

import (
	"yaice/cluster"
	"yaice/config"
	"yaice/network"
	"yaice/resource"
	"yaice/router"
)

type IServer interface {
	//启动服务器方法
	Start()
	//停止服务器方法
	Stop()
	//开启业务服务方法
	Serve()
}


type yaice struct {
	//选配网络模式
	networkMgr			network.IServer
	//模块配置
	ModuleConfigMgr 	*config.ModuleConfig
	//路由配置
	RouterMgr			router.IRouter
	//系统资源配置
	ResourceConfigMgr	*resource.ResourceConfig
	//服务发现
	ServiceDiscoveryMgr cluster.IEtcd
	//客户端集群
	ClusterClientMgr	cluster.IClusterClient
	//服务器内部
	ClusterServerMgr 	cluster.IClusterServer
}


func NewServer()IServer {
	return &yaice{
		//模块配置
		ModuleConfigMgr:config.ModuleConfigMgr,
		//路由配置
		RouterMgr:router.RouterMgr,
		//系统资源配置
		ResourceConfigMgr:resource.ResourceConfigMgr,
		//服务发现
		ServiceDiscoveryMgr:cluster.ClusterEtcdMgr,
		//客户端集群
		ClusterClientMgr:cluster.ClusterClientMgr,
		//服务器内部
		ClusterServerMgr:cluster.ClusterServerMgr,
	}
}

//开启业务服务方法
func (this *yaice)Start()  {
	
}

//停止服务器方法
func (this *yaice)Stop()  {

}

//启动服务
func (this *yaice)Serve(){

}