package linker

// event 事件
type event struct {
	onRequest    HandleFunc
	onConnect    func(conn IConnection)
	onDisconnect func(conn IConnection)
}

func (c *event) SetOnConnect(onConnect func(conn IConnection)) {
	c.onConnect = onConnect
}

func (c *event) SetOnDisconnect(onDisconnect func(conn IConnection)) {
	c.onDisconnect = onDisconnect
}

// SetOnRequest 接收请求回调
func (c *event) SetOnRequest(hookFunc HandleFunc) {
	c.onRequest = hookFunc
}

func (c *event) OnRequest(ctx IContext) {
	if c.onRequest == nil {
		return
	}
	c.onRequest(ctx)
}

func (c *event) OnConnect(conn IConnection) {
	if c.onConnect == nil {
		return
	}
	c.onConnect(conn)
}

func (c *event) OnDisconnect(conn IConnection) {
	if c.onDisconnect == nil {
		return
	}
	c.onDisconnect(conn)
}
