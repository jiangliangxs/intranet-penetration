package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)


//配置解析
func init() {
	flag.StringVar(&mode,"m","client","服务模式")
	flag.UintVar(&ServerConf.HttpSeverPort,"h",8080,"http端口，用于接受http请求的端口")
	flag.UintVar(&ServerConf.TCPServerPort,"t",18888,"tcp端口，用于与客户端建立长链接的端口")
	flag.BoolVar(&ServerConf.HasTSL,"tls" ,false,"是否开启SSL，如果有该参数，则需要开启https，只能监听443端口")
	flag.StringVar(&ServerConf.PemPath,"pem" ,"server.pem","tsl的pem证书，如果是认证机构发放的，最好使用通配证书，否则只能对一个域名有效")
	flag.StringVar(&ServerConf.KeyPath,"key" ,"server.key","tsl的key密钥，如果是认证机构发放的，最好使用通配证书，否则只能对一个域名有效")
	flag.Parse()
}

//程序入口
func main()  {
	//日志设置
	log.SetFlags(log.LstdFlags)
	//监听停止信号
	stop := make(chan os.Signal)
	signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)
	if isServer() {
		log.Println("[START]","以服务端(-m server)模式启动")
		//如果配置文件和参数中都没有拿到端口,则报错
		if ServerConf.HttpSeverPort <= 0 || ServerConf.TCPServerPort <= 0 {
			log.Panicln("[START]","设置的端口有错误,请设置一个可用端口")
		}
		//启动http服务和tcp服务
		go startHttpServer()
		go startTcpServer()
		//启动ssl服务器,只支持443端口
		if ServerConf.HasTSL {
			go startTLSServer()
		}
	}else {
		log.Println("[START]","以客户端(-m client)模式启动")
		parseClientConfig()
		startClient()
	}
	<- stop
	destroy()
}

func destroy() {
	if isServer(){
		log.Println(TCP,"正在逐个关闭通道....")
		for s, session := range AllServerSession.AllSession {
			_ = session.Conn.Close()
			log.Println(TCP,"域名",s,"通道已经关闭")
		}
	}else {
		log.Println(TCP,"关闭客户端...")
	}
}







