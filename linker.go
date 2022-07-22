package linker

const (
	TCP = "tcp"
	WS  = "websocket"
	UDP = "udp"
)

type EventLoop interface {
	Listen(protocol string, bind string) (err error)

	OnConnect(connect ConnEvent)
	OnDisconnect(disconnect ConnEvent)
	OnRequest(request HandleFunc)
	Use(handlers ...HandleFunc)
}
