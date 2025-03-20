package redislock

type Options struct {
	prefix string
	suffix string
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.prefix = prefix
	}
}

func WithSuffix(suffix string) Option {
	return func(o *Options) {
		o.suffix = suffix
	}
}

func newDefaultOption() *Options {
	return &Options{
		prefix: "",
		suffix: "_lock",
	}
}
