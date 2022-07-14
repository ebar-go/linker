package linker

import "runtime"

type Config struct {
	Debug      bool
	Bind       []string // 服务地址
	Accept     int      // 线程数
	QueueSize  int      // 队列长度
	DataLength int      // 协议的数据长度

	Sndbuf    int
	Rcvbuf    int
	KeepAlive bool
}

func defaultConfig() *Config {
	return &Config{
		Accept:     runtime.NumCPU(),
		QueueSize:  10,
		DataLength: 0,
		Sndbuf:     1024,
		Rcvbuf:     1024,
	}
}

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
