package linker

import (
	"linker/core"
	"linker/pkg/poller"
	"linker/pkg/pool"
	"log"
	"net"
	"sync"
)

type MainReactor struct {
	*Engine
	*EventHandler

	poll     poller.Poller
	acceptor *core.Acceptor
	children []*SubReactor
}

func (reactor *MainReactor) Listen(protocol string, bind string) (err error) {
	log.Printf("%s server listen: %s\n", protocol, bind)
	switch protocol {
	case TCP:
		err = reactor.listenTCP(bind)
	case UDP:
	case WS:

	}

	if err != nil {
		return
	}

	reactor.Use(reactor.EventHandler.HandleRequest)

	reactor.run()
	return
}

func (reactor *MainReactor) init() {
	reactor.poll, _ = poller.CreateEpoll()
	for i := range reactor.children {
		reactor.children[i] = &SubReactor{
			core:        reactor,
			connections: make(map[int]Conn, 1024),
			fd:          make(chan int, 64),      // 同时处理64个连接
			workerPool:  pool.NewWorkerPool(100), // 能同时处理1000个请求
		}
	}
	reactor.acceptor = core.NewAcceptor(reactor.dispatcher)
}

func (reactor *MainReactor) dispatcher(conn net.Conn) {
	c := newConn(conn)
	sub := reactor.children[c.FD()%len(reactor.children)]
	if err := sub.Register(c); err != nil {
		c.Close()
		return
	}
	c.closedCallback = sub.Release
}

func (reactor *MainReactor) listenTCP(bind string) (err error) {
	var (
		addr *net.TCPAddr
	)

	if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
		log.Printf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
		return
	}

	go reactor.acceptor.AcceptTCP(addr)

	return
}

func (reactor *MainReactor) run() {
	for _, sub := range reactor.children {
		go sub.Polling(reactor.Engine.HandleRequest)
	}
	for {
		read, closed, err := reactor.poll.Wait()
		if err != nil {
			log.Println("unable to get active socket connection from epoll:", err)
			continue
		}

		// 处理已关闭的链接
		for _, fd := range closed {
			if conn := reactor.chooseSubReactor(fd).GetConn(fd); conn != nil {
				conn.Close()
			}
		}

		// 处理待读取数据的链接
		for _, fd := range read {
			reactor.chooseSubReactor(fd).Offer(fd)
		}

	}
}

func (reactor *MainReactor) chooseSubReactor(fd int) *SubReactor {
	return reactor.children[fd%len(reactor.children)]
}

type SubReactor struct {
	core *MainReactor

	rmu         sync.RWMutex
	connections map[int]Conn
	fd          chan int
	workerPool  *pool.WorkerPool
}

func (reactor *SubReactor) Register(conn Conn) error {
	fd := conn.FD()
	if err := reactor.core.poll.Add(fd); err != nil {
		return err
	}

	reactor.rmu.Lock()
	reactor.connections[fd] = conn
	reactor.core.HandleConnect(conn)
	reactor.rmu.Unlock()
	return nil
}

func (reactor *SubReactor) GetConn(fd int) Conn {
	reactor.rmu.RLock()
	conn := reactor.connections[fd]
	reactor.rmu.RUnlock()
	return conn
}

func (reactor *SubReactor) Release(conn Conn) {
	reactor.core.HandleDisconnect(conn)
	fd := conn.FD()
	if err := reactor.core.poll.Remove(fd); err != nil {
		return
	}
	delete(reactor.connections, fd)
	return
}

func (reactor *SubReactor) Polling(processor func(conn Conn)) {
	for {
		fd, ok := <-reactor.fd
		if !ok {
			return
		}
		conn := reactor.GetConn(fd)
		if conn == nil {
			continue
		}

		//reactor.workerPool.Submit(func() {
		//	processor(conn)
		//})
		processor(conn)

	}
}

func (reactor *SubReactor) Offer(fd int) {
	select {
	case reactor.fd <- fd:
	default:
	}

}
