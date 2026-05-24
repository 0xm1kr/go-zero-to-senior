package lesson

// memoryLessons covers value vs pointer semantics, escape analysis, and sync.Pool object reuse.
var memoryLessons = []Lesson{
	{
		ID:       "value-vs-pointer",
		Category: "Memory & Performance",
		Title:    "Value vs Pointer Semantics",
		Description: `
<p>Go is the language where this distinction matters MOST. Every type is
either copied (value) or referenced (pointer) when assigned/passed.
Choose wrong and you'll have bugs that look like the laws of physics
broke.</p>

<h3>Decision rules every senior engineer knows</h3>
<ul>
  <li><b>Mutate the receiver?</b> → pointer receiver.</li>
  <li><b>Struct is &gt; ~64 bytes?</b> → pointer (avoid copy cost).</li>
  <li><b>Contains a sync.Mutex / sync.WaitGroup?</b> → pointer (these MUST NOT be copied).</li>
  <li><b>Slice/map/channel/func/interface field?</b> → those are already reference-ish; value receiver is usually fine.</li>
  <li><b>Pure value, immutable, small?</b> → value receiver. Easier reasoning.</li>
</ul>

<p>Don't mix value and pointer receivers on the same type — it confuses
both you and Go's interface-satisfaction rules.</p>
`,
		Code: `package main

import "fmt"

type Counter struct{ n int }

// Pointer receiver: mutates the actual Counter.
func (c *Counter) Inc() { c.n++ }

// Value receiver: works on a COPY. Mutations are lost.
func (c Counter) IncBroken() { c.n++ }

func main() {
	a := Counter{}
	a.Inc()
	a.Inc()
	fmt.Println("a.n after Inc x2:", a.n) // 2

	b := Counter{}
	b.IncBroken()
	b.IncBroken()
	fmt.Println("b.n after IncBroken x2:", b.n) // 0 — silently wrong

	// Method-set rule: only *Counter satisfies an interface { Inc() }.
	var ptr interface{ Inc() } = &a
	ptr.Inc()
	fmt.Println("via interface:", a.n) // 3
	// var notOk interface{ Inc() } = a  // compile error: Inc has *Counter receiver
}
`,
		Notes: []string{
			"Method-set rule: *T includes BOTH value- and pointer-receiver methods; T includes only value-receiver methods.",
			"Classic bug: `func (c Counter) Inc() { c.n++ }` compiles, runs, and silently does nothing.",
			"Never copy a sync.Mutex / WaitGroup / Once. `go vet` catches this — read its warnings.",
			"Pointers don't always escape to the heap — the compiler's escape analysis decides (next lesson).",
		},
	},
	{
		ID:       "escape-analysis",
		Category: "Memory & Performance",
		Title:    "Escape Analysis & Allocations",
		Description: `
<p>"Stack or heap?" In Go, you don't choose — the compiler does, via
<b>escape analysis</b>. A value <i>escapes</i> when its lifetime exceeds
the containing function's frame.</p>

<p>Stack allocations are essentially free. Heap allocations cost GC
pressure. Senior optimization in Go is often "rearrange code so this
doesn't escape."</p>

<h3>Common reasons a value escapes</h3>
<ul>
  <li>You returned a pointer to a local.</li>
  <li>You stored it in an interface (boxing).</li>
  <li>You captured it in a closure that outlives the frame.</li>
  <li>The compiler couldn't prove the lifetime is bounded.</li>
</ul>

<h3>See the analysis</h3>
<pre>go build -gcflags='-m' ./...
# /tmp/x.go:7:6: moved to heap: u
# /tmp/x.go:12:13: ... argument does not escape</pre>

<p>Pair with <code>go test -bench . -benchmem</code> to see allocs/op on
any hot path.</p>
`,
		Code: `package main

import "fmt"

type Point struct{ X, Y int }

// Does NOT escape: caller gets a copy by value.
func makePointValue() Point {
	return Point{X: 1, Y: 2}
}

// ESCAPES: returning a pointer means the value must outlive this frame.
func makePointHeap() *Point {
	return &Point{X: 1, Y: 2}
}

// Interface boxing: storing a concrete value in an interface heap-allocates.
func boxIt(x int) interface{} {
	return x
}

func main() {
	a := makePointValue()
	b := makePointHeap()
	boxed := boxIt(42)
	fmt.Println(a, b, boxed)
}
`,
		Notes: []string{
			"`go build -gcflags='-m' file.go` shows exactly what escapes and why.",
			"Premature pointer-everywhere is an anti-pattern; values fit on the stack and avoid GC.",
			"interface{} / any always boxes — high-perf code sticks to concrete types and generics.",
			"Slice growth (append past cap) reallocates — preallocate with make([]T, 0, n) when you know n.",
		},
	},
	{
		ID:       "sync-pool",
		Category: "Memory & Performance",
		Title:    "sync.Pool & Object Reuse",
		Description: `
<p>For per-request scratch objects (buffers, encoders, parsers) that
generate GC pressure, <code>sync.Pool</code> is the standard solution.
It's a thread-safe pool of reusable values; Get() returns a pooled value
or makes a new one, Put() returns it.</p>

<p>Used heavily by net/http, encoding/json, and most high-throughput
servers. Canonical example: per-request <code>bytes.Buffer</code>.</p>

<h3>Pool semantics gotchas</h3>
<ul>
  <li>The pool may DROP items between GC cycles. Treat Get() as "maybe a fresh one." ALWAYS Reset before reuse.</li>
  <li>Put a value back only when you're done with it. Aliasing leaks bugs.</li>
  <li>Don't pool tiny structs — bookkeeping costs more than the alloc you save.</li>
</ul>
`,
		Code: `package main

import (
	"bytes"
	"fmt"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func render(name string) string {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset() // CRITICAL: don't pool dirty state
		bufPool.Put(buf)
	}()

	fmt.Fprintf(buf, "<h1>hello, %s</h1>", name)
	return buf.String()
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println(render(fmt.Sprintf("user-%d", i)))
		}(i)
	}
	wg.Wait()
}
`,
		Notes: []string{
			"ALWAYS Reset() before Put() — otherwise the next consumer sees stale data.",
			"The pool's New func is the fallback when Get() finds nothing pooled.",
			"Don't pool tiny structs (<16 bytes); the bookkeeping costs more than you save.",
			"Benchmark before AND after with `go test -bench . -benchmem`. If allocs/op didn't move, remove the Pool.",
		},
	},
}
