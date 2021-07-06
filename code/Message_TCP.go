package main

//TcpMessage 总消息
type TcpMessage struct {
	MessageType int8 //消息类型[1,心跳消息,2,服务端->客户端请求,3,客户端->服务端响应,4,客户端->服务端认证,5,服务端->客户端认证响应]
	MessageLength int32 //消息长度,用于消息的拆包粘包
	MessageBody []byte //消息体
}

//LoginResponse 登录响应
type LoginResponse struct {
	Status int 	//状态 0,失败,1成功
	Message string //消息
}

//LoginRequest 登录请求
type LoginRequest struct {
	//同一个域名,同一个clientIp的时候,则可以快速建立重新链接
	ClientIp string
	//这个就当做认证条件了
	Domain string
}


