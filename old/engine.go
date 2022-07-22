package old

import (
	"math"
	"sync"
)

const maxIndex = math.MaxInt8 / 2

type engine struct {
	mask         int
	pools        []*sync.Pool
	handleChains []HandleFunc
}

func newengine(n int) *engine {
	e := &engine{mask: n - 1}
	for i := 0; i < n; i++ {
		e.pools = append(e.pools, &sync.Pool{New: func() interface{} {
			return e.allocateContext()
		}})
	}
	return e
}

func (e *engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

func (e *engine) allocateContext() *Context {
	return &Context{engine: e}
}

func (e *engine) contextPool(r int) *sync.Pool {
	return e.pools[r&e.mask]
}
