package reactor

type ConnEvent func(conn Conn)

type EventHandler struct {
}

func (handler EventHandler) HandleRequest(ctx Context) {

}

func (handler EventHandler) HandleConnect(conn Conn) {

}

func (handler EventHandler) HandleDisconnect(conn Conn) {

}
