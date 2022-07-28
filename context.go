package linker

import (
	"context"
	"log"
	"math"
)

type Context struct {
	context.Context
	engine *Engine
	conn   Conn
	body   []byte
	index  int8
}

func (ctx *Context) Conn() Conn {
	return ctx.conn
}

func (ctx *Context) Body() []byte {
	return ctx.body
}
func (ctx *Context) Run() {
	ctx.engine.handleChains[0](ctx)
	ctx.engine.releaseContext(ctx)
}

func (ctx *Context) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.engine.handleChains[ctx.index](ctx)
	}
}
func (ctx *Context) Abort() {
	ctx.index = maxIndex
	log.Println("已被终止...")
}

const (
	maxIndex = math.MaxInt8 / 2
)
