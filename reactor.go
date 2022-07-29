package linker

import (
	"github.com/pkg/errors"
	"linker/pkg/poller"
	"linker/pkg/pool"
	"log"
	"net"
	"sync"
)

type MainReactor struct {
	*Engine
	*EventHandler

	poll poller.Poller

	children []*SubReactor
}

func (reactor *MainReactor) Run(protocol string, bind string) (err error) {
	reactor.Use(reactor.EventHandler.HandleRequest)

	log.Printf("%s server listen: %s\n", protocol, bind)
	var accept Acceptor
	switch protocol {
	case TCP:
		accept = NewTCPAcceptor(reactor.dispatcher)
	case UDP:
		accept = NewUDPAcceptor(reactor.dispatcher)
	case WS:

	}

	err = accept.Listen(bind)
	if err != nil {
		return
	}
	reactor.run()
	return
}

func (reactor *MainReactor) init() {
	var err error
	reactor.poll, err = poller.CreateEpoll()
	if err != nil {
		panic(errors.WithMessage(err, "create poll"))
	}
	for i := range reactor.children {
		reactor.children[i] = &SubReactor{
			core:        reactor,
			connections: make(map[int]Conn, 1024),
			fd:          make(chan int, 1000),     // 同时处理1000个连接
			workerPool:  pool.NewWorkerPool(1024), // 允许同时处理1024个请求
		}
	}
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

func (reactor *MainReactor) run() {
	for _, sub := range reactor.children {
		go sub.Polling(reactor.Engine.buildContext)
	}
	for {
		read, err := reactor.poll.Wait()
		if err != nil {
			log.Println("unable to get active socket connection from epoll:", err)
			continue
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
	workerPool  pool.Worker
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

func (reactor *SubReactor) Polling(contextBuilder func(conn Conn) (*Context, error)) {
	var (
		ctx *Context
		err error
	)
	for {
		fd := <-reactor.fd
		conn := reactor.GetConn(fd)
		if conn == nil {
			continue
		}

		ctx, err = contextBuilder(conn)
		if err != nil {
			conn.Close()
			continue
		}
		// 读取数据不能放在协程里执行
		reactor.workerPool.Schedule(ctx.Run)
	}
}

func (reactor *SubReactor) Offer(fd int) {
	select {
	case reactor.fd <- fd:
	default:
	}

}
