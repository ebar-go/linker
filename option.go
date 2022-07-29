package linker

import "runtime"

type options struct {
	processor int
}

func defaultOption() *options {
	return &options{
		processor: runtime.NumCPU(),
	}
}

type Option func(opts *options)

func WithProcessor(n int) Option {
	return func(opts *options) {
		opts.processor = n
	}
}
