package reactor

import (
	"log"
	"net"
)

type AcceptorLoop struct {
	lis            net.Listener
	Sndbuf         int
	Rcvbuf         int
	KeepAlive      bool
	connDispatcher func(conn net.Conn)
}

func (loop AcceptorLoop) Run(lis net.Listener) {
	if listener, ok := lis.(*net.TCPListener); ok {
		loop.runTCP(listener)
	}
}

func (loop AcceptorLoop) runTCP(lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
	)

	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Printf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(loop.KeepAlive); err != nil {
			log.Printf("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(loop.Rcvbuf); err != nil {
			log.Printf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(loop.Sndbuf); err != nil {
			log.Printf("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		loop.connDispatcher(conn)
	}
}
