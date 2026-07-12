package platform

type options struct {
	port               int
	mux                Mux
	checkers           map[string]Checker
	logger             Logger
	skipReadinessProbe bool

	listener Listener

	enableSignalsHandling bool
	ignoreErrors          []error
	ignoreErrorsFuncs     []func(error) bool

	workers []Worker
	metrics Metrics
}

type option func(o *options)

func (o option) listener(os *options) { o(os) }
func (o option) instance(os *options) { o(os) }
