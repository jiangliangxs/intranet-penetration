package main

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"
)

//客户端是否链接成功
var clientConnect = false

//启动客户端
func startClient() {
	tcp,err := openConnect()
	if err != nil{
		log.Panicln(ERROR,"客户端启动失败!",err.Error())
	}else {
		//正常链接的启动心条
		go lookUpHeat(tcp)
	}
}

//处理请求
func clientRequest(request UserRequest,tcp *net.TCPConn) {
	startTime := time.Now().UnixNano()//请求开始纳秒时间戳
	targetHost := ClientConf.getTargetByServerHost(request.Host)
	log.Println(HTTP,"序号:",request.Id,"目标地址："+targetHost+request.URL.Path)
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
	requestTime := (time.Now().UnixNano() - startTime) / 1e6 //请求耗时
	log.Println(TCP,"序号:",userResponse.Id,"响应码:"+strconv.Itoa(userResponse.StatusCode),"耗时:"+strconv.Itoa(int(requestTime))+"ms")
	if err != nil {
		log.Println(TCP,"写入响应是发生错误:",err.Error())
	}
}

//循环心跳发送
func lookUpHeat(tcp *net.TCPConn) {
	count ,max := 0,3
	ticker := time.NewTicker(10 * time.Second)
	needWrite := true
	var err error = nil
	for true {
		<- ticker.C
		if needWrite {
			_, err = tcp.Write(buildMessage(1, []byte("ping")))
		}
		if err != nil {
			clientConnect = false
			count++
			if count < 3 {
				log.Println(TCP,"写入心跳出错:",err.Error())
				log.Println(TCP,"第",count,"次发送心跳包失败",max,"次则重连程序")
			} else {
				log.Println(TCP,"第",count-2,"尝试重新链接")
				tcp,err = openConnect()
				//链接成功,则重新计算失败错误次数
				if err == nil{
					count =0
					needWrite = true
				} else {
					needWrite = false
				}
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
func openConnect() (*net.TCPConn, error) {
	//批量认证
	addr, _ := net.ResolveTCPAddr("tcp", ClientConf.ServerHost)
	tcp, err := net.DialTCP("tcp", nil, addr)
	if err == nil {
		log.Println(TCP, "已经初步链接,本地打开地址:",tcp.LocalAddr())
		request := LoginRequest{}
		request.ClientIp = ClientConf.ClientIp
		_ = tcp.SetKeepAlivePeriod(time.Minute)
		go tcpReadHandler(tcp)
		for _, info := range ClientConf.Clients {
			request.Domain = info.Domain
			log.Println(TCP, "进行认证,认证的域名是:", request.Domain)
			requestBytes, _ := json.Marshal(request)
			_, err = tcp.Write(buildMessage(4, requestBytes))
			if err == nil {
				log.Println(TCP,"无法链接服务端,出错了,域名:",request.Domain)
			}
		}
	}
	return tcp,err
}