package linker

import "context"

type HandleFunc func(ctx Context)

type Engine struct {
	handleChains []HandleFunc
}

func (e *Engine) allocateContext(conn Conn) Context {
	return &selfContext{Context: context.Background(), engine: e, conn: conn}
}

func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

func (e *Engine) HandleRequest(conn Conn) {
	body, err := conn.read()
	if err != nil {
		conn.Close()
		return
	}

	ctx := e.allocateContext(conn)
	ctx.SetBody(body)
	ctx.Run()
}
