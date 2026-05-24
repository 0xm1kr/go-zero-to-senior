package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DefaultPlaygroundURL is the public Go Playground compile endpoint.
// Override at runtime with $PLAYGROUND_URL (e.g. to point at a self-hosted
// instance — see https://github.com/golang/playground).
const DefaultPlaygroundURL = "https://go.dev/_/compile"

// Playground executes Go source via the public Go Playground API.
//
// Compared to Local, user code:
//   - never touches the server's filesystem
//   - cannot reach the network
//   - runs inside Google's hardened sandbox VM
//   - is capped at ~5s wall-clock by the Playground itself
//
// This is the only safe backend for a public-facing deployment.
//
// Note that the Playground is a shared public service. For heavy use
// you should self-host (the Playground is open source) and set
// $PLAYGROUND_URL accordingly.
type Playground struct {
	URL     string        // compile endpoint
	Timeout time.Duration // request deadline (HTTP layer)
	Client  *http.Client
}

// NewPlayground returns a Playground client. Timeout caps the HTTP round-
// trip (plus a small buffer for the Playground's own ~5s execution limit).
func NewPlayground(timeout time.Duration) *Playground {
	u := strings.TrimSpace(os.Getenv("PLAYGROUND_URL"))
	if u == "" {
		u = DefaultPlaygroundURL
	}
	return &Playground{
		URL:     u,
		Timeout: timeout,
		Client:  &http.Client{Timeout: timeout + 5*time.Second},
	}
}

// Backend identifies this implementation in startup logs.
func (p *Playground) Backend() string { return "playground (" + p.URL + ")" }

// playgroundEvent is one stdout/stderr chunk returned by /compile.
type playgroundEvent struct {
	Message string `json:"Message"`
	Kind    string `json:"Kind"` // "stdout" or "stderr"
	Delay   int64  `json:"Delay"`
}

// playgroundResponse matches the Go Playground's JSON response shape.
type playgroundResponse struct {
	Errors    string            `json:"Errors"` // non-empty = compile/build failure
	Events    []playgroundEvent `json:"Events"`
	VetErrors string            `json:"VetErrors"`
	VetOK     bool              `json:"VetOK"`
}

// Run posts code to the configured /compile endpoint and translates the
// response into a runner.Result the rest of the app already understands.
func (p *Playground) Run(ctx context.Context, code string) Result {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout+5*time.Second)
	defer cancel()

	form := url.Values{}
	form.Set("version", "2")
	form.Set("body", code)
	form.Set("withVet", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, strings.NewReader(form.Encode()))
	if err != nil {
		return Result{Err: fmt.Errorf("playground: build request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "golang-tut/1.0 (+https://github.com/0xm1kr)")

	res, err := p.Client.Do(req)
	if err != nil {
		return Result{Err: fmt.Errorf("playground: call: %w", err)}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return Result{Err: fmt.Errorf("playground: HTTP %d: %s",
			res.StatusCode, strings.TrimSpace(string(body)))}
	}

	var resp playgroundResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return Result{Err: fmt.Errorf("playground: decode: %w", err)}
	}

	var stdout, stderr strings.Builder
	for _, e := range resp.Events {
		switch e.Kind {
		case "stderr":
			stderr.WriteString(e.Message)
		default: // "stdout" and anything unexpected
			stdout.WriteString(e.Message)
		}
	}

	out := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if msg := strings.TrimSpace(resp.Errors); msg != "" {
		// Compile / build failure. Surface it on stderr too so users see
		// the diagnostic in the same place as runtime stderr.
		if out.Stderr == "" {
			out.Stderr = msg + "\n"
		}
		out.Err = errors.New(msg)
	}
	return out
}
