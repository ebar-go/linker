package linker

import (
	"fmt"
)

// Request 请求
type Request struct {
	body []byte // 请求内容
}

// Body 获取原始报文
func (request Request) Body() []byte {
	return request.body
}

// Context 上下文
type Context struct {
	Param

	index int8

	engine *Engine

	connection IConnection
	request    Request
}

// Channel 获取会话
func (c *Context) Channel() IConnection {
	return c.connection
}

// Request 获取请求
func (c *Context) Request() IRequest {
	return c.request
}

// Output 输出数据到客户端
func (c *Context) Output(msg []byte) {
	c.connection.Push(msg)
}

func (c *Context) Reset(body []byte, connection IConnection) {
	c.index = 0
	c.ResetKeys()
	c.request.body = body
	c.connection = connection
}

func (ctx *Context) Run() {
	ctx.engine.handleChains[0](ctx)
}

func (ctx *Context) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.engine.handleChains[ctx.index](ctx)
	}
}
func (ctx *Context) Abort() {
	ctx.index = maxIndex
	fmt.Println("已被终止...")
}
