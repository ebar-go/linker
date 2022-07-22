package poller

type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]int, error)
}
