package linker

// Callback 回调
type Callback struct {
	onReceive    HandleFunc
	onConnect    func(conn IConnection)
	onDisconnect func(conn IConnection)
}

func (c *Callback) SetOnConnect(onConnect func(conn IConnection)) {
	c.onConnect = onConnect
}

func (c *Callback) SetOnDisconnect(onDisconnect func(conn IConnection)) {
	c.onDisconnect = onDisconnect
}

// SetOnReceive 接收请求回调
func (c *Callback) SetOnReceive(hookFunc HandleFunc) {
	c.onReceive = hookFunc
}

func (c *Callback) OnReceive(ctx IContext) {
	if c.onReceive == nil {
		return
	}
	c.onReceive(ctx)
}

func (c *Callback) OnConnect(conn IConnection) {
	if c.onConnect == nil {
		return
	}
	c.onConnect(conn)
}

func (c *Callback) OnDisconnect(conn IConnection) {
	if c.onDisconnect == nil {
		return
	}
	c.onDisconnect(conn)
}
