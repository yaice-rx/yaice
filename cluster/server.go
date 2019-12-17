package cluster

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/constant"
	"github.com/yaice-rx/yaice/job"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/proto"
	"github.com/yaice-rx/yaice/resource"
	"github.com/yaice-rx/yaice/router"
	"strings"
	"sync"
	"time"
)

//集群服务治理
type IClusterServiceManager interface {
	//注册服务配置数据
	RegisterClusterServiceData(data *ClusterConf) error
	//获取服务配置
	GetClusterServiceData() []*ClusterConf
	//监听
	Watch(data chan *ClusterConf)
}

type ClusterServiceManager struct {
	sync.Mutex
	_NetworkType   network.IServer
	_Path          string
	_Prefix        string
	_Conn          *clientv3.Client
	_LeaseRes      *clientv3.LeaseGrantResponse //自己配置租约
	_KeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	_ConnManager   network.IConnManager
}

var ClusterServiceManagerMgr = _NewClusterServiceManager()

func _NewClusterServiceManager() IClusterServiceManager {
	if resource.ServiceResMgr == nil {
		logrus.Debug("服务器资源文件尚未加载")
		return nil
	}
	service := &ClusterServiceManager{
		_NetworkType: tcp.TcpServerMgr,
		_Path:        constant.ServerNamespace + "/" + ClusterConfMgr.GroupId + "/" + ClusterConfMgr.TypeId,
		_Prefix:      constant.ServerNamespace,
		_ConnManager: tcp.NewConnManager(),
	}
	config := clientv3.Config{
		Endpoints:   strings.Split(resource.ServiceResMgr.EtcdConnectMap, ";"),
		DialTimeout: 5 * time.Second,
	}
	conn, err := clientv3.New(config)
	if nil != err {
		logrus.Debug("ETCD 服务启动错误：", err.Error())
		return nil
	}
	service._Conn = conn
	//启动监听Server端口
	portChan := make(chan int)
	go service._NetworkType.Start(portChan)
	port := <-portChan
	if port <= 0 {
		logrus.Debug("集群服务器启动失败，端口已全部被占用")
		return nil
	}
	ClusterConfMgr.InPort = port
	close(portChan)
	return service
}

//注册服务配置数据
func (this *ClusterServiceManager) RegisterClusterServiceData(data *ClusterConf) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.Debug("ETCD 注册数据,序列化失败：", err.Error())
		return err
	}
	this.Lock()
	defer this.Unlock()
	//保持连接的时间
	keepErr := this.grantSetLeaseKeepAlive(constant.ClusterConnectTTL)
	if nil != keepErr {
		logrus.Debug("ETCD 设置租约续期时间失败：", err.Error())
		return keepErr
	}
	//存储
	_, err = this._Conn.Put(context.TODO(), this._Path, string(jsonData), clientv3.WithLease(this._LeaseRes.ID))
	if err != nil {
		logrus.Debug("ETCD 注册数据失败：", err.Error())
		return err
	}
	go this.listenLease()
	return nil
}

//获取服务配置数据
func (this *ClusterServiceManager) GetClusterServiceData() []*ClusterConf {
	this.Lock()
	defer this.Unlock()
	var data []*ClusterConf
	resp, err := this._Conn.Get(context.TODO(), this._Prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	for _, value := range this.readData(resp) {
		config := &ClusterConf{}
		if json.Unmarshal(value, config) != nil {
			continue
		}
		data = append(data, config)
	}
	return data
}

//检测
func (this *ClusterServiceManager) Watch(dataChan chan *ClusterConf) {
	watcher := clientv3.NewWatcher(this._Conn)
	for {
		rch := watcher.Watch(context.TODO(), this._Prefix, clientv3.WithPrefix())
		for response := range rch {
			for _, event := range response.Events {
				data := &ClusterConf{}
				json.Unmarshal(event.Kv.Value, data)
				switch event.Type {
				case mvccpb.PUT:
					dataChan <- data
					break
				case mvccpb.DELETE:
					//删除该key后，重新连接同组服务下的新连接
					break
				}
			}
		}
	}
}

//读取节点数据
func (this *ClusterServiceManager) readData(resp *clientv3.GetResponse) [][]byte {
	var data [][]byte
	if resp == nil || resp.Kvs == nil {
		return nil
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			data = append(data, v)
		}
	}
	return data
}

//授权租期，自动续约
func (this *ClusterServiceManager) grantSetLeaseKeepAlive(ttl int64) error {
	response, err := this._Conn.Lease.Grant(context.TODO(), ttl)
	if nil != err {
		return err
	}
	this._LeaseRes = response
	aliveRes, err := this._Conn.KeepAlive(context.TODO(), response.ID)
	if nil != err {
		return err
	}
	this._KeepAliveChan = aliveRes
	return nil
}

//监测是否续约
func (this *ClusterServiceManager) listenLease() {
	for {
		select {
		case res := <-this._KeepAliveChan:
			if nil == res {
				return
			}
			break
		}
	}
}

func (this *ClusterServiceManager) _RegisterMsgHandler() {
	router.RouterMgr.AddRouter(&proto.C2SServiceAssociate{}, func(conn network.IConn, content []byte) {
		var data proto.C2SServiceAssociate
		if json.Unmarshal(content, &data) != nil {
			return
		}
		//连接服务句柄
		this._ConnManager.Add(data.TypeName, conn)
		//收到消息后，每10秒钟ping一次服务
		job.Crontab.AddCronTask(10, -1, func() {
			protoData := proto.C2SServicePing{}
			err := conn.SendMsg(&protoData)
			if err != nil {
				logrus.Debug(conn, "发送消息失败，", err.Error(), protoData)
			}
		})
	})

	router.RouterMgr.AddRouter(&proto.C2SServicePing{}, func(conn network.IConn, content []byte) {
		/*if conn.GetNetworkConn().(*net.TCPConn) != nil {
			conn.GetNetworkConn().(*net.TCPConn).SetDeadline(time.Now().Add(time.Duration(20) * time.Second))
		}*/
	})
}
