package platform

import (
	"os"
	"strconv"
	"time"
)

const (
	envVarPort  = "PORT"
	defaultPort = 8080

	readinessProbeEndpoint = "/health"
	livenessProbeEndpoint  = "/livez"

	ExitCodeSuccess = 0
	ExitCodeFailure = 1
)

func envInt(env string, def int) int {
	if p, err := strconv.Atoi(os.Getenv(env)); err == nil {
		return p
	}
	return def
}

func envDuration(env string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(env)); err == nil {
		return d
	}
	return def
}
