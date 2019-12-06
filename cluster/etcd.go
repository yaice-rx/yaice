package cluster

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/yaice-rx/yaice/constant"
	"github.com/yaice-rx/yaice/resource"
	"strings"
	"sync"
	"time"
)

type IClusterDiscovery interface {
	GetData() [][]byte
	Register(data interface{}) error
	DelData(key string) error
	SetKey()
	Watch()
	Close()
}

var TTL int64 = 20

type ClusterDiscovery struct {
	sync.Mutex
	key           string
	Prefix        string
	conn          *clientv3.Client
	leaseRes      *clientv3.LeaseGrantResponse //自己配置租约
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

var ClusterEtcdMgr = newClusterDiscovery()

//初始服务发现
func newClusterDiscovery() IClusterDiscovery {
	var err error
	mgr := &ClusterDiscovery{
		Prefix: constant.ServerNamespace,
	}
	config := clientv3.Config{
		Endpoints:   strings.Split(resource.ServiceResMgr.EtcdConnectMap, ";"),
		DialTimeout: 5 * time.Second,
	}
	mgr.conn, err = clientv3.New(config)
	if nil != err {
		return nil
	}
	return mgr
}

func (this *ClusterDiscovery) SetKey() {
	this.key = constant.ServerNamespace + "/" + ClusterConfMgr.GroupId + "/" + ClusterConfMgr.TypeId
}

//注册数据到服务上
func (this *ClusterDiscovery) Register(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	this.Lock()
	defer this.Unlock()
	//保持连接的时间
	keepErr := this.grantSetLeaseKeepAlive(TTL)
	if nil != keepErr {
		return keepErr
	}
	//存储
	_, err = this.conn.Put(context.TODO(), this.key, string(jsonData), clientv3.WithLease(this.leaseRes.ID))
	if err != nil {
		return err
	}
	go this.listenLease()
	return nil
}

//删除节点
func (this *ClusterDiscovery) DelData(prefixKey string) error {
	this.Lock()
	defer this.Unlock()
	delResponse, err := this.conn.Delete(context.TODO(), prefixKey)
	if nil != err || len(delResponse.PrevKvs) <= 0 {
		return err
	}
	return nil
}

//获取节点数据
func (this *ClusterDiscovery) GetData() [][]byte {
	this.Lock()
	defer this.Unlock()
	var data [][]byte
	resp, err := this.conn.Get(context.TODO(), this.Prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	for _, value := range this.readData(resp) {
		data = append(data, value)
	}
	return data
}

//观察
func (this *ClusterDiscovery) Watch() {
	watcher := clientv3.NewWatcher(this.conn)
	for {
		rch := watcher.Watch(context.TODO(), this.Prefix, clientv3.WithPrefix())
		for response := range rch {
			for _, event := range response.Events {
				switch event.Type {
				//todo
				case mvccpb.PUT:
					//判断新服务，是否是自己所需要的
					break
				case mvccpb.DELETE:
					//删除该key后，重新连接同组服务下的新连接
					break
				}
			}
		}
	}
}

//关闭
func (this *ClusterDiscovery) Close() {
	this.conn.Close()
}

//读取节点数据
func (this *ClusterDiscovery) readData(resp *clientv3.GetResponse) [][]byte {
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
func (this *ClusterDiscovery) grantSetLeaseKeepAlive(ttl int64) error {
	response, err := this.conn.Lease.Grant(context.TODO(), ttl)
	if nil != err {
		return err
	}
	this.leaseRes = response
	aliveRes, err := this.conn.KeepAlive(context.TODO(), response.ID)
	if nil != err {
		return err
	}
	this.keepAliveChan = aliveRes
	return nil
}

//监测是否续约
func (this *ClusterDiscovery) listenLease() {
	for {
		select {
		case res := <-this.keepAliveChan:
			if nil == res {
				return
			}
			break
		}
	}
}
