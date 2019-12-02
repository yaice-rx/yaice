package yaice

import (
	"yaice/cluster"
	"yaice/resource"
	"yaice/router"
)

type IServer interface {
	//停止服务器方法
	Stop()
	//开启业务服务方法
	Serve()
}

type yaice struct {
	//路由配置
	RouterMgr router.IRouter
	//资源配置
	resConfMgr *resource.ResourceConf
	//服务发现
	etcdMgr cluster.IEtcd
	//集群-客户端
	clusterClientMgr cluster.IClusterClient
	//集群-服务器
	clusterServerMgr cluster.IClusterServer
}

func NewServer() IServer {
	return &yaice{
		//路由配置
		RouterMgr: router.RouterMgr,
		//系统资源配置
		resConfMgr: resource.ResourceConfMgr,
		//服务发现
		etcdMgr: cluster.ClusterEtcdMgr,
		//客户端集群
		clusterClientMgr: cluster.ClusterClientMgr,
		//服务器内部
		clusterServerMgr: cluster.ClusterServerMgr,
	}
}

/**
 * 选择网络
 */
func (this *yaice) SelectNetwork(network string) {
	switch network {
	case "tcp":

		break
	case "kcp":

		break
	default:
		break
	}
}

//停止服务器方法
func (this *yaice) Stop() {

}

//启动服务
func (this *yaice) Serve() {

}
