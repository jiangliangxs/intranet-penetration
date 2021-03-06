package main

import (
	"net"
	"sync"
	"time"
)

//ServerStorage ServerSession仓库
type ServerStorage struct {
	sync.RWMutex
	AllSession map[string] ServerSession
}

//getSessionByDomain 根据域名获取对应的客户端
func (s *ServerStorage) getSessionByDomain(domain string) *ServerSession  {
	s.RLock()
	defer s.RUnlock()
	if session,ok := s.AllSession[domain];ok {
		return &session
	}
	return nil
}


//删除已经认证的客户端
func (s *ServerStorage) removeServer(domain string)   {
	s.Lock()
	defer s.Unlock()
	delete(s.AllSession,domain)
}

//添加已经认证的客户端到仓库
func (s *ServerStorage) addServer(domain string,clientIp string,tcp *net.TCPConn) bool  {
	s.Lock()
	defer s.Unlock()
	s.AllSession[domain] = ServerSession{
		domain,clientIp,tcp,0,time.Now()}
	return true
}

//ServerSession 是已经登录成功的客户端会话,存在一个Conn对应多个Domain
type ServerSession struct {
	//域名(Http协议解析到的域名)
	Domain string
	//客户端Ip(主要是完成客户端重连问题)
	ClientIp string
	//链接通道
	Conn *net.TCPConn
	//连续错误次数(每次+1),成功请求后清零,无并发处理
	ErrCount int
	//最后一次请求时间
	LastRequestTime time.Time
}
