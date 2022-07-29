package linker

import (
	"context"
	"errors"
	"sync"
)

type HandleFunc func(ctx *Context)

type Engine struct {
	handleChains []HandleFunc
	contextPools []*sync.Pool
	mask         int
}

func newEngine(ctxPoolSize int) *Engine {
	engine := &Engine{
		mask: ctxPoolSize,
	}
	engine.init()
	return engine
}

func (e *Engine) allocateContext(conn Conn, body []byte) *Context {
	return &Context{engine: e, conn: conn, body: body}
}

func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

func (e *Engine) init() {
	for i := 0; i < e.mask; i++ {
		e.contextPools = append(e.contextPools, &sync.Pool{New: func() interface{} {
			return &Context{engine: e}
		}})
	}
}

func (e *Engine) buildContext(conn Conn) (*Context, error) {
	body, err := conn.read()
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, errors.New("empty data")
	}

	ctx := e.withPool(conn.FD()).Get().(*Context)
	ctx.body = body
	ctx.index = 0
	ctx.conn = conn
	ctx.Context = context.Background()
	return ctx, nil
}

func (e *Engine) releaseContext(ctx *Context) {
	e.withPool(ctx.Conn().FD()).Put(ctx)
}

func (e *Engine) withPool(key int) *sync.Pool {
	return e.contextPools[key&(e.mask-1)]
}
