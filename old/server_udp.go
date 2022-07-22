package old

import (
	"log"
	"net"
)

type UDPServer struct {
	components

	conf *Config
}

func (s *UDPServer) Start() error {
	s.engine.Use(s.OnRequest)

	return s.init()
}

// accept 一般使用cpu核数作为参数，提高处理能力
func (s *UDPServer) init() (err error) {
	var (
		bind string
		addr *net.UDPAddr
	)
	for _, bind = range s.conf.Bind {
		if addr, err = net.ResolveUDPAddr("udp", bind); err != nil {
			log.Printf("net.ResolveUDPAddr(udp, %s) error(%v)", bind, err)
			return
		}

		log.Printf("start udp listen: %s", bind)

		// 利用多线程处理连接初始化
		for i := 0; i < s.conf.Accept; i++ {
			go s.listen(addr)
		}
	}
	return
}

func (s *UDPServer) listen(addr *net.UDPAddr) {
	var (
		conn *net.UDPConn
		err  error
	)

	for {
		if conn, err = net.ListenUDP("udp", addr); err != nil {
			log.Printf("net.ListenUDP(udp, %s) error(%v)", addr.IP, err)
			return
		}

		if err = conn.SetReadBuffer(s.conf.Rcvbuf); err != nil {
			log.Printf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(s.conf.Sndbuf); err != nil {
			log.Printf("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		if s.conf.Debug {
			log.Printf("client new request ,ip: %v", conn.RemoteAddr())
		}

		// 一个goroutine处理一个连接
		go s.handle(conn)

	}
}

func (s *UDPServer) handle(conn *net.UDPConn) {
	return
}
