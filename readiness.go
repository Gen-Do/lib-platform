package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

const (
	statusOK    = "OK"
	statusError = "ERROR"
)

type CheckerFunc func() (interface{}, error)

func (f CheckerFunc) Check() (interface{}, error) { return f() }

type readinessInfo struct {
	IP              string                   `json:"ip,omitempty"`
	Host            string                   `json:"host,omitempty"`
	OS              string                   `json:"os,omitempty"`
	Language        string                   `json:"language,omitempty"`
	LanguageVersion string                   `json:"languageVersion,omitempty"`
	GitCommit       string                   `json:"gitCommit,omitempty"`
	BuildID         string                   `json:"buildId,omitempty"`
	BuildUID        string                   `json:"buildUid,omitempty"`
	Resources       map[string]*readinessRes `json:"resources,omitempty"`
}

func (i *readinessInfo) check(logger Logger) bool {
	if len(i.Resources) == 0 {
		return true
	}

	var wg sync.WaitGroup
	for _, r := range i.Resources {
		wg.Add(1)
		go func(f func(Logger)) {
			f(logger)
			wg.Done()
		}(r.check)
	}
	wg.Wait()

	for _, r := range i.Resources {
		if r.Status == statusError {
			return false
		}
	}

	return true
}

type readinessRes struct {
	Status  string      `json:"status,omitempty"`
	Error   string      `json:"error,omitempty"`
	Info    interface{} `json:"info,omitempty"`
	name    string
	checker Checker
}

func (r *readinessRes) check(logger Logger) {
	var err error
	r.Info, err = r.checker.Check()
	r.Status = statusOK
	r.Error = ""
	if err != nil {
		r.Status = statusError
		r.Error = err.Error()
		if logger != nil {
			ctx := logger.WithField(context.Background(), "probe_name", r.name)
			ctx = logger.WithError(ctx, err)
			logger.Warn(ctx, "readiness probe failed")
		}
	}
}

func readinessProbe(checkers map[string]Checker, logger Logger) http.Handler {
	inf := readinessInfo{
		IP:   execute("hostname", "-I"),
		Host: execute("uname", "-n"),
		OS: strings.Join([]string{
			execute("uname", "-s"),
			execute("uname", "-r"),
			execute("uname", "-v"),
			execute("uname", "-m"),
		}, "  "),
		Language:        "go",
		LanguageVersion: runtime.Version(),
		Resources:       nil,
	}

	if len(checkers) > 0 {
		inf.Resources = make(map[string]*readinessRes)
		for name, checker := range checkers {
			inf.Resources[name] = &readinessRes{checker: checker, name: name}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		if !inf.check(logger) {
			status = http.StatusInternalServerError
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		body, _ := json.MarshalIndent(&inf, "", "  ")
		_, _ = w.Write(body)
	})
}

func execute(name string, args ...string) string {
	var out bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}
