package lesson

// basicsLessons introduces Go fundamentals: declarations, zero values, types, constants, plus a TS→Go mental model for engineers arriving from JS/TS.
var basicsLessons = []Lesson{
	{
		ID:       "intro",
		Category: "Basics",
		Title:    "Welcome & Setup",
		Description: `
<p>Welcome! This app teaches Go through small, runnable examples. Every lesson
has an editor on the right – modify the code, click <b>Run</b>, and the
server compiles and executes it with <code>go run</code>.</p>

<h3>Why Go?</h3>
<ul>
  <li><b>Fast to learn</b> – ~25 keywords, no inheritance, no generics-soup. The spec fits in an evening.</li>
  <li><b>Fast to build</b> – single static binary, sub-second compile, no runtime to ship.</li>
  <li><b>Built for servers</b> – goroutines and channels make concurrency first-class.</li>
  <li><b>Strong standard library</b> – production HTTP, JSON, crypto, SQL, templating out of the box.</li>
</ul>

<h3>Installing Go</h3>
<pre>brew install go         # macOS
sudo apt install golang # Debian/Ubuntu
# or download from https://go.dev/dl/</pre>

<h3>Project anatomy</h3>
<pre>myapp/
├── go.mod          # module name + dependencies (like package.json)
├── go.sum          # checksums for reproducible builds
├── main.go         # entry point of the &quot;main&quot; package
└── internal/...    # private packages (Go enforces this directory name)</pre>

<p>Hit <b>Run</b> below to confirm the runner works.</p>
`,
		Code: `package main

import "fmt"

func main() {
	fmt.Println("Hello, Gopher!")
}
`,
		Notes: []string{
			"Every executable Go program starts in package main with a main() function.",
			"fmt is the formatted I/O package — used constantly.",
			"Imports are explicit; unused imports are a compile error.",
		},
	},
	{
		ID:       "ts-vs-go",
		Category: "Basics",
		Title:    "Coming from TypeScript: Mental Model",
		Description: `
<p>If you're fluent in TypeScript/Node, here's the translation table that
saves the most time. Most of Go's choices feel restrictive at first, then
liberating once your brain rewires.</p>

<table style="width:100%;border-collapse:collapse;font-size:13px;margin-top:6px">
<tr style="background:#1d2230;text-align:left">
<th style="padding:6px">TypeScript</th><th>Go</th><th>Note</th></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>let / const</code></td>
<td style="border-top:1px solid #2a3040"><code>var</code>, <code>:=</code>, <code>const</code></td>
<td style="border-top:1px solid #2a3040">No "let" — Go has zero values, no "undefined" state.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>null | undefined</code></td>
<td style="border-top:1px solid #2a3040"><code>nil</code></td>
<td style="border-top:1px solid #2a3040">Only one "nothing"; ints default to 0, strings to "".</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>interface Foo {}</code></td>
<td style="border-top:1px solid #2a3040"><code>type Foo interface {}</code></td>
<td style="border-top:1px solid #2a3040">Structural — like TS — but satisfied implicitly. No <code>implements</code>.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>class Foo {}</code></td>
<td style="border-top:1px solid #2a3040"><code>type Foo struct {}</code> + methods</td>
<td style="border-top:1px solid #2a3040">No classes, no <code>new</code>, no inheritance. Composition via embedding.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>throw new Error(...)</code></td>
<td style="border-top:1px solid #2a3040"><code>return ..., err</code></td>
<td style="border-top:1px solid #2a3040">Errors are values. Reserve <code>panic</code> for unrecoverable bugs.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>async / await / Promise</code></td>
<td style="border-top:1px solid #2a3040"><code>go func(){}</code> + channels</td>
<td style="border-top:1px solid #2a3040">No event loop. Real OS-thread parallelism, runtime-scheduled.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>any</code></td>
<td style="border-top:1px solid #2a3040"><code>any</code> (alias for <code>interface{}</code>)</td>
<td style="border-top:1px solid #2a3040">Since Go 1.18.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>Array&lt;T&gt;</code></td>
<td style="border-top:1px solid #2a3040"><code>[]T</code> (slice) or <code>[N]T</code> (array)</td>
<td style="border-top:1px solid #2a3040">Slices ≠ arrays. Almost always slices.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>Record / Map</code></td>
<td style="border-top:1px solid #2a3040"><code>map[K]V</code></td>
<td style="border-top:1px solid #2a3040">Iteration order is RANDOMIZED. Yes, really.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>npm / package.json</code></td>
<td style="border-top:1px solid #2a3040"><code>go mod / go.mod</code></td>
<td style="border-top:1px solid #2a3040">Built into the toolchain. No <code>node_modules</code>.</td></tr>
<tr><td style="padding:6px;border-top:1px solid #2a3040"><code>tsc / prettier / eslint</code></td>
<td style="border-top:1px solid #2a3040">One <code>go</code> binary: <code>build</code>, <code>fmt</code>, <code>vet</code></td>
<td style="border-top:1px solid #2a3040">Single canonical style. Debates extinguished.</td></tr>
</table>

<h3>Three things TS devs take longest to internalize</h3>
<ol>
  <li><b>Pointer vs value semantics.</b> In TS, objects are reference and primitives are value. In Go, you <i>choose</i> per type. Mis-choosing causes the most "why didn't my mutation stick?" bugs.</li>
  <li><b>The (val, err) return convention.</b> Your "if (err) return" muscle memory from Node callbacks is back. There's no exception bubbling.</li>
  <li><b>Capitalization = visibility.</b> <code>Name</code> is exported, <code>name</code> is package-private. There is no <code>public</code> keyword anywhere.</li>
</ol>
`,
		Code: `package main

import "fmt"

// In TS: interface User { id: number; name: string; pwd?: string }
// In Go:
type User struct {
	ID   int // capital I = exported (public)
	Name string
	pwd  string // lowercase = package-private
}

// In TS: function greet(u: User): string { return "Hi " + u.name; }
// In Go:
func greet(u User) string {
	return "Hi " + u.Name
}

func main() {
	users := []User{
		{ID: 1, Name: "Alice", pwd: "secret"},
		{ID: 2, Name: "Bob"},
	}
	for _, u := range users {
		fmt.Println(greet(u))
	}
}
`,
		Notes: []string{
			"Interfaces are structural (like TS) but satisfied implicitly — no `implements` keyword.",
			"There's only `nil`, not null vs undefined. Every type has a deterministic zero value.",
			"Async work is `go func()` + channels — no Promise, no async/await keywords needed.",
			"gofmt is canonical: one style across all Go code on Earth. Stop arguing about tabs.",
		},
	},
	{
		ID:       "variables",
		Category: "Basics",
		Title:    "Variables & Zero Values",
		Description: `
<p>Go has three ways to declare variables. Pick whichever reads cleanest.</p>
<ul>
  <li><code>var x int</code>             — declare, zero value (<code>0</code>).</li>
  <li><code>var x int = 5</code>         — declare with type and value.</li>
  <li><code>var x = 5</code>             — type inferred.</li>
  <li><code>x := 5</code>                — short declaration (function bodies only).</li>
</ul>
<p>Every type has a <b>zero value</b>: <code>0</code> for numbers, <code>""</code> for
strings, <code>false</code> for bools, <code>nil</code> for pointers/slices/maps/channels/functions/interfaces.
There is no "uninitialized" state. This eliminates a huge class of bugs.</p>
`,
		Code: `package main

import "fmt"

func main() {
	var a int // zero value
	var b int = 42
	var c = "inferred"
	d := 3.14 // short declaration

	fmt.Println(a, b, c, d)

	// Multiple assignment / swap
	x, y := 1, 2
	x, y = y, x
	fmt.Println(x, y)
}
`,
		Notes: []string{
			":= can only appear inside a function. At package scope use var.",
			"Unused local variables are a compile error — Go forces clean code.",
			"Use _ to discard a value you don't need: x, _ := something().",
		},
	},
	{
		ID:       "constants",
		Category: "Basics",
		Title:    "Constants & iota",
		Description: `
<p>Constants are evaluated at compile time. They can be typed or untyped;
untyped constants are extremely flexible and convert implicitly.</p>
<p><code>iota</code> is a counter you can use inside a <code>const</code> block. It
starts at 0 and increments by 1 per line. Perfect for enums.</p>
`,
		Code: `package main

import "fmt"

const Pi = 3.14159
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Enum with iota.
type Weekday int

const (
	Sunday    Weekday = iota // 0
	Monday                   // 1
	Tuesday                  // 2
	Wednesday                // 3
)

// Bit flags with iota.
const (
	Read    = 1 << iota // 1
	Write               // 2
	Execute             // 4
)

func main() {
	fmt.Println(Pi, StatusActive)
	fmt.Println(Sunday, Monday, Tuesday, Wednesday)
	fmt.Println("rwx =", Read|Write|Execute)
}
`,
		Notes: []string{
			"Constants must be known at compile time — no function calls.",
			"iota resets to 0 at every const block.",
			"Bit-shifting with iota gives you compact, idiomatic flag enums.",
		},
	},
	{
		ID:       "types",
		Category: "Basics",
		Title:    "Basic Types & Conversions",
		Description: `
<p>Go's built-in types:</p>
<ul>
  <li><b>Integers:</b> <code>int</code>, <code>int8/16/32/64</code>, <code>uint8/16/32/64</code>, <code>byte</code> (alias for uint8), <code>rune</code> (alias for int32, holds a Unicode code point).</li>
  <li><b>Floats:</b> <code>float32</code>, <code>float64</code>.</li>
  <li><b>Other:</b> <code>bool</code>, <code>string</code>, <code>complex64/128</code>.</li>
</ul>
<p>Go is <b>strict</b> about types: there is no implicit numeric conversion. You
must say <code>float64(x)</code> explicitly. This catches a huge amount of bugs
that C/JS hide.</p>
`,
		Code: `package main

import "fmt"

func main() {
	var i int = 42
	var f float64 = float64(i) // explicit conversion required
	var u uint = uint(f)
	fmt.Println(i, f, u)

	// Strings <-> bytes
	s := "héllo"
	b := []byte(s)              // bytes (UTF-8 encoded)
	r := []rune(s)              // runes (code points)
	fmt.Println(len(b), len(r)) // bytes != runes for non-ASCII

	// String concat
	name := "Gopher"
	fmt.Println("Hi, " + name)
}
`,
		Notes: []string{
			"Use int unless you have a reason — it's platform-sized (32 or 64-bit).",
			"len(string) returns BYTES, not characters. Use []rune for code points.",
			"strconv (not casts) converts strings ↔ numbers: strconv.Atoi, strconv.Itoa.",
		},
	},
}
