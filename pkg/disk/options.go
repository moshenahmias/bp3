package disk

import "github.com/moshenahmias/bp3/pkg/bp3"

type options struct {
	order          int
	pages          []ReadWriteSeekSyncTruncater
	maxCachedPages int
}

// Option represents a functional option for configuring a B+ Tree instance
type Option func(*options)

func buildOptions(options_ ...Option) options {
	opts := options{
		order: bp3.MinOrder,
	}

	for _, opt := range options_ {
		opt(&opts)
	}

	if opts.maxCachedPages == 0 {
		opts.maxCachedPages = max(1, (len(opts.pages)+1)/2)
	}

	return opts
}

// WithOrder sets the order (degree) of the B+ Tree.
func WithOrder(order int) Option {
	return func(o *options) {
		o.order = order
	}
}

// WithIndexPage adds a storage page for the B+ Tree index.
func WithIndexPage(page ReadWriteSeekSyncTruncater) Option {
	return func(o *options) {
		o.pages = append(o.pages, page)
	}
}

// WithIndexPages adds multiple storage pages for the B+ Tree index.
func WithIndexPages(pages []ReadWriteSeekSyncTruncater) Option {
	return func(o *options) {
		o.pages = append(o.pages, pages...)
	}
}

// WithMaxCachedPages sets the maximum number of cached pages for the B+ Tree.
func WithMaxCachedPages(max int) Option {
	return func(o *options) {
		o.maxCachedPages = max
	}
}
