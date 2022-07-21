package reactor

import (
	"bufio"
	"linker/reactor/buffer"
	"net"
	"reflect"
	"sync"
)

type Conn interface {
	FD() int
	Close()

	read() ([]byte, error)
}

type Connection struct {
	mu            sync.Mutex
	loop          *SubReactor
	instance      net.Conn
	inboundBuffer *buffer.Buffer
	fd            int
	scanner       *bufio.Scanner
}

func (conn *Connection) FD() int {
	return conn.fd
}

func newConn(conn net.Conn) *Connection {
	return &Connection{instance: conn, fd: SocketFD(conn), scanner: bufio.NewScanner(conn)}
}

func (conn *Connection) read() ([]byte, error) {
	if !conn.scanner.Scan() {
		return nil, conn.scanner.Err()
	}
	return conn.scanner.Bytes(), nil
}

func (conn *Connection) Close() {
	_ = conn.loop.Release(conn)
	_ = conn.instance.Close()
}

// SocketFD get socket connection fd
func SocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
