package linker

const (
	TCP = "tcp"
	WS  = "websocket"
	UDP = "udp"
)

type EventLoop interface {
	Run(protocol string, bind string) (err error)

	OnConnect(connect ConnEvent)
	OnDisconnect(disconnect ConnEvent)
	OnRequest(request HandleFunc)
	Use(handlers ...HandleFunc)
}

func NewReactor() EventLoop {
	reactor := &MainReactor{
		EventHandler: new(EventHandler),
		Engine:       newEngine(),
		children:     make([]*SubReactor, 32),
	}
	reactor.init()
	return reactor
}
