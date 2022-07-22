package reactor

import "context"

type Context interface {
	context.Context
	Run()
	SetBody(body []byte)
	Body() []byte
	Conn() Conn
}

type ContextPool interface {
	Get() Context
	Put(context Context)
}

type selfContext struct {
	context.Context
	engine *Engine
	conn   Conn
	body   []byte
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
