package platform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// listenerFunc adapts a plain func to the Listener interface without
// actually starting a network listener - useful to capture the final mux
// that listen() produced and drive it directly through httptest.
type listenerFunc func(ctx context.Context, port int, h http.Handler) error

func (f listenerFunc) Listen(ctx context.Context, port int, h http.Handler) error {
	return f(ctx, port, h)
}

func captureMux(t *testing.T, opts ...Option) http.Handler {
	t.Helper()

	var captured http.Handler
	lsn := listenerFunc(func(_ context.Context, _ int, h http.Handler) error {
		captured = h
		return nil
	})

	if err := Listen(context.Background(), lsn, opts...); err != nil {
		t.Fatalf("Listen returned unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("listener.Listen was not invoked, mux was not captured")
	}
	return captured
}

func doRequest(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// (a) Without any new options, behavior must be identical to v1.0.1:
// "/health" is registered by the library, and the new "/livez" is always
// registered too.
func TestListen_Default_RegistersHealthAndLivez(t *testing.T) {
	mux := http.NewServeMux()
	h := captureMux(t, WithMux(mux))

	healthRec := doRequest(h, readinessProbeEndpoint)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected /health to respond 200 by default (no checkers), got %d", healthRec.Code)
	}

	livezRec := doRequest(h, livenessProbeEndpoint)
	if livezRec.Code != http.StatusOK {
		t.Fatalf("expected /livez to respond 200, got %d", livezRec.Code)
	}
	if got := livezRec.Body.String(); got != `{"status":"alive"}` {
		t.Fatalf("unexpected /livez body: %q", got)
	}
}

// (b) With WithoutReadinessProbe(), the library must NOT register any
// handler on "/health" - a service that owns "/health" itself must see its
// own handler invoked, not the library's readiness probe.
func TestListen_WithoutReadinessProbe_DoesNotRegisterHealth(t *testing.T) {
	mux := http.NewServeMux()

	const serviceOwnedBody = "service-owned-health"
	mux.HandleFunc(readinessProbeEndpoint, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot) // distinguishable marker status
		_, _ = w.Write([]byte(serviceOwnedBody))
	})

	// If listen() still tried to call mux.Handle("/health", ...) here, the
	// stdlib http.ServeMux would panic with "multiple registrations for
	// /health" - so a non-panicking Listen() call is itself part of the
	// assertion.
	h := captureMux(t, WithMux(mux), WithoutReadinessProbe())

	rec := doRequest(h, readinessProbeEndpoint)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected service-owned /health handler (418), got %d", rec.Code)
	}
	if rec.Body.String() != serviceOwnedBody {
		t.Fatalf("expected service-owned /health body, got %q", rec.Body.String())
	}

	// /livez must still be registered unconditionally.
	livezRec := doRequest(h, livenessProbeEndpoint)
	if livezRec.Code != http.StatusOK {
		t.Fatalf("expected /livez to respond 200, got %d", livezRec.Code)
	}
}

// (c) /livez must always be 200 and must not depend on registered
// Checkers - even when a checker is failing (which makes /health report
// 500), liveness must stay healthy.
func TestLiveness_AlwaysOK_RegardlessOfCheckers(t *testing.T) {
	failingChecker := CheckerFunc(func() (interface{}, error) {
		return nil, errors.New("db down")
	})

	mux := http.NewServeMux()
	h := captureMux(t, WithMux(mux), WithChecker("db", failingChecker))

	healthRec := doRequest(h, readinessProbeEndpoint)
	if healthRec.Code != http.StatusInternalServerError {
		t.Fatalf("expected /health to report failure via failing checker, got %d", healthRec.Code)
	}

	livezRec := doRequest(h, livenessProbeEndpoint)
	if livezRec.Code != http.StatusOK {
		t.Fatalf("expected /livez to stay 200 even when a checker fails, got %d", livezRec.Code)
	}
}

// /livez must also be always-OK when the readiness probe is disabled
// entirely, and it must be safe to hit concurrently (exercised under -race).
func TestLiveness_ConcurrentRequests(t *testing.T) {
	mux := http.NewServeMux()
	h := captureMux(t, WithMux(mux), WithoutReadinessProbe())

	const n = 20
	done := make(chan int, n)
	for i := 0; i < n; i++ {
		go func() {
			done <- doRequest(h, livenessProbeEndpoint).Code
		}()
	}
	for i := 0; i < n; i++ {
		if code := <-done; code != http.StatusOK {
			t.Fatalf("concurrent /livez request returned %d, want 200", code)
		}
	}
}
