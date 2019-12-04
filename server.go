package yaice

import (
	"errors"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/http"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/resource"
	"github.com/yaice-rx/yaice/router"
)

type IServer interface {
	//适配网络
	AdaptationNetwork(network string)
	//开启业务服务方法
	Serve() error
	//停止服务器方法
	Stop()
}

type yaice struct {
	routerMgr           router.IRouter            //路由配置
	Network             network.IServer           //适配网络
	serviceResMgr       *resource.ServiceResource //资源配置
	serviceDiscoveryMgr cluster.IServiceDiscovery //服务发现
	clusterClientMgr    cluster.IClusterClient    //集群-客户端
	clusterServerMgr    cluster.IClusterServer    //集群-服务器
}

func NewServer() IServer {
	return &yaice{
		routerMgr:           router.RouterMgr,         //路由配置
		serviceResMgr:       resource.ServiceResMgr,   //系统资源配置
		serviceDiscoveryMgr: cluster.ClusterEtcdMgr,   //服务发现
		clusterClientMgr:    cluster.ClusterClientMgr, //客户端集群
		clusterServerMgr:    cluster.ClusterServerMgr, //服务器内部
	}
}

/**
 * 适配网络
 */
func (this *yaice) AdaptationNetwork(network string) {
	switch network {
	case "tcp":
		this.Network = tcp.TcpServerMgr
		break
	case "kcp":
		break
	case "raknet":
		break
	case "http":
		this.Network = http.HttpServerMgr
		break
	default:
		break
	}
}

//启动服务
func (this *yaice) Serve() error {
	if this.Network == nil {
		return errors.New("network is null")
	}
	return nil
}

//停止服务器方法
func (this *yaice) Stop() {

}
