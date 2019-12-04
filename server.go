package yaice

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/network"
	http_ "github.com/yaice-rx/yaice/network/http"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/resource"
	"github.com/yaice-rx/yaice/router"
	"net/http"
	"os"
)

type IServer interface {
	//适配网络
	AdaptationNetwork(network string)
	//添加路由
	AddRouter(message proto.Message, handler func(conn network.IConnect, content []byte))
	//添加HTTP路由
	AddHttpHandler(router string, handler func(write http.ResponseWriter, request *http.Request))
	//开启业务服务方法
	Serve() error
	//停止服务器方法
	Stop()
}

//服务运行状态
var running = make(chan bool, 1)

type yaice struct {
	routerMgr           router.IRouter            //路由配置
	network             network.IServer           //适配网络
	serviceResMgr       *resource.ServiceResource //资源配置
	clusterDiscoveryMgr cluster.IClusterDiscovery //服务发现
	clusterClientMgr    cluster.IClusterClient    //集群-客户端
	clusterServerMgr    cluster.IClusterServer    //集群-服务器
}

func NewServer(typeId string, groundId string) IServer {
	cluster.ClusterConfMgr.Pid = os.Getpid()
	cluster.ClusterConfMgr.TypeId = typeId
	cluster.ClusterConfMgr.GroupId = groundId
	return &yaice{
		routerMgr:           router.RouterMgr,         //路由配置
		serviceResMgr:       resource.ServiceResMgr,   //系统资源配置
		clusterDiscoveryMgr: cluster.ClusterEtcdMgr,   //服务发现
		clusterClientMgr:    cluster.ClusterClientMgr, //客户端集群
		clusterServerMgr:    cluster.ClusterServerMgr, //服务器内部
	}
}

//适配网络
func (this *yaice) AdaptationNetwork(network string) {
	switch network {
	case "tcp":
		this.network = tcp.TcpServerMgr
		break
	case "kcp":
		break
	case "raknet":
		break
	case "http":
		this.network = http_.HttpServerMgr
		break
	default:
		break
	}
}

func (this *yaice) AddRouter(message proto.Message, handler func(conn network.IConnect, content []byte)) {
	this.routerMgr.RegisterRouterFunc(message, handler)
}

func (this *yaice) AddHttpHandler(router string, handler func(write http.ResponseWriter, request *http.Request)) {
	this.routerMgr.RegisterHttpHandlerFunc(router, handler)
}

//启动服务
func (this *yaice) Serve() error {
	if this.network == nil {
		running <- true
	}
	//开启网络
	port, err := this.network.Start()
	if err != nil {
		running <- true
	}
	cluster.ClusterConfMgr.OutHost = this.serviceResMgr.ExtranetHost
	cluster.ClusterConfMgr.InHost = this.serviceResMgr.IntranetHost
	cluster.ClusterConfMgr.OutPort = port
	cluster.ClusterConfMgr.Network = this.network.GetNetwork()
	//注册配置中心数据
	this.clusterDiscoveryMgr.Register(cluster.ClusterConfMgr)
	//退出运行
	<-running
	return nil
}

//停止服务器方法
func (this *yaice) Stop() {

}
