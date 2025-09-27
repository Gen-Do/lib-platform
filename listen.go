package platform

import (
	"context"
	"fmt"
	"net/http"
)

func Listen(ctx context.Context, listener Listener, opts ...Option) error {
	o := options{
		port: envInt(envVarPort, defaultPort),
		mux:  http.DefaultServeMux,
	}

	for _, opt := range opts {
		opt.listener(&o)
	}

	return listen(ctx, listener, o)
}

func listen(ctx context.Context, listener Listener, o options) error {
	probeHandler := readinessProbe(o.checkers, o.logger)
	o.mux.Handle(readinessProbeEndpoint, probeHandler)

	if o.logger != nil {
		o.logger.Info(ctx, fmt.Sprintf("start server on :%d", o.port))
	}

	return listener.Listen(ctx, o.port, o.mux)
}
