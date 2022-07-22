package linker

import (
	"context"
	"fmt"
	"math"
)

type Context interface {
	context.Context
	Run()
	SetBody(body []byte)
	Body() []byte
	Conn() Conn
}

type selfContext struct {
	context.Context
	engine *Engine
	conn   Conn
	body   []byte
	index  int8
}

func (ctx *selfContext) Conn() Conn {
	return ctx.conn
}

func (ctx *selfContext) SetBody(body []byte) {
	ctx.body = body
}

func (ctx *selfContext) Body() []byte {
	return ctx.body
}
func (ctx *selfContext) Run() {
	ctx.engine.handleChains[0](ctx)
}

func (ctx *selfContext) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.engine.handleChains[ctx.index](ctx)
	}
}
func (ctx *selfContext) Abort() {
	ctx.index = maxIndex
	fmt.Println("已被终止...")
}

const (
	maxIndex = math.MaxInt8 / 2
)
