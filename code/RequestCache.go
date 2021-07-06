package main

import (
	"sync/atomic"
	"time"
)

//缓存中心
var id uint64 = 1

//GetId 为了提升并发性，给每个请求配置一个id，再得到响应时也能快速知道对应的writer
func GetId() uint64 {
	return atomic.AddUint64(&id,1)
}

//ResponseWriterMap 请求缓存池
var ResponseWriterMap = map[uint64] * WriterWait{}

//AddResponseWriter 存入请求
func AddResponseWriter(id uint64, ch chan UserResponse)  {
	ResponseWriterMap[id]  = & WriterWait{ch,time.Now()}
}

//WriterWait 响应通道和请求发送时间（备用，防止有部分请求丢失，导致通道未关闭，可能性极小）
type WriterWait struct {
	//响应等待通道
	ResponseChan chan UserResponse
	//请求发送时间
	SendTime time.Time
}