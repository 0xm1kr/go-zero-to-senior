// Package runner executes Go source code submitted from the browser and
// returns its stdout / stderr / error. Two backends are provided:
//
//   - Local       shells out to `go run` in a temp dir on the host.
//   - Playground  POSTs to https://go.dev/_/compile (the Go Playground).
//
// The Playground backend is the only safe choice for any deployment that
// is reachable from the public internet, because the local backend runs
// untrusted code with the same privileges as the server process.
//
// Pick a backend either by constructing it directly or via NewFromEnv,
// which reads $RUNNER.
package runner

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
)

// Runner is the contract the API layer depends on. Both Local and
// Playground implement it.
type Runner interface {
	Run(ctx context.Context, code string) Result
	// Backend returns a short human-readable label, surfaced in startup
	// logs so operators can confirm which executor is wired up.
	Backend() string
}

// Result captures everything the UI needs to render a run.
type Result struct {
	Stdout string
	Stderr string
	Err    error // compile error, non-zero exit, timeout, etc.
}

// NewFromEnv constructs the configured Runner based on $RUNNER:
//
//	RUNNER=playground   -> Go Playground API (safe for cloud deploys)
//	RUNNER=local        -> shell out to `go run` locally
//	unset               -> auto: Playground if running on Cloud Run /
//	                       Knative (detected via $K_SERVICE), else Local
//
// Pass timeout for the per-run hard cap; the Playground also enforces
// its own internal 5-second limit, which usually wins.
func NewFromEnv(timeout time.Duration) Runner {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("RUNNER"))) {
	case "playground", "play":
		return NewPlayground(timeout)
	case "local":
		return NewLocal(timeout)
	case "":
		// Auto-detect: K_SERVICE is set on Cloud Run and other Knative
		// runtimes. Defaulting to Playground there is the safe choice
		// because Local would happily run untrusted code in the container.
		if os.Getenv("K_SERVICE") != "" {
			return NewPlayground(timeout)
		}
		return NewLocal(timeout)
	default:
		log.Printf("runner: unknown RUNNER=%q, falling back to local", os.Getenv("RUNNER"))
		return NewLocal(timeout)
	}
}
