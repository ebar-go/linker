package linker

const (
	TCP = "tcp"
	WS  = "websocket"
	UDP = "udp"
)

type EventLoop interface {
	OnConnect(connect ConnEvent)
	OnDisconnect(disconnect ConnEvent)
	OnRequest(request HandleFunc)
	Use(handlers ...HandleFunc)
	Run(protocol string, bind string) (err error)
}

func NewReactor(opts ...Option) EventLoop {
	option := defaultOption()
	for _, setter := range opts {
		setter(option)
	}
	reactor := &MainReactor{
		EventHandler: new(EventHandler),
		Engine:       newEngine(),
		children:     make([]*SubReactor, option.processor),
	}
	reactor.init()
	return reactor
}
