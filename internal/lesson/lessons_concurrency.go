package lesson

// concurrencyLessons covers goroutines, channels, select, sync primitives, context, worker pools, errgroup, and race detection.
var concurrencyLessons = []Lesson{
	{
		ID:       "goroutines",
		Category: "Concurrency",
		Title:    "Goroutines",
		Description: `
<p>A goroutine is a lightweight thread managed by the Go runtime. Start one
with <code>go f()</code>. They cost ~2KB of stack each (vs MB for OS threads),
so spawning thousands is normal.</p>
<p>This example uses <code>sync.WaitGroup</code> to wait for several goroutines to
finish – next lesson shows the channel-based alternative.</p>
`,
		Code: `package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("worker %d done\n", id)
}

func main() {
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}
	wg.Wait()
	fmt.Println("all workers done")
}
`,
		Notes: []string{
			"`go f()` returns immediately — f runs concurrently.",
			"The main goroutine exiting kills the program, even if others are still running.",
			"Race conditions are real: use channels OR sync.Mutex when sharing state.",
		},
	},
	{
		ID:       "channels",
		Category: "Concurrency",
		Title:    "Channels",
		Description: `
<p>Channels are typed pipes that synchronize goroutines: <code>ch &lt;- v</code> sends,
<code>v := &lt;-ch</code> receives. They're the heart of Go's concurrency style:
<i>"Don't communicate by sharing memory; share memory by communicating."</i></p>
<p>Unbuffered channels block until both sides are ready. Buffered channels
(<code>make(chan T, n)</code>) hold up to n values without blocking.</p>
`,
		Code: `package main

import "fmt"

func producer(ch chan<- int) {
	for i := 1; i <= 5; i++ {
		ch <- i * i
	}
	close(ch) // tells receivers no more values are coming
}

func main() {
	ch := make(chan int)
	go producer(ch)

	// range receives until the channel is closed.
	for v := range ch {
		fmt.Println("got:", v)
	}
}
`,
		Notes: []string{
			"Only the SENDER should close a channel. Never close from the receiver side.",
			"Sending on a closed channel panics. Receiving from a closed channel returns the zero value with ok=false.",
			"`chan<- T` is send-only, `<-chan T` is receive-only — useful for restricting APIs.",
		},
	},
	{
		ID:       "select",
		Category: "Concurrency",
		Title:    "select & Timeouts",
		Description: `
<p><code>select</code> is like <code>switch</code> for channel operations: it waits until
one of its cases can proceed. Combined with <code>time.After</code> or
<code>context.Done()</code> you get clean timeouts and cancellation.</p>
`,
		Code: `package main

import (
	"fmt"
	"time"
)

func main() {
	results := make(chan string, 1)

	go func() {
		time.Sleep(200 * time.Millisecond)
		results <- "completed"
	}()

	select {
	case r := <-results:
		fmt.Println(r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("timed out!")
	}
}
`,
		Notes: []string{
			"`default:` makes select non-blocking.",
			"If multiple cases are ready, one is chosen at random — don't rely on order.",
			"Pair select with context.Context for proper cancellation.",
		},
	},
	{
		ID:       "sync",
		Category: "Concurrency",
		Title:    "sync.Mutex, WaitGroup, Once",
		Description: `
<p>When channels feel like overkill, reach for <code>sync</code>. The most common
primitives:</p>
<ul>
  <li><code>sync.Mutex</code> – mutual exclusion, <code>Lock()</code>/<code>Unlock()</code>.</li>
  <li><code>sync.RWMutex</code> – many readers OR one writer.</li>
  <li><code>sync.WaitGroup</code> – wait for N goroutines.</li>
  <li><code>sync.Once</code> – run an initializer exactly once, threadsafe.</li>
</ul>
`,
		Code: `package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	mu sync.Mutex
	n  int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func main() {
	c := &SafeCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Println("final:", c.n) // 1000, every time
}
`,
		Notes: []string{
			"Always `defer mu.Unlock()` right after locking — exception-safe.",
			"Run tests/programs with `-race` to detect data races: go test -race ./...",
			"Atomic operations (sync/atomic) are even faster for simple counters.",
		},
	},
	{
		ID:       "context",
		Category: "Concurrency",
		Title:    "context: Cancellation & Deadlines",
		Description: `
<p><code>context.Context</code> propagates deadlines, cancellation signals, and
request-scoped values across API boundaries. Every standard-library
function that can block — net/http, database/sql, exec — accepts one.</p>
<p>Rule: pass <code>ctx context.Context</code> as the FIRST parameter of any function
that can block.</p>
`,
		Code: `package main

import (
	"context"
	"fmt"
	"time"
)

// Long-running task that respects cancellation.
func work(ctx context.Context) error {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err() // context.Canceled or DeadlineExceeded
		case <-time.After(50 * time.Millisecond):
			fmt.Println("tick", i)
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := work(ctx); err != nil {
		fmt.Println("stopped:", err)
	}
}
`,
		Notes: []string{
			"context.Background() at the top of the call stack; never store a context in a struct.",
			"WithCancel, WithTimeout, WithDeadline, WithValue cover 99% of needs.",
			"`defer cancel()` always — even if you also set a timeout. Frees resources.",
		},
	},
	{
		ID:       "worker-pools",
		Category: "Concurrency",
		Title:    "Worker Pools (Fan-Out / Fan-In)",
		Description: `
<p>Spawning a goroutine per task is fine until "tasks" means "100k tasks
that each open a database connection." Worker pools cap concurrency: N
workers pull from a shared <code>jobs</code> channel and push to a shared
<code>results</code> channel.</p>

<p>Fan-out = many workers consume one channel. Fan-in = many producers
send into one channel. The two together = canonical Go concurrency.</p>

<p>This pattern shows up in EVERY senior Go interview. Know it cold.</p>
`,
		Code: `package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID  int
	URL string
}

type Result struct {
	JobID int
	Body  string
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs { // exits when jobs is closed
		time.Sleep(20 * time.Millisecond) // simulate I/O
		results <- Result{JobID: j.ID, Body: fmt.Sprintf("w%d fetched %s", id, j.URL)}
	}
}

func main() {
	const numWorkers = 3
	jobs := make(chan Job)
	results := make(chan Result)

	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	// Closer goroutine: when all workers finish, close results.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Feed jobs, then close to signal "no more work."
	go func() {
		for i := 1; i <= 8; i++ {
			jobs <- Job{ID: i, URL: fmt.Sprintf("/page/%d", i)}
		}
		close(jobs)
	}()

	for r := range results {
		fmt.Printf("got result job=%d body=%q\n", r.JobID, r.Body)
	}
}
`,
		Notes: []string{
			"Close `jobs` from the producer side to tell workers \"no more.\" Never close from the consumer side.",
			"A separate goroutine waits on the WaitGroup, then closes `results` — that's the fan-in completion signal.",
			"Channel direction in params (<-chan, chan<-) is a senior-level signal of intent. Use them.",
			"Pool size is usually tied to a downstream resource: DB connections, HTTP semaphore, CPU count.",
		},
	},
	{
		ID:       "errgroup",
		Category: "Concurrency",
		Title:    "errgroup: Coordinated Goroutines",
		Description: `
<p><code>golang.org/x/sync/errgroup</code> bundles three things every
"run N things in parallel, fail fast" use case needs:</p>

<ul>
  <li>A <code>WaitGroup</code> that returns the FIRST non-nil error.</li>
  <li>A derived <code>context.Context</code> cancelled the moment any
  goroutine errors — so the others stop immediately.</li>
  <li>An optional <code>SetLimit</code> to cap concurrency.</li>
</ul>

<p>"Implement errgroup from scratch" is a real senior interview
question, so this lesson rolls a minimal version using only stdlib.
It's ~30 lines. In production, use the real package
(<code>go get golang.org/x/sync</code>) — it adds SetLimit, panic
recovery, and is battle-tested.</p>
`,
		Code: `package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Group is a minimal stand-in for x/sync/errgroup.Group.
// It collects the first non-nil error and cancels the derived context.
type Group struct {
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	errOnce sync.Once
	err     error
}

// WithContext returns a Group + a context that is cancelled either when
// Wait() returns OR when the first goroutine returns an error.
func WithContext(parent context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(parent)
	return &Group{cancel: cancel}, ctx
}

// Go launches fn in a goroutine. The first error wins and cancels ctx.
func (g *Group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

// Wait blocks until all Go calls finish, then returns the first error.
func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

func fetch(ctx context.Context, url string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}
	if url == "/boom" {
		return "", errors.New("server exploded")
	}
	return "body of " + url, nil
}

func fetchAll(ctx context.Context, urls []string) (map[string]string, error) {
	g, ctx := WithContext(ctx)

	var mu sync.Mutex
	results := make(map[string]string, len(urls))

	for _, u := range urls {
		g.Go(func() error {
			body, err := fetch(ctx, u)
			if err != nil {
				return fmt.Errorf("%s: %w", u, err)
			}
			mu.Lock()
			defer mu.Unlock()
			results[u] = body
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, err // partial results may still be useful
	}
	return results, nil
}

func main() {
	out, err := fetchAll(context.Background(), []string{"/a", "/b", "/boom", "/c"})
	fmt.Println("err:    ", err)
	fmt.Println("partial:", out)
}
`,
		Notes: []string{
			"errgroup = WaitGroup + first-error + derived-context cancellation. ~30 lines, as shown.",
			"sync.Once around setting `err` and calling cancel() is the trick — guarantees first error wins.",
			"Real x/sync/errgroup adds SetLimit (concurrency cap, Go 1.20+) and panic recovery — use it in prod.",
			"Go 1.22 fixed for-loop variable capture, so `for _, u := range urls { g.Go(...) }` is safe without `u := u`.",
		},
	},
	{
		ID:       "race-detection",
		Category: "Concurrency",
		Title:    "Race Conditions & the Race Detector",
		Description: `
<p>A data race is two goroutines accessing the same memory location with
at least one write, without synchronization. In Go, races are
<b>undefined behavior</b> — the runtime makes no guarantees.</p>

<p>The fix is always: use a channel, a mutex, an atomic, or just don't
share. The detector is one of Go's killer features.</p>

<h3>The race detector</h3>
<pre>go run -race main.go
go test -race ./...
go build -race main.go</pre>

<p>Adds ~5x runtime overhead but catches races at runtime. Run your test
suite with <code>-race</code> in CI — non-negotiable for production Go.</p>

<p>The example below intentionally races. Run with <code>-race</code> on
your own machine to see the report (the embedded runner here doesn't
enable -race by default).</p>
`,
		Code: `package main

import (
	"fmt"
	"sync"
)

// BAD: classic race. Without the mutex, the final value is undefined.
type BrokenCounter struct {
	n int
}

func (c *BrokenCounter) Inc() { c.n++ } // read-modify-write, NOT atomic

// GOOD: protect with a mutex.
type SafeCounter struct {
	mu sync.Mutex
	n  int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func main() {
	const N = 1000
	bc := &BrokenCounter{}
	sc := &SafeCounter{}

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); bc.Inc() }()
		go func() { defer wg.Done(); sc.Inc() }()
	}
	wg.Wait()

	fmt.Println("broken counter (probably <", N, "):", bc.n)
	fmt.Println("safe counter (exactly", N, "):", sc.Value())
}
`,
		Notes: []string{
			"Run with `go test -race ./...` in CI. Period.",
			"The detector finds RUNTIME races — needs test coverage that actually exercises concurrent paths.",
			"sync/atomic is faster than a mutex for single-word counters/flags, but mutex is more readable for compound state.",
			"Common races: concurrent map writes (Go panics), shared slice append, closure captures from a for-range loop.",
		},
	},
}
