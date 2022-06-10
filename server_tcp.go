package linker

import (
	"bufio"
	"context"
	"linker/core"
	"linker/utils/binary"
	"log"
	"net"
	"sync"
)

type TcpServer struct {
	Callback

	engine *Engine

	conf *Config
}

func (s *TcpServer) Start() error {
	s.engine.Use(s.OnReceive)

	return s.init()
}

func (s *TcpServer) Use(handlers ...HandleFunc) {
	s.engine.Use(handlers...)
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

		// split N core accept
		for i := 0; i < s.conf.Accept; i++ {
			go s.listen(listener)
		}
	}
	return
}

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

		// 一个goroutine处理一个channel
		go s.handle(conn, r)
		if r++; r == core.MaxInt {
			r = 0
		}

	}
}

func (s *TcpServer) handle(conn *net.TCPConn, r int) {
	if s.conf.Debug {
		lAddr := conn.LocalAddr().String()
		rAddr := conn.RemoteAddr().String()
		log.Printf("start serve \"%s\" with \"%s\"", lAddr, rAddr)
	}

	var (
		connection = &TcpConnection{instance: conn, sendQueue: make(chan []byte, s.conf.QueueSize), once: new(sync.Once)}
	)

	s.OnConnect(connection)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 分发响应数据
	go connection.Dispatch(ctx)

	// 利用bufio.Scanner解决粘包问题
	scaner := connection.getScanner(s.conf.DataLength)

	// 处理接收数据
	connection.HandleRequest(scaner, s.engine)

	s.OnDisconnect(connection)
}

type TcpConnection struct {
	sendQueue chan []byte

	instance *net.TCPConn

	once *sync.Once
}

func (conn TcpConnection) IP() string {
	ip, _, _ := net.SplitHostPort(conn.instance.RemoteAddr().String())
	return ip
}

func (conn TcpConnection) Push(msg []byte) {
	// 当入列速度大于出列速度，消息将被抛弃，需要合理设置队列长度
	select {
	case conn.sendQueue <- msg:
	default:
	}
}

func (conn TcpConnection) Connection() net.Conn {
	return conn.instance
}

func (conn TcpConnection) Close() {
	conn.once.Do(func() {
		close(conn.sendQueue)
		_ = conn.instance.Close()
	})

}

// 分发数据
func (conn TcpConnection) Dispatch(ctx context.Context) {
	defer conn.Close()

	var err error
	for {
		select {
		case <-ctx.Done():
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

func (conn TcpConnection) getScanner(packetDataLength int) *bufio.Scanner {
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

func (conn TcpConnection) HandleRequest(scanner *bufio.Scanner, engine *Engine) {
	// 利用对象池实例化context,避免GC
	// 会导致内存随着连接的增加而增加
	ctxPool := engine.ContextPool()

	for {
		if !scanner.Scan() {
			log.Println("scanner failed:", scanner.Err())
			return
		}
		// 通过对象池初始化时，会导致内存缓慢上涨,直到稳定
		//ctx := &Context{engine: engine}
		ctx := ctxPool.Get().(*Context)
		ctx.Reset(scanner.Bytes(), conn)

		// 执行回调
		ctx.Run()

		// 回收
		ctxPool.Put(ctx)
	}

}
