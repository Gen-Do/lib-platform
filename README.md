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
