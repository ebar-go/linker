package poller

type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() (read []int, closed []int, err error)
}
