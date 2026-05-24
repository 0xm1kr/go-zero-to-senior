package lesson

// toolingLessons covers packages and modules, go fmt/vet/build/test, writing tests, and benchmarking + profiling.
var toolingLessons = []Lesson{
	{
		ID:       "packages",
		Category: "Tooling & Packages",
		Title:    "Packages & Modules",
		Description: `
<p>A <b>package</b> is every .go file in a single directory that shares a
<code>package</code> declaration. A <b>module</b> is a tree of packages versioned
together via <code>go.mod</code>.</p>

<h3>Visibility</h3>
<p><b>Capitalized</b> names are exported (public). <b>lowercase</b> are
package-private. That's the entire access-control story.</p>

<h3>Module commands</h3>
<pre>go mod init github.com/me/myapp   # create go.mod
go get github.com/pkg/errors      # add a dep
go get -u                         # upgrade deps
go mod tidy                       # prune unused, add missing
go mod why github.com/x/y         # explain why a dep is needed</pre>

<h3>Layout convention</h3>
<pre>myapp/
├── go.mod
├── cmd/
│   └── myapp/main.go     # binary
├── internal/             # private to this module
│   └── store/store.go
└── pkg/                  # optional, importable by others
    └── client/client.go</pre>

<p>This lesson's code just uses standard library packages – the point is to
read the notes and try the commands in your terminal.</p>
`,
		Code: `package main

import (
	"fmt"
	"math/rand" // sub-packages use slashes
	"strings"   // standard library
)

func main() {
	fmt.Println(strings.ToUpper("hello packages"))
	rand.Seed(42)
	fmt.Println("random:", rand.Intn(100))
}
`,
		Notes: []string{
			"`internal/` is enforced by the compiler — only the parent module can import it.",
			"`go mod tidy` is the single command you'll run constantly.",
			"goimports (separate tool) auto-adds/removes imports on save — use it in your editor.",
		},
	},
	{
		ID:       "tooling",
		Category: "Tooling & Packages",
		Title:    "go fmt, vet, build, test, run",
		Description: `
<p>Go ships with one official toolchain. Learn these six commands and you're
80% set:</p>
<ul>
  <li><code>go run main.go</code> – compile + run in one step.</li>
  <li><code>go build ./...</code> – compile everything to a binary, no run.</li>
  <li><code>go test ./...</code> – run all tests. Add <code>-race</code>, <code>-cover</code>, <code>-bench</code>.</li>
  <li><code>go fmt ./...</code> – format code (gofmt under the hood). One style, no debates.</li>
  <li><code>go vet ./...</code> – static checks (shadowing, printf format strings, etc).</li>
  <li><code>go mod tidy</code> – fix up dependencies.</li>
</ul>

<h3>Extra tools worth installing</h3>
<ul>
  <li><b>golangci-lint</b> – aggregate linter (the de facto standard in CI).</li>
  <li><b>goimports</b> – like gofmt but also manages imports.</li>
  <li><b>dlv (Delve)</b> – the Go debugger; integrates with VSCode/JetBrains.</li>
  <li><b>air</b> or <b>reflex</b> – live-reload during development.</li>
</ul>

<h3>Cross-compilation</h3>
<p>Set <code>GOOS</code> and <code>GOARCH</code> and you get a static binary for any
platform – no Docker needed.</p>
<pre>GOOS=linux GOARCH=arm64 go build -o myapp ./cmd/myapp</pre>
`,
		Code: `package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("go %s on %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
`,
		Notes: []string{
			"`go build -ldflags=\"-s -w\"` strips symbols → smaller binary.",
			"Run with -race during dev: `go run -race .` catches data races.",
			"Build tags (`//go:build linux`) compile files conditionally.",
		},
	},
	{
		ID:       "testing",
		Category: "Tooling & Packages",
		Title:    "Writing Tests",
		Description: `
<p>Tests live next to the code in <code>*_test.go</code> files. Each test is a
function <code>TestXxx(t *testing.T)</code>. No assertion library is required —
the standard <code>testing</code> package + plain <code>if</code> covers most needs.</p>
<p>Run with <code>go test ./...</code>. Add <code>-v</code> for verbose, <code>-run TestName</code>
to filter, <code>-cover</code> for coverage.</p>
<p>This lesson <i>simulates</i> a test by calling testing-style helpers manually so it can run here. In a real project you'd put this in <code>math_test.go</code> and run <code>go test</code>.</p>
`,
		Code: `package main

import "fmt"

// The code under test.
func Add(a, b int) int { return a + b }

// In a real project this lives in foo_test.go:
//
// func TestAdd(t *testing.T) {
//     cases := []struct{ a, b, want int }{
//         {1, 2, 3},
//         {0, 0, 0},
//         {-1, 1, 0},
//     }
//     for _, c := range cases {
//         got := Add(c.a, c.b)
//         if got != c.want {
//             t.Errorf("Add(%d,%d) = %d, want %d", c.a, c.b, got, c.want)
//         }
//     }
// }

func main() {
	// Manually drive the table to demonstrate the pattern.
	cases := []struct{ a, b, want int }{
		{1, 2, 3},
		{0, 0, 0},
		{-1, 1, 0},
	}
	fails := 0
	for _, c := range cases {
		got := Add(c.a, c.b)
		ok := got == c.want
		if !ok {
			fails++
		}
		fmt.Printf("Add(%d,%d) = %d want %d  pass=%v\n", c.a, c.b, got, c.want, ok)
	}
	fmt.Printf("%d failures\n", fails)
}
`,
		Notes: []string{
			"Table-driven tests are THE idiom in Go — see this code's pattern.",
			"t.Run(\"name\", func(t *testing.T){...}) creates subtests with filtering and parallelism.",
			"`go test -run TestAdd -v` to focus on one test.",
		},
	},
	{
		ID:       "benchmarks",
		Category: "Tooling & Packages",
		Title:    "Benchmarks & Profiling",
		Description: `
<p>Same testing package, function signature <code>BenchmarkXxx(b *testing.B)</code>.
Run with <code>go test -bench=. -benchmem</code>.</p>
<p>For deeper profiling: <code>go test -cpuprofile=cpu.out</code> then
<code>go tool pprof cpu.out</code>. Profiles can be visualized in a browser.</p>
`,
		Code: `package main

import (
	"fmt"
	"strings"
	"time"
)

// Two strategies for the same problem — let's measure.
func concatLoop(parts []string) string {
	s := ""
	for _, p := range parts {
		s += p
	}
	return s
}

func concatBuilder(parts []string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

func time1(name string, f func()) {
	start := time.Now()
	f()
	fmt.Printf("%-15s %s\n", name, time.Since(start))
}

func main() {
	parts := make([]string, 10000)
	for i := range parts {
		parts[i] = "x"
	}
	time1("concatLoop", func() { _ = concatLoop(parts) })
	time1("concatBuilder", func() { _ = concatBuilder(parts) })
}
`,
		Notes: []string{
			"strings.Builder avoids O(n²) string copies — huge difference on 10k+ items.",
			"`-benchmem` reports allocations per op — usually the biggest wins.",
			"benchstat (separate tool) compares before/after benchmark results statistically.",
		},
	},
}
