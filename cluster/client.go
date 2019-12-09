package cluster

import (
	"encoding/json"
	"github.com/yaice-rx/yaice/job"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	proto_ "github.com/yaice-rx/yaice/proto"
	"github.com/yaice-rx/yaice/router"
	"sync"
)

//集群-客户端
type IClusterClient interface {
}

//集群-客户端
type clusterClient struct {
	sync.Mutex
	network network.IClient
}

//集群客户端
var ClusterClientMgr = newClusterClient()

// 初始化客户端
func newClusterClient() IClusterClient {
	this := &clusterClient{
		network: tcp.TCPClientMgr,
	}
	this.registerRouter()
	this.connectServices()
	return this
}

// 连接服务
func (this *clusterClient) connectServices() {
	for _, data := range ClusterEtcdMgr.GetData() {
		//集群配置文件
		var config clusterConf
		if json.Unmarshal(data, &config) != nil {
			continue
		}
		//首先判读服务是否属于自己
		if config.TypeId == ClusterConfMgr.TypeId {
			continue
		}
		if config.AllowConnect {
			//连接服务句柄
			conn := this.network.Connect(config.InHost, config.InPort)
			if conn == nil {
				continue
			}
			//发送服务关联协议数据
			protoData := proto_.C2SServiceAssociate{TypeName: ClusterConfMgr.TypeId, Pid: int64(ClusterConfMgr.Pid)}
			err := conn.Send(&protoData)
			if err != nil {
				continue
			}
		}
	}
}

// 注册路由方法
func (this *clusterClient) registerRouter() {
	router.RouterMgr.RegisterRouterFunc(&proto_.S2CServiceAssociate{}, this.serviceAssociateFunc)
}

func (this *clusterClient) serviceAssociateFunc(conn network.IConn, content []byte) {
	var data proto_.C2SServiceAssociate
	if json.Unmarshal(content, &data) != nil {
		return
	}
	//添加服务到列表中
	this.network.GetConns().Add(conn)
	//心跳
	job.Crontab.AddCronTask(10, -1, func() {
		data := proto_.C2SServicePing{}
		if conn.Send(&data) != nil {
			return
		}
	})
}
