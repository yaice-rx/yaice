package resource

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

/**
 * 资源配置
 */
type ResourceConf struct {
	sync.Mutex
	EtcdConnectMap      []string
	EtcdNamespace       string
	ExtranetPortStart   int
	ExtranetPortEnd     int
	ExtranetPingService int
	IntranetPortStart   int
	IntranetPortEnd     int
	IntranetPingService int
	MaxConnectNumber    int
	ClusterName         string
	ExcelFilePath       string
	ConfigFilePath      string
	LogFilePath         string
}

var mutex sync.Mutex

var ResourceConfMgr = newResourceConf()

//初始化资源配置
func newResourceConf() *ResourceConf {
	//初始化
	serviceConfMgr := &ResourceConf{}
	//同步
	serviceConfMgr.Lock()
	defer serviceConfMgr.Unlock()
	//读取文件数据
	fileData, err := ioutil.ReadFile("../resource/server_config.yaml")
	if err != nil {
		return nil
	}
	err = yaml.Unmarshal(fileData, serviceConfMgr)
	if err != nil {
		return nil
	}
	return serviceConfMgr
}
