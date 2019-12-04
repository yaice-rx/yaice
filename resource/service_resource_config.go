package resource

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

/**
 * 资源配置
 */
type ServiceResource struct {
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
	HttpPort            int
	ClusterName         string
	ExcelFilePath       string
	ConfigFilePath      string
	LogFilePath         string
}

var mutex sync.Mutex

var ServiceResMgr = newServiceRes()

//初始化资源配置
func newServiceRes() *ServiceResource {
	//初始化
	this := &ServiceResource{}
	//同步
	this.Lock()
	defer this.Unlock()
	//读取文件数据
	fileData, err := ioutil.ReadFile("../resource/server_config.yaml")
	if err != nil {
		return nil
	}
	err = yaml.Unmarshal(fileData, this)
	if err != nil {
		return nil
	}
	return this
}
