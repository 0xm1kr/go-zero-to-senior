package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Local executes Go source via `go run` in a fresh temp directory on the
// host. The temp dir (including its private GOCACHE) is deleted before
// Run returns.
//
// SECURITY: user code runs with the same UID and filesystem access as the
// server process. Suitable for `go run .` on your laptop; NOT suitable for
// anything reachable from the public internet. Use Playground there.
type Local struct {
	Timeout time.Duration
	GoBin   string // path to the `go` binary; defaults to "go" on $PATH
}

// NewLocal returns a Local with sane defaults.
func NewLocal(timeout time.Duration) *Local {
	return &Local{Timeout: timeout, GoBin: "go"}
}

// Backend identifies this implementation in startup logs.
func (l *Local) Backend() string { return "local (go run)" }

// Run writes code to a temp dir, executes it under `go run`, and returns
// stdout / stderr / Err. A timeout is enforced via context.
func (l *Local) Run(ctx context.Context, code string) Result {
	dir, err := os.MkdirTemp("", "gotut-*")
	if err != nil {
		return Result{Err: err}
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte(code), 0o600); err != nil {
		return Result{Err: err}
	}

	ctx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, l.GoBin, "run", file)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(dir, "gocache"),
		"GOFLAGS=-mod=mod",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return Result{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Err:    fmt.Errorf("execution timed out after %s", l.Timeout),
		}
	}
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Err: runErr}
}
