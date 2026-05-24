package lesson

// controlFlowLessons covers if/for/switch, functions, the functional options idiom, closures, and defer/panic/recover.
var controlFlowLessons = []Lesson{
	{
		ID:       "control-flow",
		Category: "Control Flow",
		Title:    "if / for / switch",
		Description: `
<p>Three control structures. No while, no do/while — <code>for</code> covers all of them.
No parentheses around conditions. Braces are mandatory.</p>
<ul>
  <li><code>if</code> can declare a variable scoped to the if/else block.</li>
  <li><code>for</code> has three forms: C-style, while-style, and infinite.</li>
  <li><code>switch</code> has no fall-through by default (the opposite of C). Cases can be expressions, not just constants.</li>
</ul>
`,
		Code: `package main

import "fmt"

func main() {
	// if with initializer
	if n := 7; n%2 == 0 {
		fmt.Println("even")
	} else {
		fmt.Println("odd")
	}

	// classic for
	sum := 0
	for i := 1; i <= 5; i++ {
		sum += i
	}
	fmt.Println("sum 1..5 =", sum)

	// while-style for
	n := 1
	for n < 100 {
		n *= 2
	}
	fmt.Println("doubled past 100:", n)

	// range over a slice
	for i, v := range []string{"go", "is", "fun"} {
		fmt.Printf("%d:%s ", i, v)
	}
	fmt.Println()

	// switch on expression
	grade := 85
	switch {
	case grade >= 90:
		fmt.Println("A")
	case grade >= 80:
		fmt.Println("B")
	default:
		fmt.Println("C or below")
	}
}
`,
		Notes: []string{
			"for i, v := range coll — v is a COPY. Modify coll[i] if you need to mutate.",
			"switch has no break needed — and supports `fallthrough` if you really want it.",
			"`for { ... }` is the infinite loop. Combine with break / return to exit.",
		},
	},
	{
		ID:       "functions",
		Category: "Control Flow",
		Title:    "Functions & Multiple Returns",
		Description: `
<p>Functions are declared with <code>func</code>. The return type goes <i>after</i> the
parameters. Functions can return multiple values – this is how Go handles
errors instead of exceptions.</p>
<p>Named returns let you treat returns like pre-declared variables. <code>...T</code>
parameters are variadic.</p>
`,
		Code: `package main

import (
	"fmt"
	"strings"
)

// Multiple return values — common for (result, error) pairs.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide by zero")
	}
	return a / b, nil
}

// Named returns + naked return.
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

// Variadic.
func join(sep string, parts ...string) string {
	return strings.Join(parts, sep)
}

func main() {
	q, err := divide(10, 4)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("quotient:", q)
	}

	a, b := split(17)
	fmt.Println("split:", a, b)

	fmt.Println(join("-", "go", "is", "great"))
}
`,
		Notes: []string{
			"Idiomatic error handling: `if err != nil { return err }` everywhere.",
			"Named returns improve docs but can hurt readability in long functions.",
			"Pass variadic with `...`: parts := []string{\"a\",\"b\"}; join(\",\", parts...)",
		},
	},
	{
		ID:       "variadic-options",
		Category: "Control Flow",
		Title:    "Functional Options Pattern",
		Description: `
<p>Variadic functions take 0+ args of one type. They power one of Go's
most-loved idioms: the <b>functional options pattern</b>, which gives you
named, optional, extensible "keyword arguments" without keyword-argument
syntax.</p>

<p>In TypeScript you'd reach for an options object:
<code>new Server({port, tls, timeout})</code>. Go doesn't have keyword
args, but functional options achieve the same thing AND let you add new
options in v2 without breaking v1 callers.</p>

<p>Used by gRPC, the AWS SDK, OpenTelemetry, and basically every modern
Go library.</p>
`,
		Code: `package main

import (
	"fmt"
	"time"
)

type Server struct {
	addr    string
	timeout time.Duration
	tls     bool
}

// Option mutates a *Server. Each option is just a function value.
type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }
func WithTLS() Option                    { return func(s *Server) { s.tls = true } }

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{addr: addr, timeout: 5 * time.Second} // defaults
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func main() {
	a := NewServer(":8080")
	b := NewServer(":443", WithTLS(), WithTimeout(30*time.Second))
	fmt.Printf("a = %+v\nb = %+v\n", a, b)
}
`,
		Notes: []string{
			"Variadic params arrive as a slice inside the function (opts is a []Option).",
			"Pass a slice with `...`: opts := []Option{WithTLS()}; NewServer(\":80\", opts...).",
			"Adding a new option later is backward-compatible — existing callers don't change.",
			"Compare to TS: `class Server { constructor(opts: ServerOpts = {}) {...} }` — same goal, different idiom.",
		},
	},
	{
		ID:       "closures",
		Category: "Control Flow",
		Title:    "Closures & First-Class Functions",
		Description: `
<p>Functions in Go are first-class values: you can pass them as arguments,
return them, store them in fields, and they form <b>closures</b> over their
surrounding scope.</p>
`,
		Code: `package main

import "fmt"

// Returns a closure that captures its own count.
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

// Higher-order function — takes a func parameter.
func mapInts(in []int, f func(int) int) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func main() {
	c := counter()
	fmt.Println(c(), c(), c())

	doubled := mapInts([]int{1, 2, 3, 4}, func(n int) int { return n * 2 })
	fmt.Println(doubled)
}
`,
		Notes: []string{
			"Closures are how you build functional helpers, decorators, and middleware.",
			"There is no built-in `map` / `filter` — you write tiny generic helpers (or use Go 1.21+ slices package).",
		},
	},
	{
		ID:       "defer",
		Category: "Control Flow",
		Title:    "defer, panic, recover",
		Description: `
<p><code>defer</code> schedules a function call to run when the enclosing function
returns – regardless of how it returns. It's how you ensure cleanup
(closing files, unlocking mutexes, etc).</p>
<p><code>panic</code> stops normal execution; <code>recover</code> catches a panic inside
a deferred function. Use sparingly – panics are for <i>unrecoverable</i> bugs,
not control flow. Real errors flow back as <code>error</code> return values.</p>
`,
		Code: `package main

import "fmt"

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil
}

func main() {
	// Deferred calls run in LIFO order.
	defer fmt.Println("third")
	defer fmt.Println("second")
	defer fmt.Println("first")

	fmt.Println("running...")

	r, err := safeDivide(10, 0)
	fmt.Println("result:", r, "err:", err)
}
`,
		Notes: []string{
			"Always pair resource acquisition with defer: f, _ := os.Open(...); defer f.Close().",
			"Deferred args are evaluated immediately, but the CALL runs later.",
			"Don't use panic/recover for normal errors — return them.",
		},
	},
}
