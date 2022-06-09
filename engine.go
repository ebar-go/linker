package linker

import (
	"math"
)

const maxIndex = math.MaxInt8 / 2

type Engine struct {
	handleChains []HandleFunc
}

func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

func (e *Engine) allocateContext() *Context {
	return &Context{engine: e}
}
