//go:build linux

package poller

import (
	"fmt"
	"golang.org/x/sys/unix"
	"syscall"
)

// Epoll implements of Poller
type Epoll struct {
	// 句柄
	fd int
	// max event size, default: 100
	maxEventSize int
}

func (impl *Epoll) Add(fd int) error {
	return unix.EpollCtl(impl.fd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
}

func (impl *Epoll) Remove(fd int) error {
	return unix.EpollCtl(impl.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (impl *Epoll) Wait() ([]int, error) {
	events := make([]unix.EpollEvent, impl.maxEventSize)
	n, err := unix.EpollWait(impl.fd, events, 100)
	if err != nil {
		return nil, err
	}

	fds := make([]int, n)
	for i := 0; i < n; i++ {
		if events[i].Fd == 0 {
			continue
		}
		fmt.Println(events[i].Events)
		fds[i] = int(events[i].Fd)
	}
	return fds, nil
}

func CreateEpoll() (*Epoll, error) {
	fd, err := unix.EpollCreate(1)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:           fd,
		maxEventSize: 100,
	}, nil
}
