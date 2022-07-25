package linker

import (
	"context"
	"errors"
)

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

func (e *Engine) buildContext(conn Conn) (Context, error) {
	// TODO 使用linkedBuffer优化bytes拷贝
	body, err := conn.read()
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, errors.New("empty data")
	}

	ctx := e.allocateContext(conn)
	ctx.SetBody(body)
	return ctx, nil
}
