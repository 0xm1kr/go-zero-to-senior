package lesson

// ecosystemLessons covers popular Go frameworks and where Go shines as a language.
var ecosystemLessons = []Lesson{
	{
		ID:       "frameworks",
		Category: "Ecosystem",
		Title:    "Frameworks: Gin, Echo, Fiber, Chi",
		Description: `
<p>The standard library is excellent, but here's when you might reach for a
framework:</p>

<table style="width:100%;border-collapse:collapse;margin-top:8px">
<tr style="background:#1d2230;text-align:left">
  <th style="padding:6px">Framework</th><th>Best for</th><th>Notes</th>
</tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><b>chi</b></td>
<td style="border-top:1px solid #2a3040">REST APIs that stay close to net/http</td>
<td style="border-top:1px solid #2a3040">100% compatible with std handlers; minimal magic. Great default pick.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><b>Gin</b></td>
<td style="border-top:1px solid #2a3040">JSON APIs, productivity</td>
<td style="border-top:1px solid #2a3040">Most popular, big middleware ecosystem, fast.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><b>Echo</b></td>
<td style="border-top:1px solid #2a3040">Similar niche to Gin</td>
<td style="border-top:1px solid #2a3040">Cleaner API IMO, built-in validator, binding.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><b>Fiber</b></td>
<td style="border-top:1px solid #2a3040">Express.js-like, max throughput</td>
<td style="border-top:1px solid #2a3040">Built on fasthttp (NOT net/http) — incompatible with std middleware.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><b>Buffalo</b></td>
<td style="border-top:1px solid #2a3040">Full-stack Rails-like</td>
<td style="border-top:1px solid #2a3040">Less common today; most Go shops prefer composable libraries.</td></tr>
</table>

<h3>Other libraries you'll meet</h3>
<ul>
  <li><b>sqlx</b> / <b>sqlc</b> / <b>GORM</b> – database access (sqlc generates typed code from SQL, the modern favorite).</li>
  <li><b>pgx</b> – the high-perf Postgres driver.</li>
  <li><b>zap</b> / <b>zerolog</b> – structured logging (or std <code>log/slog</code> in 1.21+).</li>
  <li><b>cobra</b> + <b>viper</b> – CLI framework + config.</li>
  <li><b>testify</b> – assertion library + mocks if std testing feels too bare.</li>
  <li><b>wire</b> – compile-time dependency injection (from Google).</li>
  <li><b>protobuf</b> + <b>grpc-go</b> – the standard for gRPC services.</li>
</ul>

<p>The code below uses only the standard library to recreate a Gin-style
route handler — to show how thin most frameworks really are.</p>
`,
		Code: `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// Tiny chi/Gin-like helper for JSON responses.
type Ctx struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Ctx) JSON(status int, body interface{}) {
	c.W.Header().Set("Content-Type", "application/json")
	c.W.WriteHeader(status)
	json.NewEncoder(c.W).Encode(body)
}

func wrap(h func(*Ctx)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { h(&Ctx{w, r}) }
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/me", wrap(func(c *Ctx) {
		c.JSON(200, map[string]string{"name": "Gopher"})
	}))

	srv := httptest.NewServer(mux)
	defer srv.Close()
	res, _ := http.Get(srv.URL + "/me")
	defer res.Body.Close()
	var m map[string]string
	json.NewDecoder(res.Body).Decode(&m)
	fmt.Printf("status=%s body=%v\n", res.Status, m)
}
`,
		Notes: []string{
			"Most frameworks are ~500 lines of helpers on top of net/http.",
			"For new projects in 2026: start with chi or Echo, add sqlc, log/slog, cobra — that's a complete stack.",
			"Avoid Fiber unless throughput is your dominant concern (it doesn't compose with std middleware).",
		},
	},
	{
		ID:       "use-cases",
		Category: "Ecosystem",
		Title:    "What Go Is Great At",
		Description: `
<p>Go's sweet spot is <b>backend systems</b> where you'd otherwise reach for
Java, C++, or a heavyweight Node service.</p>

<h3>Where Go shines</h3>
<ul>
  <li><b>HTTP/gRPC microservices</b> – fast start, low memory, easy concurrency.</li>
  <li><b>CLI tools</b> – single static binary cross-compiled for every OS (Docker, kubectl, gh, terraform, hugo).</li>
  <li><b>Network proxies & load balancers</b> – Traefik, Caddy, Envoy bindings.</li>
  <li><b>DevOps & infra tooling</b> – the entire cloud-native ecosystem (Kubernetes, Prometheus, etcd, containerd) is Go.</li>
  <li><b>Data pipelines / stream processors</b> – channels + goroutines map naturally to pipelines.</li>
  <li><b>SaaS backends</b> – many fintech, infrastructure, and SaaS startups standardize on Go.</li>
</ul>

<h3>Where Go is okay but not the best</h3>
<ul>
  <li><b>Frontends & UI</b> – Go can compile to WASM, but it's not the goal.</li>
  <li><b>Data science / ML</b> – ecosystem is thinner than Python's; usually call out to Python.</li>
  <li><b>Game engines</b> – GC pauses and lack of generics-heavy patterns hurt; people use C++/Rust.</li>
  <li><b>Embedded / no_std</b> – Go has a runtime/GC; for tiny MCUs use Rust or C.</li>
</ul>

<h3>Recommended next steps</h3>
<ol>
  <li>Build a small JSON CRUD API + Postgres (use chi + sqlc + pgx).</li>
  <li>Write a CLI tool with cobra + viper.</li>
  <li>Read the <a href="https://go.dev/doc/effective_go" target="_blank">Effective Go</a> guide.</li>
  <li>Skim the standard library docs (https://pkg.go.dev/std) – the stdlib is your superpower.</li>
  <li>Read source of one tool you already use (kubectl, hugo, gh) – best way to learn idioms.</li>
</ol>
`,
		Code: `package main

import "fmt"

// A self-contained example of a "next step" challenge:
// build the smallest possible TODO API in your head, then go write it.
type Todo struct {
	ID    int
	Title string
	Done  bool
}

func main() {
	store := map[int]Todo{}
	store[1] = Todo{1, "Read Effective Go", false}
	store[2] = Todo{2, "Build a CLI", false}

	for id, t := range store {
		fmt.Printf("#%d %-25s done=%v\n", id, t.Title, t.Done)
	}
	fmt.Println("\nNow go build the HTTP version with chi.")
}
`,
		Notes: []string{
			"Pick ONE small project and finish it — Go rewards small + done over big + planned.",
			"Read other people's Go more than you write your own at first; the idioms are subtle.",
			"Join the Gophers Slack and r/golang for daily reading.",
		},
	},
}
