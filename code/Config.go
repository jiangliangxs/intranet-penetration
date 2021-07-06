package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
)

// mode 启动方式，是服务端还是客户端
var mode string


//AllServerSession 所有服务端会话
var AllServerSession = ServerStorage{sync.RWMutex{}, map[string]ServerSession{}}

//HTTP 日志头
var HTTP  = "[-HTTP]"

//TCP 日志头
var TCP  = "[--TCP]"

//ERROR 错误日志头
var ERROR  = "[ERROR]"

//DEBUG 调试日志头
var DEBUG  = "[DEBUG]"

//ServerConf 服务端配置
var ServerConf = ServerConfig{}

//HttpClient 客户端的httpClient,用于完成请求的处理
var HttpClient = http.Client{}

//ClientConf 客户端的配置文件
var ClientConf = ClientConfig{}


//ServerConfig 服务端配置
type ServerConfig struct {
	//HTTP协议的域名访问端口
	HttpSeverPort uint `json:"httpSeverPort"`
	//和客户端长链接的TCP端口
	TCPServerPort uint `json:"tcpServerPort"`
	//是否开启TSL
	HasTSL bool `json:"hasTsl"`
	//证书路径
	PemPath string `json:"pemPath"`
	//密钥路径
	KeyPath string `json:"keyPath"`
}

//ClientConfig 客户端配置
type ClientConfig struct {
	//客户端Ip,解决立马重连的问题
	ClientIp string `json:"clientIp"`
	//远程服务器的tcp路径,链接远程服务端的
	ServerHost string `json:"serverHost"`
	//所有的客户端列表
	Clients [] ClientInfo `json:"clients"`
}

//ClientInfo 客户端信信息
type ClientInfo struct {
	//远程域名
	Domain string `json:"domain"`
	//内网转发地址
	Forward string `json:"forward"`
}

//getTargetByServerHost 根据服务端域名,找到要转发的本地地址
func (c ClientConfig) getTargetByServerHost(serverHost string) string {
	for _,v := range ClientConf.Clients {
		if v.Domain == serverHost {
			return v.Forward
		}
	}
	return serverHost
}

//解析配置,要求client.json在项目目录
func parseClientConfig() {
	//读取配置文件
	bytes, err := ioutil.ReadFile("client.json")
	if err == nil {
		//解析配置文件
		if e := json.Unmarshal(bytes, &ClientConf);e != nil{
			log.Println(e.Error())
		}
		ClientConf.ClientIp = GetLocalIP()
	}else {
		//配置文件解析失败
		log.Println(err.Error())
	}
}

//GetLocalIP 获取本机IP，client向服务端汇报本地IP
func GetLocalIP() string {
	addrS, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrS {
			// 检查ip地址判断是否回环地址
			if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}
	return "127.0.0.1"
}

func isServer() bool{
	return mode == "server"
}