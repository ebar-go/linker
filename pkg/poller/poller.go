package poller

import (
	"net"
	"reflect"
)

type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() (read []int, closed []int, err error)
}

// SocketFD get socket connection fd
func SocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
