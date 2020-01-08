package cluster

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

const prefix = "yaice"
const ConnectTTL = 20

type IManager interface {
	Listen(endpoints []string) error
	Set(path string, data config.Config) error
	Get(path string) []*config.Config
	Watch(eventHandler func(eventType mvccpb.Event_EventType, config *config.Config))
	Close()
}

type manager struct {
	sync.Mutex
	prefix        string
	connect       string
	connServices  []string //需要连接的服务
	conn          *clientv3.Client
	leaseRes      *clientv3.LeaseGrantResponse //自己配置租约
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

var ManagerMgr = _NewServer()

func _NewServer() IManager {
	return &manager{
		prefix: prefix,
	}
}

func (s *manager) Listen(endpoints []string) error {
	conn, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if nil != err {
		log.AppLogger.Fatal("ETCD 服务启动错误："+err.Error(), zap.String("function", "cluster.manager.Listen"))
		return nil
	}
	s.conn = conn
	return nil
}

func (s *manager) Set(path string, data config.Config) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.Debug("ETCD 注册数据,序列化失败：", err.Error())
		return err
	}
	s.Lock()
	defer s.Unlock()
	//保持连接的时间
	keepErr := s.grantSetLeaseKeepAlive(ConnectTTL)
	if nil != keepErr {
		logrus.Debug("ETCD 设置租约续期时间失败：", err.Error())
		return keepErr
	}
	//存储
	_, err = s.conn.Put(context.TODO(), s.prefix+"\\"+path, string(jsonData), clientv3.WithLease(s.leaseRes.ID))
	if err != nil {
		logrus.Debug("ETCD 注册数据失败：", err.Error())
		return err
	}
	go s.listenLease()
	return nil
}

func (s *manager) Get(path string) []*config.Config {
	resp, err := s.conn.Get(context.TODO(), s.prefix+"\\"+path, clientv3.WithPrefix())
	if err != nil {
		log.AppLogger.Debug("数据获取失败："+err.Error(), zap.String("function", "cluster.manager.Get"))
		return nil
	}
	configMap := []*config.Config{}
	for _, value := range s.readData(resp) {
		config := &config.Config{}
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err := json.Unmarshal(value, config); err != nil {
			log.AppLogger.Debug("序列化失败："+err.Error(), zap.String("function", "cluster.manager.Get"))
			continue
		}
		configMap = append(configMap, config)
	}
	return configMap
}

//检测
func (s *manager) Watch(eventHandler func(eventType mvccpb.Event_EventType, config *config.Config)) {
	watcher := clientv3.NewWatcher(s.conn)
	for {
		rch := watcher.Watch(context.TODO(), s.prefix, clientv3.WithPrefix())
		for response := range rch {
			for _, event := range response.Events {
				config := &config.Config{}
				var json = jsoniter.ConfigCompatibleWithStandardLibrary
				json.Unmarshal(event.Kv.Value, config)
				eventHandler(event.Type, config)
			}
		}
	}
}

//读取节点数据
func (s *manager) readData(resp *clientv3.GetResponse) [][]byte {
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
func (s *manager) grantSetLeaseKeepAlive(ttl int64) error {
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
func (s *manager) listenLease() {
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

//关闭
func (s *manager) Close() {
	s.conn.Close()
}
