package platform

import "net/http"

// livenessProbe returns a static handler that always responds 200 OK,
// regardless of any registered Checker. Liveness answers only "is the
// process alive and serving HTTP" — it must never depend on downstream
// resources (DB, cache, etc.), otherwise a transient dependency blip would
// cause k8s to kill and restart an otherwise-healthy pod.
//
// Contrast with readinessProbe (readiness.go), which runs registered
// Checkers and can report failure when a dependency is unavailable.
func livenessProbe() http.Handler {
	const body = `{"status":"alive"}`

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}
