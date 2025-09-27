package platform

type InstanceOption interface {
	instance(*options)
}

type ObservabilitySettings struct {
	Logger  Logger
	Metrics Metrics
}

func WithObservability(settings ObservabilitySettings) InstanceOption {
	return option(func(o *options) {
		if settings.Logger != nil {
			WithLogger(settings.Logger).instance(o)
		}
		if settings.Metrics != nil {
			WithMetrics(settings.Metrics).instance(o)
		}
	})
}

func WithEnableSignalHandling(enable bool) InstanceOption {
	return option(func(o *options) { o.enableSignalsHandling = enable })
}
func WithWorkers(w ...Worker) InstanceOption {
	return option(func(o *options) { o.workers = append(o.workers, w...) })
}
func WithMetrics(metrics Metrics) InstanceOption {
	return option(func(o *options) { o.metrics = metrics })
}
func WithListener(listener Listener) InstanceOption {
	return option(func(o *options) { o.listener = listener })
}
func WithIgnoreErrors(errs ...error) InstanceOption {
	return option(func(o *options) {
		o.ignoreErrors = append(o.ignoreErrors, errs...)
	})
}
func WithoutIgnoreErrors() InstanceOption {
	return option(func(o *options) {
		o.ignoreErrors = nil
	})
}
func WithIgnoreErrorsFuncs(fs ...func(error) bool) InstanceOption {
	return option(func(o *options) {
		o.ignoreErrorsFuncs = append(o.ignoreErrorsFuncs, fs...)
	})
}
func WithoutIgnoreErrorsFuncs() InstanceOption {
	return option(func(o *options) {
		o.ignoreErrorsFuncs = nil
	})
}
