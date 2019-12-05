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
	EtcdConnectMap      string `yaml:"EtcdConnectString"`
	EtcdNamespace       string `yaml:"EtcdNameSpace"`
	PortStart           int    `yaml:"PortStart"`
	PortEnd             int    `yaml:"PortEnd"`
	HttpPort            int    `yaml:"HttpPort"`
	ClusterName         string `yaml:"ClusterName"`
	IntranetHost        string `yaml:"IntranetHost"`
	ExtranetHost        string `yaml:"ExtranetHost"`
	IntranetPingService int    `yaml:"IntranetPingService"`
	ExtranetPingService int    `yaml:"ExtranetPingService"`
	MaxConnectNumber    int    `yaml:"MaxConnectNumber"`
	ExcelFilePath       string `yaml:"ExcelFilePath"`
	LogFilePath         string `yaml:"LogFilePath"`
}

var mutex sync.Mutex

var ServiceResMgr = newServiceRes()

//初始化资源配置
func newServiceRes() *ServiceResource {
	//初始化
	this := ServiceResource{}
	//同步
	mutex.Lock()
	defer mutex.Unlock()
	//读取文件数据
	fileData, err := ioutil.ReadFile("./resource/server_config.yaml")
	if err != nil {
		return nil
	}
	err = yaml.Unmarshal(fileData, &this)
	if err != nil {
		return nil
	}
	return &this
}
