package linker

import (
	"log"
	"net"
)

type TcpServer struct {
	components

	conf *Config
}

func (s *TcpServer) Start() error {
	s.engine.Use(s.OnRequest)

	return s.init()
}

// accept 一般使用cpu核数作为参数，提高处理能力
func (s *TcpServer) init() (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range s.conf.Bind {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			log.Printf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			log.Printf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
			return
		}

		log.Printf("start tcp listen: %s", bind)

		// 利用多线程处理连接初始化
		for i := 0; i < s.conf.Accept; i++ {
			go s.listen(listener)
		}
	}
	return
}

const (
	maxInt = 1<<31 - 1
)

func (s *TcpServer) listen(lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)

	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Printf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(s.conf.KeepAlive); err != nil {
			log.Printf("conn.SetKeepAlive() error(%v)", err)
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
		go s.handle(conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func (s *TcpServer) handle(conn *net.TCPConn, r int) {
	if s.conf.Debug {
		lAddr := conn.LocalAddr().String()
		rAddr := conn.RemoteAddr().String()
		log.Printf("start handle \"%s\" with \"%s\"", lAddr, rAddr)
	}

	// 初始化连接
	connection := NewConn(conn, s.conf.QueueSize)

	// 分发响应数据
	go connection.dispatchResponse()

	// 开启连接事件回调
	s.OnConnect(connection)

	//connection.handleRequest(s.engine.contextPool(r), connection.read)
	scanner := connection.newScanner(s.conf.DataLength)
	// 处理接收数据
	connection.handleRequest(s.engine.contextPool(r), func() ([]byte, error) {
		if !scanner.Scan() {
			return nil, scanner.Err()
		}

		return scanner.Bytes(), nil
	})

	// 关闭连接事件回调
	s.OnDisconnect(connection)
}
