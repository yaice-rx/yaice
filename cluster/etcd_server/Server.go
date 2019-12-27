package etcd_server

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/constant"
	"github.com/yaice-rx/yaice/rpc"
	"google.golang.org/grpc"
	"strconv"
	"sync"
	"time"
)

type IEtcdManager interface {
	Listen(connects []string) error
	Set(key string, data cluster.Config) error
	Close()
}

type _EtcdManager struct {
	sync.Mutex
	prefix        string
	connect       string
	data          map[string]map[int]*grpc.ClientConn
	conn          *clientv3.Client
	leaseRes      *clientv3.LeaseGrantResponse //自己配置租约
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

var ServerMgr = _NewServer()

func _NewServer() IEtcdManager {
	return &_EtcdManager{
		prefix: constant.Namespace,
		data:   make(map[string]map[int]*grpc.ClientConn),
	}
}

func (s *_EtcdManager) Listen(connects []string) error {
	conn, err := clientv3.New(clientv3.Config{
		Endpoints:   connects,
		DialTimeout: 5 * time.Second,
	})
	if nil != err {
		logrus.Debug("ETCD 服务启动错误：", err.Error())
		return nil
	}
	s.conn = conn
	go s.watch()
	s.Lock()
	defer s.Unlock()
	resp, err := s.conn.Get(context.TODO(), s.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	for _, value := range s.readData(resp) {
		config := &cluster.Config{}
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if json.Unmarshal(value, config) != nil {
			continue
		}
		conn, err := grpc.Dial(config.InHost+":"+strconv.Itoa(config.InPort), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logrus.Error("grpc connect fail,error :", err)
			continue
		}
		rpc.RPCMgr.CallClientRPCFunc(conn)
		if _, ok := s.data[config.TypeId]; !ok {
			data := make(map[int]*grpc.ClientConn)
			data[config.Pid] = conn
			s.data[config.TypeId] = data
		} else {
			s.data[config.TypeId][config.Pid] = conn
		}
	}
	return nil
}

func (s *_EtcdManager) Set(key string, data cluster.Config) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.Debug("ETCD 注册数据,序列化失败：", err.Error())
		return err
	}
	s.Lock()
	defer s.Unlock()
	//保持连接的时间
	keepErr := s.grantSetLeaseKeepAlive(constant.EtcdConnectTTL)
	if nil != keepErr {
		logrus.Debug("ETCD 设置租约续期时间失败：", err.Error())
		return keepErr
	}
	//存储
	_, err = s.conn.Put(context.TODO(), s.prefix+"/"+key, string(jsonData), clientv3.WithLease(s.leaseRes.ID))
	if err != nil {
		logrus.Debug("ETCD 注册数据失败：", err.Error())
		return err
	}
	go s.listenLease()
	return nil
}

func (s *_EtcdManager) Get() map[string]map[int]*grpc.ClientConn {
	s.Lock()
	defer s.Unlock()
	return s.data
}

//检测
func (s *_EtcdManager) watch() {
	watcher := clientv3.NewWatcher(s.conn)
	for {
		rch := watcher.Watch(context.TODO(), s.prefix, clientv3.WithPrefix())
		for response := range rch {
			for _, event := range response.Events {
				config := &cluster.Config{}
				var json = jsoniter.ConfigCompatibleWithStandardLibrary
				json.Unmarshal(event.Kv.Value, config)
				switch event.Type {
				case mvccpb.PUT:
					s.Lock()
					defer s.Unlock()
					conn, err := grpc.Dial(config.InHost+":"+strconv.Itoa(config.InPort), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						logrus.Error("grpc connect fail,error :", err)
						continue
					}
					rpc.RPCMgr.CallClientRPCFunc(conn)
					if _, ok := s.data[config.TypeId]; !ok {
						data := make(map[int]*grpc.ClientConn)
						data[config.Pid] = conn
						s.data[config.TypeId] = data
					} else {
						s.data[config.TypeId][config.Pid] = conn
					}
					break
				case mvccpb.DELETE:
					//删除该key后，重新连接同组服务下的新连接
					s.Lock()
					defer s.Unlock()
					if _, ok := s.data[config.TypeId]; ok {
						s.data[config.TypeId][config.Pid].Close()
						delete(s.data[config.TypeId], config.Pid)
					}
					break
				}
			}
		}
	}
}

//读取节点数据
func (s *_EtcdManager) readData(resp *clientv3.GetResponse) [][]byte {
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
func (s *_EtcdManager) grantSetLeaseKeepAlive(ttl int64) error {
	response, err := s.conn.Lease.Grant(context.TODO(), ttl)
	if nil != err {
		return err
	}
	s.leaseRes = response
	aliveRes, err := s.conn.KeepAlive(context.TODO(), response.ID)
	if nil != err {
		return err
	}
	s.keepAliveChan = aliveRes
	return nil
}

//监测是否续约
func (s *_EtcdManager) listenLease() {
	for {
		select {
		case res := <-s.keepAliveChan:
			if nil == res {
				return
			}
			break
		}
	}
}

func (s *_EtcdManager) Close() {
	s.conn.Close()
}
