package lesson

// webLessons covers net/http (server and client), middleware composition, and production-grade HTTP servers.
var webLessons = []Lesson{
	{
		ID:       "http-server",
		Category: "Web Development",
		Title:    "HTTP Server with net/http",
		Description: `
<p>You can build a production-grade HTTP server with only the standard
library. This is the exact pattern this tutorial app uses (see
<code>main.go</code>).</p>
`,
		Code: `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type pingResponse struct {
	Message string ` + "`json:\"message\"`" + `
	Echo    string ` + "`json:\"echo,omitempty\"`" + `
}

func handler(w http.ResponseWriter, r *http.Request) {
	resp := pingResponse{
		Message: "pong",
		Echo:    r.URL.Query().Get("say"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", handler)

	// httptest spins up an in-process server — perfect for demos & tests.
	srv := httptest.NewServer(mux)
	defer srv.Close()

	res, _ := http.Get(srv.URL + "/ping?say=hello")
	defer res.Body.Close()
	var out pingResponse
	json.NewDecoder(res.Body).Decode(&out)
	fmt.Printf("status=%s body=%+v\n", res.Status, out)
}
`,
		Notes: []string{
			"http.HandlerFunc adapts a plain function into an http.Handler.",
			"Middleware = function that takes http.Handler and returns http.Handler.",
			"Go 1.22+ adds method+pattern matching: mux.HandleFunc(\"GET /users/{id}\", h).",
		},
	},
	{
		ID:       "http-client",
		Category: "Web Development",
		Title:    "HTTP Client & JSON APIs",
		Description: `
<p>The same <code>net/http</code> package gives you a client. For production
use, define your own <code>http.Client</code> with a timeout – the default has none.</p>
`,
		Code: `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// Pretend remote service.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, ` + "`{\"id\":1,\"name\":\"Alice\"}`" + `)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", srv.URL+"/users/1", nil)
	req.Header.Set("Authorization", "Bearer demo")

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var user struct {
		ID   int    ` + "`json:\"id\"`" + `
		Name string ` + "`json:\"name\"`" + `
	}
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		panic(err)
	}
	fmt.Printf("got user: %+v\n", user)
}
`,
		Notes: []string{
			"ALWAYS set a Timeout on http.Client — the default is unlimited.",
			"Always `defer res.Body.Close()` or you leak file descriptors.",
			"Reuse one http.Client across requests; it pools connections.",
		},
	},
	{
		ID:       "middleware",
		Category: "Web Development",
		Title:    "Middleware Composition",
		Description: `
<p>In Go, middleware is just <code>func(http.Handler) http.Handler</code>.
That's it. No framework needed — wrap a handler with cross-cutting
behavior (logging, auth, recovery, tracing) and return the wrapped one.</p>

<p>From Express/Koa, this should look very familiar: the chain is just
function composition, and request context flows through
<code>r.Context()</code>.</p>
`,
		Code: `package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Chain composes middleware in left-to-right execution order.
func Chain(mw ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			next = mw[i](next)
		}
		return next
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s in %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type ctxKey string

const userKey ctxKey = "user"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// pretend we validated the token; stash user on the context
		ctx := context.WithValue(r.Context(), userKey, "alice")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func hello(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(userKey).(string)
	fmt.Fprintf(w, "hi %s\n", user)
}

func main() {
	stack := Chain(Recovery, Logging, Auth)
	http.Handle("/hello", stack(http.HandlerFunc(hello)))
	// Run on your own machine: http.ListenAndServe(":8080", nil)
	fmt.Println("middleware chain assembled (Recovery -> Logging -> Auth -> hello)")
}
`,
		Notes: []string{
			"Order matters: Recovery should be OUTERMOST so it catches panics from later middleware too.",
			"Use a typed context key (`type ctxKey string`) so different packages can't collide.",
			"Don't smuggle business logic into middleware. Cross-cutting only: log, auth, trace, recover, ratelimit.",
			"This pattern is exactly what chi, gorilla/mux, and net/http servemux all use under the hood.",
		},
	},
	{
		ID:       "production-http",
		Category: "Web Development",
		Title:    "Production HTTP Server",
		Description: `
<p>A "Hello, World" net/http server is one line. A <b>production</b>
server is the same one line + four things every senior engineer knows
to add:</p>

<ol>
  <li><b>Timeouts</b> — Read, Write, Idle. The defaults are unlimited.
  Misconfigured servers are how you DoS yourself.</li>
  <li><b>Graceful shutdown</b> — listen for SIGINT/SIGTERM, drain
  in-flight requests with a deadline, then exit.</li>
  <li><b>Health/readiness endpoints</b> — separate from your app
  routes so a load balancer can probe without spamming logs.</li>
  <li><b>Structured logging + recovery middleware</b> — cover this in
  the slog and middleware lessons.</li>
</ol>

<p>This is the canonical "main.go for a Go web service" template
interviewers expect you to reproduce from memory.</p>
`,
		Code: `package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})

	srv := &http.Server{
		Addr:              ":5555", // :8080 is canonical but commonly taken — change to taste
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run server in a goroutine so main can wait for shutdown signal.
	go func() {
		log.Println("listening on", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	// Block until SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	// Give in-flight requests up to 30s to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("bye")
}
`,
		Notes: []string{
			"ReadHeaderTimeout is your cheapest DoS defense — set it even if you set nothing else.",
			"srv.Shutdown stops accepting new conns and waits for active ones to finish (up to ctx deadline).",
			"signal.Notify with SIGINT (Ctrl-C local) AND SIGTERM (what Kubernetes sends to kill pods).",
			"Health probes (/healthz, /readyz) should be cheap, dependency-free, and skip your auth middleware.",
		},
	},
}
