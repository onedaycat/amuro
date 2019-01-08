package cognitoevent

type Option func(o *option)

type option struct {
	preHandlers  []PreHandler
	postHandlers []PostHandler
}

func newOption(opts ...Option) *option {
	o := &option{}
	if opts == nil {
		return o
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithPreHandlers(preHandlers ...PreHandler) Option {
	return func(o *option) {
		o.preHandlers = preHandlers
	}
}

func WithPostHandlers(postHandlers ...PostHandler) Option {
	return func(o *option) {
		o.postHandlers = postHandlers
	}
}
