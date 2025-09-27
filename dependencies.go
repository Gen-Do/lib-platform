package platform

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

type Listener interface {
	Listen(context.Context, int, http.Handler) error
}

type Mux interface {
	http.Handler
	Handle(string, http.Handler)
}

type Metrics interface {
	prometheus.Gauge
	prometheus.Summary
}

type Checker interface {
	Check() (interface{}, error)
}

type Fields map[string]any
type Logger interface {
	WithField(ctx context.Context, key string, value any) context.Context
	WithError(ctx context.Context, err error) context.Context

	Info(ctx context.Context, args ...any)
	Warn(ctx context.Context, args ...any)
}
