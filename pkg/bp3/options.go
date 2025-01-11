package bp3

type options struct {
	order int
}

type Option func(*options)

func buildOptions(options_ ...Option) options {
	opts := options{
		order: MinOrder,
	}

	for _, opt := range options_ {
		opt(&opts)
	}

	return opts
}

// WithOrder sets the order (degree) of the B+ Tree.
func WithOrder(order int) Option {
	return func(o *options) {
		o.order = order
	}
}
