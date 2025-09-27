package platform

type Option interface {
	listener(*options)
	instance(*options)
}

func WithMux(m Mux) Option {
	return option(func(o *options) { o.mux = m })
}

func WithPort(p int) Option {
	return option(func(o *options) { o.port = p })
}

func WithChecker(name string, c Checker) Option {
	return option(func(o *options) {
		if o.checkers == nil {
			o.checkers = make(map[string]Checker)
		}
		o.checkers[name] = c
	})
}

func WithLogger(logger Logger) Option {
	return option(func(o *options) { o.logger = logger })
}

var WithReporter = WithLogger
