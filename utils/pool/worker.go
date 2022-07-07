package pool

type WorkerPool struct {
	size int
	task chan struct{}
}

func NewWorkerPool(size int) *WorkerPool {
	pool := &WorkerPool{size: size, task: make(chan struct{}, size)}
	for i := 0; i < size; i++ {
		pool.task <- struct{}{}
	}
	return pool
}

// Take get a free worker from pool, it will block when pool is empty
func (pool WorkerPool) Take() {
	<-pool.task
}

// Release give back a worker to pool
func (pool WorkerPool) Release() {
	pool.task <- struct{}{}
}
