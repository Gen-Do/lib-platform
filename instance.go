package platform

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
)

type Instance struct {
	options

	workersMu sync.Mutex
}

func New(opts ...InstanceOption) (*Instance, error) {
	o := options{
		port:                  envInt(envVarPort, defaultPort),
		mux:                   http.DefaultServeMux,
		enableSignalsHandling: true,
		ignoreErrors:          []error{context.Canceled},
	}

	for _, opt := range opts {
		opt.instance(&o)
	}

	return &Instance{
		options: o,
	}, nil
}

func Run(ctx context.Context, opts ...InstanceOption) error {
	instance, err := New(opts...)
	if err != nil {
		return fmt.Errorf("can not create instance: %w", err)
	}

	return instance.Run(ctx)
}

func (i *Instance) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	i.workersMu.Lock()
	workers := append([]Worker{}, i.workers...)
	i.workersMu.Unlock()

	if i.enableSignalsHandling {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		defer stop()
	}

	if i.listener != nil {
		workers = append(workers, WorkerFunc(func(ctx context.Context) error {
			return listen(ctx, i.listener, i.options)
		}))
	}

	err := runWorkers(ctx, workers)
	if isIgnoreError(err, combineIgnoreErrors(i.ignoreErrors, i.ignoreErrorsFuncs)) {
		if i.logger != nil {
			i.logger.Info(i.logger.WithError(ctx, err), "instance finished with error, but this error is ignored")
		}
		return nil
	}
	return err
}

func (i *Instance) AddWorkers(workers ...Worker) {
	i.workersMu.Lock()
	defer i.workersMu.Unlock()

	i.workers = append(i.workers, workers...)
}
