//TCP链接中心
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"
)


//消息处理中心,目前消息仅仅只有五种类型
func tcpReadHandler(tcp *net.TCPConn) {
	for true {
		//读取到的消息
		message, rErr := readMessage(tcp)
		//如果消息有错误,则不处理
		if rErr != nil{
			log.Println(ERROR,"读取出错了,请检查原因！")
			time.Sleep(200*time.Millisecond)
			if isServer(){
				closeServerSession(tcp)
			}
			break
		}else {
			//根据类型转发消息
			switch message.MessageType {
			case 1://心跳消息不处理

			case 2://服务端->客户端请求，客户端读取后负责请求处理
				request := &UserRequest{}
				_ = json.Unmarshal(message.MessageBody, request)
				clientRequest(*request,tcp)
			case 3://客户端->服务端响应，服务端拿到响应后，交给等到响应的writer写回用户端
				userResponse := &UserResponse{}
				_ = json.Unmarshal(message.MessageBody, userResponse)
				if wait,ok := ResponseWriterMap[userResponse.Id];ok {
					wait.ResponseChan <- *userResponse
				}
			case 4://客户端->服务端认证，在服务端注册域名后，即可订阅该域名的http请求
				writeAuthResponse(message, tcp)
			case 5://服务端->客户端认证响应，客户端拿到请求后，成功则打印，失败则终止程序
				response := &LoginResponse{}
				_ = json.Unmarshal(message.MessageBody, response)
				if response.Status == 0 {
					log.Panicln(TCP,"服务认证失败:",response.Message)
				}else {
					log.Println(TCP,"服务认证成功:",response.Message)
				}
			}
		}
	}

}

//这个方法可以完成简单的拆包问题,要求是按顺序读取
func readMessage(tcp *net.TCPConn) (*TcpMessage,error) {
	headBuf := make([]byte,5)
	_, readHeadErr := io.ReadFull(tcp, headBuf)
	if readHeadErr != nil {
		return nil,readHeadErr
	}
	//读取类型
	var messageType int8
	//读取长度
	var messageLength int32

	//获取一个消息体的类型
	bufferMessageType := bytes.NewBuffer(headBuf[:1])
	_ = binary.Read(bufferMessageType, binary.BigEndian, &messageType)

	//获取一个消息体的长度
	bufferMessageLength := bytes.NewBuffer(headBuf[1:])
	_ = binary.Read(bufferMessageLength,binary.BigEndian,&messageLength)
	msgBuf := make([]byte, messageLength)
	_, readMsgErr := io.ReadFull(tcp, msgBuf)
	if readMsgErr != nil {
		log.Println(TCP,"读取消息体发生错误:",readMsgErr.Error())
		return nil, readMsgErr
	}
	//对消息解体
	return &TcpMessage{
		MessageType: messageType,
		MessageLength: messageLength,
		MessageBody: msgBuf,
	},nil
}



//构建消息体
func buildMessage(t int8, obj []byte) []byte {
	//转类型
	messageTypeBytes := Int8ToBytes(t)
	//获取请求长度
	length := len(obj)
	length32 := int32(length)
	messageLengthBytes := Int32ToBytes(length32)
	i := append([]byte{}, messageTypeBytes...)
	i = append(i, messageLengthBytes...)
	return append(i, obj[:length]...)
}

// Int32ToBytes 将int32转成字节
func Int32ToBytes(n int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

// Int8ToBytes 将int8转成字节
func Int8ToBytes(n int8) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}