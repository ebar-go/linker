package old

func NewGroupServer() *GroupServer {
	return &GroupServer{items: []IServer{}}
}

func NewTCPServer(bind []string, opts ...Option) IServer {
	conf := defaultConfig()
	conf.Bind = bind
	for _, setter := range opts {
		setter(conf)
	}

	return &TcpServer{
		components: newComponents(),
		conf:       conf,
	}
}

func NewWebsocketServer(bind []string, requestURI string, opts ...Option) IServer {
	conf := defaultConfig()
	conf.Bind = bind
	for _, setter := range opts {
		setter(conf)
	}

	return &WebsocketServer{
		components: newComponents(),
		conf:       conf,
		requestURI: requestURI,
	}
}

func NewUDPServer() IServer {
	return nil
}

type components struct {
	*event
	*engine
}

func newComponents() components {
	return components{
		event:  new(event),
		engine: newengine(32),
	}
}
