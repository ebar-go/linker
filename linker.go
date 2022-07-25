package linker

import "linker/config"

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
		Engine:       new(Engine),
		children:     make([]*SubReactor, 16),
	}
	reactor.init()
	return reactor
}

func NewTCPServer() *TcpServer {
	conf := config.Default()

	return &TcpServer{
		EventHandler: new(EventHandler),
		engine:       new(Engine),
		conf:         conf,
	}
}
