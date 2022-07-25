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

// Submit run some task
func (pool WorkerPool) Submit(fn func()) {
	pool.Take()
	go func() {
		defer pool.Release()
		fn()
	}()
}

type GoroutinePool struct {
	work chan func()
}

func NewGoroutinePool(size int) *GoroutinePool {
	gp := &GoroutinePool{
		work: make(chan func(), size),
	}

	for i := 0; i < size; i++ {
		go gp.run()
	}
	return gp
}
func (p *GoroutinePool) Schedule(task func()) {
	select {
	case p.work <- task:
	default:
	}
}

func (p *GoroutinePool) run() {
	var task func()
	for {
		task = <-p.work
		task()
	}
}
