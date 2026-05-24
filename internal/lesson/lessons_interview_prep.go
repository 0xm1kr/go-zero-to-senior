package lesson

// interviewPrepLessons is the closing cheatsheet of senior-Go interview signals.
var interviewPrepLessons = []Lesson{
	{
		ID:       "interview-prep",
		Category: "Interview Prep",
		Title:    "Senior Go Interview Cheatsheet",
		Description: `
<p>You've made it. Condensed prep list to keep in your head the day of:</p>

<h3>Things they'll ask you to articulate</h3>
<ul>
  <li><b>Goroutines vs OS threads</b> — m:n scheduling, ~few KB stack, cheap.</li>
  <li><b>Channel vs mutex — when each?</b> Channel for ownership transfer, mutex for shared state.</li>
  <li><b>Slice internals</b> — ptr/len/cap; append behavior; aliasing.</li>
  <li><b>Map internals</b> — hash table, NOT thread-safe, iteration order randomized.</li>
  <li><b>Interface internals</b> — (type, value) pair; nil-interface trap.</li>
  <li><b>Context philosophy</b> — first param, propagates deadline + cancellation.</li>
  <li><b>Error handling philosophy</b> — values not exceptions; wrap with %w; sentinel vs typed vs opaque.</li>
  <li><b>GC overview</b> — concurrent tri-color mark-sweep; tune with GOGC, GOMEMLIMIT (1.19+).</li>
</ul>

<h3>Things they'll ask you to write</h3>
<ul>
  <li>Worker pool (fan-out/fan-in).</li>
  <li>Rate limiter (token bucket).</li>
  <li>LRU cache.</li>
  <li>Concurrent counter / map (or "why would you use sync.Map?").</li>
  <li>Producer/consumer with graceful shutdown.</li>
  <li>One algorithm — typically two-pointers, BFS, or DP-easy.</li>
</ul>

<h3>Phrases that signal seniority</h3>
<ul>
  <li>"I'd run this with <code>-race</code> in CI."</li>
  <li>"I'd benchmark with <code>-benchmem</code> to check allocations."</li>
  <li>"This needs a context deadline so callers can cancel."</li>
  <li>"Typed errors at the API boundary, opaque internally."</li>
  <li>"Let me make sure this server has timeouts and graceful shutdown."</li>
</ul>

<h3>Final-skim docs</h3>
<ul>
  <li><i>Effective Go</i> — the official idiom guide.</li>
  <li>The <a href="https://go.dev/ref/spec">Go spec</a> sections on slices, interfaces, channels.</li>
  <li><a href="https://github.com/uber-go/guide/blob/master/style.md">Uber's Go style guide</a> — interview-grade idiom.</li>
  <li><a href="https://google.github.io/styleguide/go/">Google's Go style guide</a> — what your reviewer reads.</li>
</ul>

<p>Good luck.</p>
`,
		Code: `package main

import "fmt"

// One-page senior-Go reference: idioms that ALWAYS show up.

func main() {
	// 1. Tag-only struct = zero-cost set member.
	set := map[string]struct{}{"alice": {}, "bob": {}}
	_, exists := set["alice"]
	fmt.Println("set has alice?", exists)

	// 2. Comma-ok across map, channel, assertion, type switch.
	if v, ok := set["bob"]; ok {
		fmt.Printf("bob -> %v\n", v)
	}

	// 3. Slice tricks every senior should know.
	nums := []int{1, 2, 3, 4, 5}
	nums = append(nums[:2], nums[3:]...)           // delete index 2  -> [1 2 4 5]
	nums = append([]int{0}, nums...)               // prepend 0       -> [0 1 2 4 5]
	a, b := nums[:len(nums)/2], nums[len(nums)/2:] // split
	fmt.Println("nums:", nums, "halves:", a, b)

	// 4. Channel ownership: producer closes.
	ch := make(chan int, 3)
	go func() { defer close(ch); ch <- 1; ch <- 2; ch <- 3 }()
	for v := range ch {
		fmt.Println("recv", v)
	}

	//  5. errgroup-shaped pattern: WaitGroup + first-error.
	//     In production use golang.org/x/sync/errgroup directly.
}
`,
		Notes: []string{
			"Walk in able to draw the worker-pool pattern on a whiteboard without thinking.",
			"Saying \"I'd add a metric/log/trace here\" marks you as senior.",
			"Explaining WHY mutex vs channel (or vice versa) is the single highest-signal answer.",
			"Read your own code aloud during the interview — Go forces honesty; you'll catch your own bugs.",
		},
	},
}
