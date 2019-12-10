package yaice

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/job"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/resource"
	"github.com/yaice-rx/yaice/router"
	"os"
)

type IServer interface {
	//适配网络
	AdaptationNetwork(network string)
	//添加路由
	AddRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	//开启业务服务方法
	Serve() error
	//停止服务器方法
	Stop()
}

//服务运行状态
var running = make(chan bool, 1)

type yaice struct {
	routerMgr           router.IRouter  //路由配置
	network             network.IServer //适配网络
	job                 *job.Cron
	serviceResMgr       *resource.ServiceResource //资源配置
	clusterDiscoveryMgr cluster.IClusterDiscovery //服务发现
	clusterClientMgr    cluster.IClusterClient    //集群-客户端
	clusterServerMgr    cluster.IClusterServer    //集群-服务器
}

func NewServer(typeId string, groundId string, allowConn bool) IServer {
	server := &yaice{
		job:                 job.Crontab,
		routerMgr:           router.RouterMgr,         //路由配置
		serviceResMgr:       resource.ServiceResMgr,   //系统资源配置
		clusterDiscoveryMgr: cluster.ClusterEtcdMgr,   //服务发现
		clusterClientMgr:    cluster.ClusterClientMgr, //客户端集群
		clusterServerMgr:    cluster.ClusterServerMgr, //服务器内部
	}
	//更新服务配置文件
	cluster.ClusterConfMgr.Pid = os.Getpid()
	cluster.ClusterConfMgr.TypeId = typeId
	cluster.ClusterConfMgr.GroupId = groundId
	cluster.ClusterConfMgr.AllowConnect = allowConn
	server.clusterDiscoveryMgr.SetPrefix()
	return server
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
	default:
		break
	}
}

func (this *yaice) AddRouter(message proto.Message, handler func(conn network.IConn, content []byte)) {
	this.routerMgr.AddRouter(message, handler)
}

//启动服务
func (this *yaice) Serve() error {
	//开启网络
	if this.network != nil {
		portChan := make(chan int)
		go this.network.Start(portChan)
		port := <-portChan
		if port <= 0 {
			return errors.New("port listen fail")
		}
		close(portChan)
		cluster.ClusterConfMgr.OutHost = this.serviceResMgr.ExtranetHost
		cluster.ClusterConfMgr.Network = this.network.GetNetworkName()
		cluster.ClusterConfMgr.OutPort = port
	}
	//注册配置中心数据
	this.clusterDiscoveryMgr.Register(cluster.ClusterConfMgr)
	//退出运行
	<-running
	return nil
}

//停止服务器方法
func (this *yaice) Stop() {
}
