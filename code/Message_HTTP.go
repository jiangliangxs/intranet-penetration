//报文处理的一个工具类,用户消息转换的工具类

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
)

//UserRequest 用户的请求信息
type UserRequest struct {
	Id uint64
	Method string
	URL * url.URL
	Proto string
	ProtoMajor int
	ProtoMinor int
	Header http.Header
	Body []byte
	Host string
}

//将http请求转为用户请求,可以用于字节传递,使用json等,性能要求较高的话,可使用protobuf
func (u UserRequest) parseToHttpRequest(schema string,targetHost string) *http.Request {
	//基础信息,包括请求头之类的,要转发的目标地址不需要管理,交给客户端来做
	r := &http.Request{}
	r.Method = u.Method
	r.URL = u.URL
	r.Proto = u.Proto
	r.ProtoMajor = u.ProtoMajor
	r.ProtoMinor = u.ProtoMinor
	r.Header =  u.Header
	r.URL.Scheme = schema
	r.URL.Host = targetHost
	r.Body = ioutil.NopCloser(bytes.NewReader(u.Body))
	return r
}

//NewUserRequest 类似构造函数了,构建用户请求
func NewUserRequest(r *http.Request) *UserRequest {
	//基础信息,包括请求头之类的,要转发的目标地址不需要管理,交给客户端来做
	user := &UserRequest{
		Host: r.Host,
		Method: r.Method,
		URL: r.URL,
		Proto:r.Proto,
		ProtoMajor:r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header: r.Header}
	user.Body, _ = ioutil.ReadAll(r.Body)
	return user
}

//UserResponse 用户响应,用于数据传递额
type UserResponse struct {
	Id uint64
	Status     string
	StatusCode int
	Proto      string
	ProtoMajor int
	ProtoMinor int
	Header http.Header
	Body []byte
}

//NewUserResponse 用户请求的构造函数
func NewUserResponse(r * http.Response) * UserResponse{
	user := UserResponse{
		Status:     r.Status,
		StatusCode: r.StatusCode,
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header:     r.Header,
	}
	user.Body , _ = ioutil.ReadAll(r.Body)
	return &user
}

//ParseToHttpResponse 将用户请求转为HttpResponse,其实没有什么用,先备着
func (u UserResponse) parseToHttpResponse(r * http.Response){
	r.Status = u.Status
	r.StatusCode = u.StatusCode
	r.Proto = u.Proto
	r.ProtoMajor = u.ProtoMajor
	r.ProtoMinor = u.ProtoMinor
	r.Header =  u.Header
	r.Body = ioutil.NopCloser(bytes.NewReader(u.Body))
}

//NewErrUserResponse 创建一个错误的返回消息
func NewErrUserResponse(errMsg string,id uint64) * UserResponse{
	user := UserResponse{
		StatusCode: 500,
		Status: "Internal Server Error",
	}
	user.Id = id
	user.Body = []byte(errMsg)
	return &user
}