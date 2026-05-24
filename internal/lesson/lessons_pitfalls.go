package lesson

// pitfallsLessons covers the typed-nil trap and loop-variable / slice-aliasing footguns.
var pitfallsLessons = []Lesson{
	{
		ID:       "typed-nil",
		Category: "Senior Pitfalls",
		Title:    "The Typed Nil Trap",
		Description: `
<p>One of the most-Googled Go gotchas and a classic interview question.
An interface variable in Go is a TWO-WORD value: <code>(type, value)</code>.
A nil interface is <code>(nil, nil)</code>. But an interface holding a
nil concrete pointer is <code>(*MyErr, nil)</code> — which is NOT nil
when compared with <code>== nil</code>.</p>

<p>The classic bug: a function returns <code>error</code>, you build
the return through a <code>*MyError</code>, set it to nil "for no
error," and the caller's <code>if err != nil</code> still fires.</p>
`,
		Code: `package main

import "fmt"

type MyError struct{ Msg string }

func (e *MyError) Error() string { return e.Msg }

// BUG: returns an interface whose dynamic type is *MyError, value nil.
// Caller's "if err != nil" evaluates TRUE.
func broken() error {
	var e *MyError
	return e
}

// GOOD: return nil literally, or return a non-nil concrete error.
func good() error {
	return nil
}

func main() {
	err := broken()
	fmt.Println("broken err == nil?", err == nil) // false — surprise!
	fmt.Printf("  underlying: type=%T value=%v\n", err, err)

	err = good()
	fmt.Println("good err == nil?", err == nil) // true
}
`,
		Notes: []string{
			"An interface is nil only when BOTH its type and value are nil.",
			"Rule: never declare `var x *T` and return it as an `error`. Return nil literally.",
			"go vet's `nilness` analyzer catches some cases. Run it in CI.",
			"Same trap applies to ANY interface: io.Reader, http.ResponseWriter, etc.",
		},
	},
	{
		ID:       "loop-variable",
		Category: "Senior Pitfalls",
		Title:    "Loop Variable Capture & Slice Aliasing",
		Description: `
<p>Two related foot-guns every Go interviewer asks about:</p>

<h3>1. Loop variable capture (pre-Go 1.22)</h3>
<p>Before Go 1.22 the loop variable in <code>for i, v := range ...</code>
was a SINGLE variable rebound each iteration. Capturing it in a goroutine
or closure that ran later saw the LAST value.</p>

<p>Go 1.22 gave each iteration its own variable. <code>go vet</code>
warns on the old pattern.</p>

<h3>2. Slice aliasing</h3>
<p><code>s[:5]</code> shares the SAME backing array as <code>s</code>.
Appending to either can stomp on the other. Senior Go code uses
<code>slices.Clone</code> (or a manual copy) when a sub-slice escapes
the caller's lifetime.</p>
`,
		Code: `package main

import (
	"fmt"
	"sync"
)

func main() {
	// === Loop capture ===
	var wg sync.WaitGroup
	for _, name := range []string{"alice", "bob", "carol"} {
		// In Go 1.22+, this just works. In 1.21 and earlier, you needed:
		//     name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("hi", name)
		}()
	}
	wg.Wait()

	// === Slice aliasing ===
	base := []int{1, 2, 3, 4, 5}
	sub := base[:3] // shares backing array with base

	sub = append(sub, 99)      // overwrites base[3]!
	fmt.Println("base:", base) // [1 2 3 99 5] — surprise
	fmt.Println("sub: ", sub)  // [1 2 3 99]

	// Defensive copy when sub-slice outlives caller:
	safe := append([]int(nil), base[:3]...)
	safe = append(safe, 99)
	fmt.Println("safe:", safe, "base unchanged:", base[:5])
}
`,
		Notes: []string{
			"Go 1.22 fixed for-range loop variable scoping. Targeting older Go? Write `x := x` inside the loop.",
			"s[low:high:max] (three-index slicing) caps capacity so subsequent appends MUST realloc — defends against aliasing.",
			"Use slices.Clone(s) (Go 1.21+) when you need an owned copy.",
			"`go vet -copylocks -loopclosure` catches both these classes. Wire it into CI.",
		},
	},
}
