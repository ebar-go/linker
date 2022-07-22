package linker

import (
	"bufio"
	uuid "github.com/satori/go.uuid"
	"linker/pkg/buffer"
	"net"
	"reflect"
	"sync"
)

type Conn interface {
	ID() string
	FD() int
	Close()
	Push(msg []byte)

	read() ([]byte, error)
}
type Connection struct {
	mu             sync.Mutex
	instance       net.Conn
	fd             int
	scanner        *bufio.Scanner
	uuid           string // 唯一ID
	once           *sync.Once
	linkedBuffer   *buffer.Buffer
	closedCallback ConnEvent
}

func (conn *Connection) FD() int {
	return conn.fd
}

func (conn *Connection) Push(msg []byte) {
	conn.instance.Write(msg)
}

func newConn(conn net.Conn) *Connection {
	return &Connection{instance: conn,
		uuid:    uuid.NewV4().String(),
		fd:      SocketFD(conn),
		once:    new(sync.Once),
		scanner: bufio.NewScanner(conn),
	}
}

// UUID 返回连接的唯一ID
func (conn *Connection) ID() string {
	return conn.uuid
}

func (conn *Connection) read() ([]byte, error) {
	if !conn.scanner.Scan() {
		return nil, conn.scanner.Err()
	}
	return conn.scanner.Bytes(), nil
}

func (conn *Connection) Close() {
	conn.once.Do(func() {
		if conn.closedCallback != nil {
			conn.closedCallback(conn)
		}
		_ = conn.instance.Close()
	})

}

// SocketFD get socket connection fd
func SocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
