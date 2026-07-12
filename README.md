# lib-platform

```go
package main

import (
	"context"
	platform "github.com/Gen-Do/lib-platform"
	"os"
)

func main() {
	os.Exit(run())
}

func run() int {
	// ...
	err := platform.Run(
		context.Background(),
		platform.WithObservability(platform.ObservabilitySettings{
			Logger:  log,
			Metrics: metric,
		}),
		platform.WithMux(mux),
		platform.WithListener(lsn),
	)
	if err != nil {
		log.Error(log.WithError(ctx, err), "application finished with an error")
		return platform.ExitCodeFailure
	}
	
	return platform.ExitCodeSuccess
}

```

## Health endpoints

- **`/health` (readiness)** — dependency-aware. By default the library
  registers a readiness probe here that runs all `Checker`s added via
  `WithChecker` and returns `200` only if every checker succeeds (`500`
  otherwise). Suitable for k8s `readinessProbe`: a failing dependency
  should take the pod out of the load-balancer rotation, but should not
  kill it.
- **`/livez` (liveness)** — always registered, always returns a static
  `200 OK` with body `{"status":"alive"}`. It never runs `Checker`s and
  never depends on downstream state — it only answers "is the process
  alive and serving HTTP". Suitable for k8s `livenessProbe`: a transient
  dependency blip must not cause a restart loop.

### Opting out of the built-in `/health` handler

If a service wants to own `/health` itself (e.g. to implement its own
readiness semantics, or to avoid the library's handler shadowing a
service-defined one - chi/mux registration is last-wins), pass
`WithoutReadinessProbe()`:

```go
err := platform.Run(
	context.Background(),
	platform.WithMux(mux),
	platform.WithListener(lsn),
	platform.WithoutReadinessProbe(), // library will not touch "/health"
)
```

With this option, `listen()` does not call `mux.Handle("/health", ...)` at
all - the service's own handler (if any) is the only one serving that
path. `/livez` is unaffected and is always registered regardless of this
option.

**Backward compatibility**: without `WithoutReadinessProbe()`, behavior is
unchanged from v1.0.1 - the readiness probe is registered on `/health`
exactly as before. `/livez` is new in v1.1.0 and additive only.
