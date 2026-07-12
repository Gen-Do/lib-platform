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

// WithoutReadinessProbe disables the library-managed readiness probe on
// readinessProbeEndpoint ("/health"). When passed, listen() does not
// register any handler for that path, leaving the service free to own
// "/health" itself (e.g. to implement a dependency-aware readiness check
// distinct from the always-static liveness probe on "/livez").
//
// Without this option, behavior is unchanged from v1.0.1: the readiness
// probe is registered on "/health" as before.
func WithoutReadinessProbe() Option {
	return option(func(o *options) { o.skipReadinessProbe = true })
}
