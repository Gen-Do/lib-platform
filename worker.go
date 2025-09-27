package platform

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Worker interface {
	Run(ctx context.Context) error
}

type WorkerFunc func(ctx context.Context) error

func (w WorkerFunc) Run(ctx context.Context) error { return w(ctx) }

func waitUntilContext(w Worker) Worker {
	return WorkerFunc(func(ctx context.Context) error {
		if err := w.Run(ctx); err != nil {
			return err
		}

		<-ctx.Done()
		return ctx.Err()
	})
}

func recoverPanic(w Worker) Worker {
	return WorkerFunc(func(ctx context.Context) (err error) {
		defer func() {
			if pErr := recover(); pErr != nil {
				switch p := pErr.(type) {
				case error:
					err = p
				default:
					err = fmt.Errorf("%v", pErr)
				}
			}
		}()

		return w.Run(ctx)
	})
}

func runWorkers(ctx context.Context, workers []Worker) error {
	wCtx, wCancel := context.WithCancel(ctx)
	defer wCancel()

	errCh := make(chan error, len(workers))
	wg := sync.WaitGroup{}
	wg.Add(len(workers))
	for _, w := range workers {
		go func(w Worker) {
			defer wg.Done()
			defer wCancel()

			errCh <- recoverPanic(w).Run(wCtx)
		}(w)
	}

	wg.Wait()
	close(errCh)

	var joinedErr error
	for err := range errCh {
		joinedErr = errors.Join(joinedErr, err)
	}

	return joinedErr
}
