package reactor

import (
	"linker/reactor/epoll"
	"log"
	"net"
)

const (
	TCP = "tcp"
	WS  = "websocket"
	UDP = "udp"
)

type MainReactor struct {
	poll   Poller
	ev     *EventHandler
	engine *Engine

	acceptorLoop *AcceptorLoop
	children     []*SubReactor
}

func NewReactor() *MainReactor {
	reactor := &MainReactor{
		ev:           new(EventHandler),
		engine:       new(Engine),
		acceptorLoop: new(AcceptorLoop),
		children:     make([]*SubReactor, 16),
	}
	reactor.init()
	return reactor
}

func (reactor *MainReactor) init() {
	reactor.poll, _ = epoll.Create()
	for i := 0; i < 16; i++ {
		reactor.children[i] = &SubReactor{
			ev:          reactor.ev,
			poll:        reactor.poll,
			connections: make(map[int]Conn, 1024),
			fd:          make(chan int, 64),
		}
	}
	reactor.acceptorLoop.Sndbuf = 4096
	reactor.acceptorLoop.Rcvbuf = 4096
	reactor.acceptorLoop.connDispatcher = reactor.dispatcher
}

func (reactor *MainReactor) dispatcher(conn net.Conn) {
	c := newConn(conn)
	sub := reactor.children[c.FD()%len(reactor.children)]
	if err := sub.Register(c); err != nil {
		c.Close()
		return
	}
	c.loop = sub
}

func (reactor *MainReactor) Listen(protocol string, bind string) (err error) {
	log.Printf("%s server listen: %s\n", protocol, bind)
	switch protocol {
	case TCP:
		return reactor.listenTCP(bind)
	}
	return nil
}

func (reactor *MainReactor) listenTCP(bind string) (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)

	if addr, err = net.ResolveTCPAddr(TCP, bind); err != nil {
		log.Printf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
		return
	}
	if listener, err = net.ListenTCP(TCP, addr); err != nil {
		log.Printf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
		return
	}

	reactor.acceptorLoop.Run(listener)

	return
}

func (reactor *MainReactor) Start() {
	reactor.engine.Use(reactor.ev.HandleRequest)

	reactor.Run()
}

type SubReactor struct {
	ev   *EventHandler
	poll Poller

	connections map[int]Conn
	fd          chan int
}
type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]int, error)
}

func (reactor *SubReactor) Register(conn Conn) error {
	fd := conn.FD()
	if err := reactor.poll.Add(fd); err != nil {
		return err
	}

	reactor.connections[fd] = conn
	reactor.ev.HandleConnect(conn)
	return nil
}

func (reactor *SubReactor) Release(conn Conn) error {
	reactor.ev.HandleDisconnect(conn)
	fd := conn.FD()
	if err := reactor.poll.Remove(fd); err != nil {
		return err
	}
	delete(reactor.connections, fd)
	return nil
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

func (reactor *MainReactor) Run() {
	for _, sub := range reactor.children {
		go sub.Polling(func(conn Conn) {
			body, err := conn.read()
			if err != nil {
				return
			}

			if len(body) == 0 {
				conn.Close()
				return
			}

			ctx := reactor.engine.allocateContext(conn)
			ctx.SetBody(body)
			go ctx.Run()
		})
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
