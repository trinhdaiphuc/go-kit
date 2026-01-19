package header

// Option configures how we set up the request parser.
type Option func(p *parser)

func WithParser(metadataFns ...ParseMetadataFn) Option {
	return func(p *parser) {
		p.metadataFns = metadataFns
	}
}

func WithExtraKeys(extraKeys []string) Option {
	return func(p *parser) {
		p.extraKeys = extraKeys
	}
}
