package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"time"
)

//客户端是否链接成功
var clientConnect = false

//启动客户端
func startClient() {
	openConnect()
	<-stop
	os.Exit(0)
}

//处理请求
func clientRequest(request UserRequest,tcp *net.TCPConn) {
	targetHost := ClientConf.getTargetByServerHost(request.Host)
	log.Println(HTTP,"请求ID:",request.Id,"目标地址："+targetHost,request.URL.Path)
	do, e := HttpClient.Do(request.parseToHttpRequest("http", targetHost))
	userResponse := &UserResponse{}
	if e == nil {
		userResponse = NewUserResponse(do)
		userResponse.Id = request.Id
	}else {
		//构建错误的响应消息
		userResponse = NewErrUserResponse("客户端的请求失败,请检查目标地址:"+targetHost,request.Id)
	}
	responseBytes, _ := json.Marshal(*userResponse)
	_, err := tcp.Write(buildMessage(3, responseBytes))
	log.Println(TCP,"响应ID:",userResponse.Id,"响应:",userResponse.StatusCode)
	if err != nil {
		log.Println(TCP,"写入响应是发生错误:",err.Error())
	}
}

//循环心跳发送
func lookUpHeat(tcp *net.TCPConn) {
	count ,max := 0,3
	ticker := time.NewTicker(10 * time.Second)
	for true {
		<- ticker.C
		_, err := tcp.Write(buildMessage(1,[]byte("ping")))
		if err != nil {
			clientConnect = false
			count++
			log.Println(TCP,"写入心跳出错:",err.Error())
			log.Println(TCP,"第",count,"次发送心跳包失败",max,"次则重连程序")
			if count >= max {
				log.Println(TCP,"第",count,"次发送心跳包失败,重新链接了")
				ticker.Stop()
				reconnect(tcp)
			}
		}else {
			clientConnect = true
			count = 0
		}
	}
}

//重新链接程序
func reconnect(tcp *net.TCPConn) {
	//关闭旧的链接
	_ = tcp.Close()
	reConnectCount := 0
	//重新启动
	for !clientConnect {
		reConnectCount++
		log.Println(TCP, "本地链接失败,进行重新链接,第",reConnectCount,"次链接")
		openConnect()
		time.Sleep(3 * time.Second)
	}
}

//打开链接
func openConnect() {
	//批量认证
	addr, _ := net.ResolveTCPAddr("tcp", ClientConf.ServerHost)
	tcp, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Panicln(TCP, "无法与远程服务器建立链接,请检查远程服务器是否开启")
	}else {
		log.Println(TCP, "已经初步链接,本地打开地址:",tcp.LocalAddr())
	}

	request := LoginRequest{}
	request.ClientIp = ClientConf.ClientIp
	_ = tcp.SetKeepAlivePeriod(time.Minute)
	go tcpReadHandler(tcp)
	for _, info := range ClientConf.Clients {
		request.Domain = info.Domain
		log.Println(TCP, "进行认证,认证的域名是:", request.Domain)
		requestBytes, _ := json.Marshal(request)
		_, err := tcp.Write(buildMessage(4, requestBytes))
		if err != nil {
			log.Panicln(TCP,"无法链接服务端,出错了,域名:",request.Domain)
		}
	}
	go lookUpHeat(tcp)
	clientConnect = true
}