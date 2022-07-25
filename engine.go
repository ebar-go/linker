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

func (e *Engine) buildContext(conn Conn) Context {
	body, err := conn.read()
	if err != nil {
		return nil
	}

	if len(body) == 0 {
		return nil
	}

	ctx := e.allocateContext(conn)
	ctx.SetBody(body)
	return ctx
}
