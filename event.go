package linker

type ConnEvent func(conn Conn)

type EventHandler struct {
	connect    ConnEvent
	disconnect ConnEvent
	request    HandleFunc
}

func (handler *EventHandler) OnConnect(connect ConnEvent) {
	handler.connect = connect
}

func (handler *EventHandler) OnDisconnect(disconnect ConnEvent) {
	handler.disconnect = disconnect
}

func (handler *EventHandler) OnRequest(request HandleFunc) {
	handler.request = request
}

func (handler EventHandler) HandleRequest(ctx *Context) {
	if handler.request == nil {
		return
	}
	handler.request(ctx)
}

func (handler EventHandler) HandleConnect(conn Conn) {
	if handler.connect == nil {
		return
	}

	handler.connect(conn)
}

func (handler EventHandler) HandleDisconnect(conn Conn) {
	if handler.disconnect == nil {
		return
	}
	handler.disconnect(conn)
}
