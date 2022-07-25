//go:build linux

package poller

import (
	"golang.org/x/sys/unix"
	"syscall"
)

// Epoll implements of Poller
type Epoll struct {
	// 注册的事件的文件描述符
	fd int
	// max event size, default: 100
	maxEventSize int
}

func (impl *Epoll) Add(fd int) error {
	// 向 epoll 实例注册文件描述符对应的事件
	// POLLIN 表示对应的文件描述字可以读
	// POLLHUP 表示对应的文件描述字被挂起
	// 只有当链接有数据可以读或者连接被关闭时，wait才会唤醒
	return unix.EpollCtl(impl.fd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
}

func (impl *Epoll) Remove(fd int) error {
	// 向 epoll 实例删除文件描述符对应的事件
	return unix.EpollCtl(impl.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (impl *Epoll) Wait() (read []int, closed []int, err error) {
	events := make([]unix.EpollEvent, impl.maxEventSize)
	n, err := unix.EpollWait(impl.fd, events, 100)
	if err != nil {
		return
	}

	for i := 0; i < n; i++ {
		if events[i].Fd == 0 {
			continue
		}
		
		if events[i].Events == unix.POLLIN {
			read = append(read, int(events[i].Fd))
		} else if events[i].Events == unix.POLLHUP {
			closed = append(read, int(events[i].Fd))
		}

	}

	return
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
