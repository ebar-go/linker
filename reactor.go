package linker

import (
	"linker/core"
	"linker/pkg/poller"
	"log"
	"net"
)

type MainReactor struct {
	*Engine
	*EventHandler

	poll     poller.Poller
	acceptor *core.Acceptor
	children []*SubReactor
}

func NewReactor() *MainReactor {
	reactor := &MainReactor{
		EventHandler: new(EventHandler),
		Engine:       new(Engine),
		children:     make([]*SubReactor, 16),
	}
	reactor.init()
	return reactor
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
	for i := 0; i < 16; i++ {
		reactor.children[i] = &SubReactor{
			core:        reactor,
			connections: make(map[int]Conn, 1024),
			fd:          make(chan int, 64),
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
		fds, err := reactor.poll.Wait()
		if err != nil {
			log.Println("unable to get active socket connection from epoll:", err)
			continue
		}

		for _, fd := range fds {
			sub := reactor.children[fd%len(reactor.children)]
			sub.Offer(fd)
		}

	}
}

type SubReactor struct {
	core *MainReactor

	connections map[int]Conn
	fd          chan int
}

func (reactor *SubReactor) Register(conn Conn) error {
	fd := conn.FD()
	if err := reactor.core.poll.Add(fd); err != nil {
		return err
	}

	reactor.connections[fd] = conn
	reactor.core.HandleConnect(conn)
	return nil
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

func (reactor *SubReactor) Polling(fn func(conn Conn)) {
	for {
		fd, ok := <-reactor.fd
		if !ok {
			return
		}
		conn, exist := reactor.connections[fd]
		if !exist {
			return
		}
		fn(conn)
	}
}

func (reactor *SubReactor) Offer(fd int) {
	select {
	case reactor.fd <- fd:
	default:
	}

}
