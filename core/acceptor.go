package core

import (
	"log"
	"net"
)

type Acceptor struct {
	send           int
	receive        int
	keepalive      bool
	connDispatcher func(conn net.Conn)
}

func NewAcceptor(dispatcher func(conn net.Conn)) *Acceptor {
	return &Acceptor{
		send:           4096,
		receive:        4096,
		keepalive:      false,
		connDispatcher: dispatcher,
	}
}

func (loop *Acceptor) WithReadBuffer(bytes int) *Acceptor {
	loop.receive = bytes
	return loop
}
func (loop *Acceptor) WithWriteBuffer(bytes int) *Acceptor {
	loop.send = bytes
	return loop
}

func (loop Acceptor) AcceptTCP(addr *net.TCPAddr) {
	var (
		lis  *net.TCPListener
		conn *net.TCPConn
		err  error
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

func (loop Acceptor) AcceptUDP(addr *net.UDPAddr) {
	var (
		conn *net.UDPConn
		err  error
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
