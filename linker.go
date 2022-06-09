package linker

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
		Callback: Callback{},
		engine:   new(Engine),
		conf:     conf,
	}
}

func NewWebsocketServer() IServer {
	return nil
}

func NewUDPServer() IServer {
	return nil
}