package reactor

import (
	"log"
	"math/rand"
	"net"
)

const (
	TCP = "tcp"
	WS  = "websocket"
	UDP = "udp"
)

type MainReactor struct {
	ev     *EventHandler
	engine *Engine

	acceptorLoop *AcceptorLoop
	children     []*SubReactor
}

func NewReactor() *MainReactor {
	return &MainReactor{
		ev:           new(EventHandler),
		engine:       new(Engine),
		acceptorLoop: new(AcceptorLoop),
		children:     make([]*SubReactor, 16),
	}
}

func (reactor *MainReactor) init() {
	for i := 0; i < 16; i++ {
		reactor.children[i] = &SubReactor{
			connected:    reactor.ev.HandleConnect,
			disconnected: reactor.ev.HandleDisconnect,
			poll:         nil,
			connections:  make(map[int]Conn, 1024),
		}
	}
	reactor.acceptorLoop.connDispatcher = func(conn net.Conn) {
		sub := reactor.children[rand.Intn(len(reactor.children)-1)]
		c := newConn(conn)
		if err := sub.Register(c); err != nil {
			c.Close()
		}
	}
}

func (reactor *MainReactor) Listen(protocol string, bind string) (err error) {
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

	for _, sub := range reactor.children {
		sub.Run(reactor.engine)
	}
}

type SubReactor struct {
	connected, disconnected ConnEvent
	poll                    Poller
	connections             map[int]Conn
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
	reactor.connected(conn)
	return nil
}

func (reactor *SubReactor) Release(conn Conn) error {
	reactor.disconnected(conn)
	fd := conn.FD()
	if err := reactor.poll.Remove(fd); err != nil {
		return err
	}
	delete(reactor.connections, fd)
	return nil
}

func (reactor *SubReactor) Trigger(fn func(conn Conn)) error {
	fds, err := reactor.poll.Wait()
	if err != nil {
		return err
	}
	for _, fd := range fds {
		conn := reactor.getConn(fd)
		if conn == nil {
			continue
		}
		fn(conn)
	}
	return nil
}

func (reactor *SubReactor) getConn(fd int) Conn {
	return reactor.connections[fd]
}

func (reactor *SubReactor) Run(e *Engine) {
	for {
		// 通过wait方法获取到epoll管理的活跃socket连接
		err := reactor.Trigger(func(conn Conn) {
			body, err := conn.read()
			if err != nil {

			}
			ctx := e.allocateContext(conn)
			ctx.SetBody(body)
			go ctx.Run()
		})
		if err != nil {
			log.Println("unable to get active socket connection from epoll:", err)
			continue
		}

	}
}
