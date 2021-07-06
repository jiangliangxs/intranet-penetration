package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

//启动TCP服务器,接受客户端认证
func startTcpServer() {
	//启动
	strPort := strconv.Itoa(int(ServerConf.TCPServerPort))
	addr, _ := net.ResolveTCPAddr("tcp", ":"+strPort)
	listen, err := net.ListenTCP("tcp", addr)
	//失败终止
	if err != nil {
		log.Fatalln(ERROR,"监听地址失败:",err.Error())
	}
	//启动成功
	log.Println(TCP,"监听地址成功,端口:",ServerConf.TCPServerPort,"该端口接受客户端认证链接")
	for true {
		//监听链接
		tcp, listenErr := listen.AcceptTCP()
		//链接是吧
		if listenErr != nil {
			log.Println(ERROR,"接受一个新的链接出错了额",listenErr.Error())
		}else {
			log.Println(TCP,"成功的接受了一个新的链接")
			//设置会话保持时间为1分钟,心跳包是10s
			_ = tcp.SetKeepAlivePeriod(time.Minute)
			//交给独立的处理器处理
			go tcpReadHandler(tcp)
		}
	}
}

//http服务器,用于接受用户的请求
func startHttpServer() {
	//请求处理器
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//处理请求
		forwardRequest(request.Host,writer,request)
	})
	log.Println(HTTP,"服务器started,端口:",ServerConf.HttpSeverPort,"将由该端口接受所有业务请求")
	//意外停止,则终止程序
	strPort := strconv.Itoa(int(ServerConf.HttpSeverPort))
	if e := http.ListenAndServe(":"+ strPort, nil); e != nil {
		log.Panicln(HTTP,"服务器closed",e.Error())
	}
}

//启动http服务器
func startTLSServer() {
	log.Println(HTTP,"TSL服务器started,端口:",443,"仅此端口可以使用https协议")
	//意外停止,则终止程序
	if e := http.ListenAndServeTLS(":443", ServerConf.PemPath,ServerConf.KeyPath,nil); e != nil {
		log.Panicln(HTTP,"TSL服务器closed",e.Error())
	}
}

//请求转发
func forwardRequest(host string,writer http.ResponseWriter, request *http.Request) {
	log.Println(HTTP,"来自:",host,"请求:",request.RequestURI)
	//查找请求的链接
	serverSession := AllServerSession.getSessionByDomain(host)
	if serverSession != nil {
		//写入用户请求(根据ID确定请求唯一性)
		var rId = GetId()
		log.Println(HTTP,"向Tcp通道写入请求,请求ID:",rId)
		writeErr := writeRequest(rId,request, serverSession.Conn)
		if writeErr != nil {
			//连续错误次数
			serverSession.ErrCount ++
			log.Println(ERROR,writeErr.Error())
			us := NewErrUserResponse("写入客户端通道时出错了,可能客户端已经断开了",rId)
			writeResponse(*us,&writer)
		}else {
			//设置最新请求时间
			serverSession.ErrCount = 0
			serverSession.LastRequestTime = time.Now()
			//等待响应额
			uc := make(chan UserResponse)
			AddResponseWriter(rId,uc)
			userResponse := <- uc
			//写返回数据
			log.Println(HTTP,"得到响应数据,进行数据会写,响应ID:",userResponse.Id)
			//关闭通道
			close(uc)
			writeResponse(userResponse,&writer)
		}
	}else {
		log.Println(HTTP,"没有找到对应的ServerSession")
		us := NewErrUserResponse("客户端并没有注册,请求无法转发",500)
		writeResponse(*us,&writer)
	}
}

//处理响应
func writeResponse(response UserResponse, writerR *http.ResponseWriter) {
	//写入请求头
	writer := *writerR
	for k, values := range response.Header {
		//对头部进行复制
		for i, value := range values {
			if i == 0 {
				writer.Header().Set(k,value)
			}else {
				writer.Header().Add(k,value)
			}
		}
	}
	//写入状态
	writer.WriteHeader(response.StatusCode)
	//写入请求体
	writer.Write(response.Body)
	//删除无用的请求（不存在也不会报错）
	delete(ResponseWriterMap,response.Id)
}

//写入请求
func writeRequest(rId uint64,request *http.Request, conn *net.TCPConn) error {
	userRequest := NewUserRequest(request)
	userRequest.Id = rId
	marshal,err := json.Marshal(userRequest)
	if err == nil {
		message := buildMessage(2, marshal)
		_, err = conn.Write(message)
	}
	if err != nil {
		return err
	}
	return nil
}

//客户端认证
func writeAuthResponse(message *TcpMessage,tcp *net.TCPConn) {
	body := message.MessageBody
	var loginRequest = new(LoginRequest)
	//解析请求
	err := json.Unmarshal(body, loginRequest)
	//解析成功
	var loginResponse = new(LoginResponse)
	if err == nil {
		//如果存在原理的域名
		serverSession := AllServerSession.getSessionByDomain(loginRequest.Domain)
		log.Println(TCP,"即将校验会话,客户端IP:",loginRequest.ClientIp)
		if serverSession != nil {
			log.Println(TCP,"已存在相同域名的会话")
			if serverSession.ClientIp != loginRequest.ClientIp {
				loginResponse.Status = 0
				loginResponse.Message = "该域名已经被其他的客户端使用!"
				log.Println(TCP,"因为域名被其他客户端占用,无法继续绑定,域名:",serverSession.Domain)
			} else {//同一ip的新链接,关闭旧的链接
				_ = serverSession.Conn.Close()
				serverSession.Conn  = tcp
				loginResponse.Status = 1
				loginResponse.Message = "重新建立链接成功!"
				log.Println(TCP,"重新建立链接并认证成功,域名:",serverSession.Domain)
			}
		} else { //如果完全是一个新的域名,则需要绑定
			log.Println(TCP,"首次建立链接成功")
			AllServerSession.addServer(loginRequest.Domain,loginRequest.ClientIp,tcp)
			loginResponse.Status = 1
			loginResponse.Message = "首次建立链接成功!"
			log.Println(TCP,"首次建立链接并认证成功,域名:",loginRequest.Domain)
		}
		//序列化
		marshal, err := json.Marshal(loginResponse);
		if err == nil {
			_, err = tcp.Write(buildMessage(5, marshal))
		}
	}

	if err != nil {
		log.Println(ERROR,"写回客户端认证消息失败,客户端IP："+ loginRequest.ClientIp )
	}
}

//关闭serverSession,并移除
func closeServerSession(tcp *net.TCPConn){
	for s, session := range AllServerSession.AllSession {
		if session.Conn == tcp {
			log.Println(TCP,"域名：",session.Domain,"因读取错误,被移除！")
			_ = session.Conn.Close()
			AllServerSession.removeServer(s)
		}
	}
}




