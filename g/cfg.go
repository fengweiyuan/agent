package g

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/toolkits/file"
)

type PluginConfig struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
	Git     string `json:"git"`
	LogDir  string `json:"logs"`
}

/**
 * 心跳上报配置
 */
type HeartbeatConfig struct {
	Enabled  bool   `json:"enabled"`    /*是否启动心跳上报*/
	Addr     string `json:"addr"`       /*应该是服务器地址*/
	Interval int    `json:"interval"`   /*心跳上报的频率， 如3代表3秒*/
	Timeout  int    `json:"timeout"`
}

/**
 * Transfer服务器配置
 */
type TransferConfig struct {
	Enabled  bool     `json:"enabled"`   /*是否开启，默认开启*/
	Addrs    []string `json:"addrs"`     /*无状态Transfer服务的地址，数组形式是可以设置多个*/
	Interval int      `json:"interval"`  /*上报间隔*/
	Timeout  int      `json:"timeout"`   /*上报超时时间*/
}

/**
 * 作为Http服务器，相关配置
 */
type HttpConfig struct {
	Enabled  bool   `json:"enabled"`   /*是否开启本Agent的Http服务*/
	Listen   string `json:"listen"`    /*监听地址*/
	Backdoor bool   `json:"backdoor"`  /**/
}

type CollectorConfig struct {
	IfacePrefix []string `json:"ifacePrefix"`
}

/**
 * 全局配置
 */
type GlobalConfig struct {
	Debug         bool             `json:"debug"`     /*是否开启debug模式*/
	Hostname      string           `json:"hostname"`  /*当前主机名*/
	IP            string           `json:"ip"`        /*IP*/
	Plugin        *PluginConfig    `json:"plugin"`
	Heartbeat     *HeartbeatConfig `json:"heartbeat"`
	Transfer      *TransferConfig  `json:"transfer"`  /*类似proxy服务的一个组件，agent可以先把数据上报给无状态的transfer，再由后者把数据分发到有状态的时序数据库等*/
	Http          *HttpConfig      `json:"http"`
	Collector     *CollectorConfig `json:"collector"`
	IgnoreMetrics map[string]bool  `json:"ignore"`
}

var (
	ConfigFile string
	config     *GlobalConfig          /*变量config是指针，指向结构体GlobalConfig，后者存储了全局配置*/
	lock       = new(sync.RWMutex)
)

/**
 * 该函数的作为是读取配置。
 * 读取配置过程中加了读锁，防止配置在读的过程中被修改，导致读出配置的前后部分处于不同的版本。
 */
func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

/**
 * 取得主机名
 */
func Hostname() (string, error) {
	/*从配置取得主机名，能取到就返回*/
	hostname := Config().Hostname
	if hostname != "" {
		return hostname, nil
	}
	/*如果上述无法取到主机名，那么，*/
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ERROR: os.Hostname() fail", err)
	}
	return hostname, err
}

/**
 * 取得IP
 */
func IP() string {
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	if len(LocalIp) > 0 {
		ip = LocalIp
	}

	return ip
}

/**
 * 解析配置文件的内容
 */
func ParseConfig(cfg string) {

	/*没有指定配置文件则报错退出。因为-c配置有默认值，所以这里的逻辑基本不会进入*/
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	/*指定的配置文件不存在则报错退出*/
	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	/*读出文件的内容*/
	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	/* 把cfg.json内容解析成json */
	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	/*
	 * 加一个互斥锁，为啥不是在函数的开头就加上一个互斥锁？而且这个函数貌似不是goroutine，有必要加锁吗？
	 */
	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}
