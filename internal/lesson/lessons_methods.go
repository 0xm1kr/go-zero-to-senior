package lesson

// methodsLessons covers methods (value vs pointer receivers), interfaces, embedding, and type assertions/switches.
var methodsLessons = []Lesson{
	{
		ID:       "methods",
		Category: "Methods & Interfaces",
		Title:    "Methods (value vs pointer receivers)",
		Description: `
<p>Go has no classes. You attach methods to any named type with a
<i>receiver</i>. Choose receiver type carefully:</p>
<ul>
  <li><b>Value receiver</b> <code>func (u User)</code> – method sees a COPY.</li>
  <li><b>Pointer receiver</b> <code>func (u *User)</code> – can mutate the original; cheaper for big structs.</li>
</ul>
<p>Rule of thumb: be consistent across all methods of a type, and prefer
pointer receivers when in doubt.</p>
`,
		Code: `package main

import "fmt"

type Counter struct{ n int }

func (c Counter) Get() int   { return c.n }
func (c *Counter) Inc()      { c.n++ }
func (c *Counter) Add(d int) { c.n += d }

// You can attach methods to non-struct types too.
type Celsius float64

func (c Celsius) Fahrenheit() float64 { return float64(c)*9/5 + 32 }

func main() {
	c := Counter{}
	c.Inc()
	c.Inc()
	c.Add(5)
	fmt.Println("counter:", c.Get())

	t := Celsius(100)
	fmt.Println("100°C =", t.Fahrenheit(), "°F")
}
`,
		Notes: []string{
			"Go automatically takes the address for pointer-receiver calls on addressable values.",
			"Methods on built-ins like int aren't allowed — but methods on `type Money int` are.",
			"There's no `this` or `self`; you name the receiver explicitly.",
		},
	},
	{
		ID:       "interfaces",
		Category: "Methods & Interfaces",
		Title:    "Interfaces (Implicit Satisfaction)",
		Description: `
<p>An interface is a set of method signatures. Crucially, <b>nothing declares
that a type implements an interface</b> – if the methods exist, it satisfies
the interface (structural / duck typing).</p>
<p>This makes Go interfaces feel like protocols: you can write tiny
interfaces (often just one method) wherever you need them, including in your
own package, without touching the third-party type that implements them.</p>
`,
		Code: `package main

import "fmt"

type Greeter interface {
	Greet() string
}

type English struct{ Name string }
type Spanish struct{ Name string }

func (e English) Greet() string { return "Hello, " + e.Name }
func (s Spanish) Greet() string { return "Hola, " + s.Name }

func sayAll(gs ...Greeter) {
	for _, g := range gs {
		fmt.Println(g.Greet())
	}
}

func main() {
	sayAll(English{"Alice"}, Spanish{"Carlos"})

	// Type assertion: extract concrete type.
	var g Greeter = English{"Bob"}
	if e, ok := g.(English); ok {
		fmt.Println("english greeter, name =", e.Name)
	}
}
`,
		Notes: []string{
			"Tiny interfaces are idiomatic: io.Reader, io.Writer, fmt.Stringer all have ONE method.",
			"The empty interface `interface{}` (or `any` in Go 1.18+) means \"any value\".",
			"Declare interfaces where they're USED, not where they're implemented.",
		},
	},
	{
		ID:       "embedding",
		Category: "Methods & Interfaces",
		Title:    "Embedding & Composition",
		Description: `
<p>Go has no inheritance, but it has <b>embedding</b>: include a type inside
another and its methods are promoted. This achieves code reuse without
classical inheritance pitfalls.</p>
`,
		Code: `package main

import "fmt"

type Animal struct{ Name string }

func (a Animal) Describe() string { return "I am " + a.Name }

type Dog struct {
	Animal // embedded — Dog gets Animal's fields and methods
	Breed  string
}

func (d Dog) Bark() string { return d.Name + " says woof" }

func main() {
	d := Dog{
		Animal: Animal{Name: "Rex"},
		Breed:  "Husky",
	}
	fmt.Println(d.Describe()) // promoted from Animal
	fmt.Println(d.Bark())
	fmt.Println("breed:", d.Breed, "name:", d.Name)
}
`,
		Notes: []string{
			"Embedding works for interfaces too — bigger interfaces can embed smaller ones.",
			"It's composition, not inheritance: methods are forwarded, not overridden polymorphically.",
			"You can shadow embedded methods by defining one on the outer type.",
		},
	},
	{
		ID:       "type-assertions",
		Category: "Methods & Interfaces",
		Title:    "Type Assertions & Type Switches",
		Description: `
<p>Once you have a value as an interface (especially <code>any</code>),
you need to extract its concrete type. Two mechanisms: type assertion and
type switch.</p>

<p>From TypeScript, type assertions look superficially like
<code>value as Type</code> — but Go's are <b>checked at runtime</b> and
can fail. TS's <code>as</code> is purely compile-time erasure.</p>

<ul>
  <li><code>v.(T)</code> — single-value form. Panics if v isn't a T.</li>
  <li><code>v.(T) ok</code> — two-value form. Never panics; ok is false if mismatch.</li>
  <li><code>switch x := v.(type) { case T: ... }</code> — type switch.</li>
</ul>
`,
		Code: `package main

import "fmt"

type Stringer interface{ String() string }

type Point struct{ X, Y int }

func (p Point) String() string { return fmt.Sprintf("(%d, %d)", p.X, p.Y) }

func describe(v any) {
	// Two-value form: never panics.
	if s, ok := v.(Stringer); ok {
		fmt.Println("stringable:", s.String())
		return
	}

	// Type switch: branch on dynamic concrete type.
	switch x := v.(type) {
	case int:
		fmt.Println("int doubled:", x*2)
	case string:
		fmt.Println("string len:", len(x))
	case nil:
		fmt.Println("nil value")
	default:
		fmt.Printf("unknown type %T\n", x)
	}
}

func main() {
	describe(Point{3, 4})
	describe(42)
	describe("hello")
	describe(3.14)
	describe(nil)
}
`,
		Notes: []string{
			"v.(T) panics if v isn't a T — always prefer the two-value form unless you've already proven the type.",
			"A type switch is many assertions in one — much cleaner than chained ifs.",
			"%T in Printf prints a value's dynamic type — invaluable for debugging interfaces.",
			"Coming from TS: `as` is erased at runtime; Go's assertion is a real check.",
		},
	},
}
