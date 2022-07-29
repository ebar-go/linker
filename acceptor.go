package linker

import (
	"log"
	"net"
)

type Acceptor interface {
	Listen(addr string) error
}

type acceptor struct {
	send           int
	receive        int
	keepalive      bool
	connDispatcher func(conn net.Conn)
}

func newAcceptor(dispatcher func(conn net.Conn)) *acceptor {
	return &acceptor{
		send:           4096,
		receive:        4096,
		keepalive:      false,
		connDispatcher: dispatcher,
	}
}

func (loop *acceptor) WithReadBuffer(bytes int) *acceptor {
	loop.receive = bytes
	return loop
}
func (loop *acceptor) WithWriteBuffer(bytes int) *acceptor {
	loop.send = bytes
	return loop
}

type TCPAcceptor struct {
	*acceptor
}

func (loop TCPAcceptor) Listen(bind string) (err error) {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		log.Printf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
		return
	}
	return loop.accept(addr)
}

func (loop TCPAcceptor) accept(addr *net.TCPAddr) (err error) {
	var (
		lis  *net.TCPListener
		conn *net.TCPConn
	)

	if lis, err = net.ListenTCP("tcp", addr); err != nil {
		return
	}

	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Printf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(loop.keepalive); err != nil {
			log.Printf("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(loop.receive); err != nil {
			log.Printf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(loop.send); err != nil {
			log.Printf("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		loop.connDispatcher(conn)
	}

}

type UDPAcceptor struct {
	*acceptor
}

func (loop UDPAcceptor) Listen(bind string) (err error) {
	addr, err := net.ResolveUDPAddr("udp", bind)
	if err != nil {
		log.Printf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
		return
	}
	return loop.accept(addr)
}
func (loop UDPAcceptor) accept(addr *net.UDPAddr) (err error) {
	var (
		conn *net.UDPConn
	)

	for {
		if conn, err = net.ListenUDP("udp", addr); err != nil {
			log.Printf("net.ListenUDP(udp, %s) error(%v)", addr.IP, err)
			return
		}

		if err = conn.SetReadBuffer(loop.receive); err != nil {
			log.Printf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(loop.send); err != nil {
			log.Printf("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		loop.connDispatcher(conn)

	}
}

type WebsocketAcceptor struct {
}

func NewTCPAcceptor(dispatcher func(conn net.Conn)) *TCPAcceptor {
	return &TCPAcceptor{acceptor: newAcceptor(dispatcher)}
}
func NewUDPAcceptor(dispatcher func(conn net.Conn)) *UDPAcceptor {
	return &UDPAcceptor{acceptor: newAcceptor(dispatcher)}
}
