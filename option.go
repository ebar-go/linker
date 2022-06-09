package linker

type Option func(conf *Config)

func WithAcceptCount(count int) Option {
	return func(conf *Config) {
		conf.Accept = count
	}
}

func WithQueueSize(size int) Option {
	return func(conf *Config) {
		conf.QueueSize = size
	}
}

func WithWebsocketRequestURI(uri string) Option {
	return func(conf *Config) {

	}
}

func WithDebug() Option {
	return func(conf *Config) {
		conf.Debug = true
	}
}

func WithDataLength(length int) Option {
	return func(conf *Config) {
		conf.DataLength = length
	}
}
