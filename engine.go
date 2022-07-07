package linker

import (
	"math"
	"sync"
)

const maxIndex = math.MaxInt8 / 2

type Engine struct {
	mask         int
	pools        []*sync.Pool
	handleChains []HandleFunc
}

func newEngine(n int) *Engine {
	e := &Engine{mask: n - 1}
	for i := 0; i < n; i++ {
		e.pools = append(e.pools, &sync.Pool{New: func() interface{} {
			return e.allocateContext()
		}})
	}
	return e
}

func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

func (e *Engine) allocateContext() *Context {
	return &Context{engine: e}
}

func (e *Engine) ContextPool(r int) *sync.Pool {
	return e.pools[r&e.mask]
}
