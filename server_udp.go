package linker

import (
	"bufio"
	"linker/utils/binary"
	"log"
	"net"
	"sync"
)

type UDPServer struct {
	Callback

	engine *Engine

	conf *Config
}

func (s *UDPServer) Start() error {
	s.engine.Use(s.OnReceive)

	return s.init()
}

func (s *UDPServer) Use(handlers ...HandleFunc) {
	s.engine.Use(handlers...)
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
	if s.conf.Debug {
		lAddr := conn.LocalAddr().String()
		rAddr := conn.RemoteAddr().String()
		log.Printf("start handle \"%s\" with \"%s\"", lAddr, rAddr)
	}

	// 初始化连接
	connection := &UDPConnection{instance: conn}
	connection.init(s.conf.QueueSize, s.conf.DataLength)

	// 分发响应数据
	go connection.dispatchResponse()

	// 开启连接事件回调
	s.OnConnect(connection)

	// 处理接收数据
	connection.handleRequest(s.engine)

	// 关闭连接事件回调
	s.OnDisconnect(connection)
}

type UDPConnection struct {
	instance *net.UDPConn

	sendQueue chan []byte    // 发送队列
	scanner   *bufio.Scanner // 读取请求数据

	once *sync.Once
	done chan struct{} // 关闭标识
}

func (conn *UDPConnection) init(sendQueueSize int, packetDataLength int) {
	conn.sendQueue = make(chan []byte, sendQueueSize)
	conn.scanner = conn.getScanner(packetDataLength)
	conn.once = new(sync.Once)
	conn.done = make(chan struct{})
}

func (conn *UDPConnection) IP() string {
	ip, _, _ := net.SplitHostPort(conn.instance.RemoteAddr().String())
	return ip
}

func (conn *UDPConnection) Push(msg []byte) {
	// 当入列速度大于出列速度，消息将被抛弃，需要合理设置队列长度
	select {
	case conn.sendQueue <- msg:
	default:
	}
}

func (conn *UDPConnection) NetConn() net.Conn {
	return conn.instance
}

// Close 关闭请求
func (conn *UDPConnection) Close() {
	conn.once.Do(func() {
		close(conn.done)
		close(conn.sendQueue)
		_ = conn.instance.Close()
	})

}

// 分发数据
func (conn *UDPConnection) dispatchResponse() {
	defer conn.Close()

	var err error
	for {
		select {
		case <-conn.done:
			return
		default:
			msg, ok := <-conn.sendQueue
			if !ok { // 队列已关闭
				return
			}

			// 写数据
			if _, err = conn.instance.Write(msg); err != nil {
				return
			}
		}
	}
}

func (conn *UDPConnection) getScanner(packetDataLength int) *bufio.Scanner {
	scan := bufio.NewScanner(conn.instance)
	if packetDataLength <= 0 {
		return scan
	}

	// 处理粘包问题：先读取包体长度
	scan.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if !atEOF && len(data) > packetDataLength {
			length := int(binary.BigEndian.Int32(data[:packetDataLength]))
			if length <= len(data) {
				return length, data[:length], nil
			}
		}
		return
	})
	return scan
}

// handleRequest 处理请求
func (conn *UDPConnection) handleRequest(engine *Engine) {
	defer conn.Close()
	// 利用对象池实例化context,避免GC
	// 会导致内存随着连接的增加而增加
	ctxPool := engine.ContextPool()

	for {
		select {
		case <-conn.done: // 退出
			return
		default:
			if !conn.scanner.Scan() {
				log.Println("scanner failed:", conn.scanner.Err())
				return
			}
			// 通过对象池初始化时，会导致内存缓慢上涨,直到稳定
			//ctx := &Context{engine: engine}
			ctx := ctxPool.Get().(*Context)
			ctx.Reset(conn.scanner.Bytes(), conn)

			// 执行回调
			ctx.Run()

			// 回收
			ctxPool.Put(ctx)
		}

	}

}
