package lesson

// genericsLessons covers type parameters and constraint patterns (any, comparable, cmp.Ordered, ~T).
var genericsLessons = []Lesson{
	{
		ID:       "generics-intro",
		Category: "Generics",
		Title:    "Generics: Type Parameters",
		Description: `
<p>Go 1.18 (March 2022) added generics. Syntax: <code>[T constraint]</code>
after a function or type name. Constraints ARE interfaces — a special form
called "type sets" that can list specific types with <code>|</code>.</p>

<p>From TypeScript, the shape is familiar (<code>function map&lt;T, U&gt;(...)</code>),
but constraints work differently: you declare them as interfaces, and Go
infers type params from arguments in most call sites.</p>

<p><b>When to use generics:</b> collections (Map/Filter/Reduce), containers
(Stack, Set, LRU), channels/streams, math helpers. <b>When NOT to:</b>
"just in case" abstraction. Plain interfaces are still idiomatic for
polymorphism.</p>
`,
		Code: `package main

import "fmt"

// Map applies fn to every element. "any" = no constraint, like TS "unknown".
func Map[T, U any](in []T, fn func(T) U) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

// Filter returns elements that satisfy pred.
func Filter[T any](in []T, pred func(T) bool) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

func main() {
	nums := []int{1, 2, 3, 4, 5}
	doubled := Map(nums, func(n int) int { return n * 2 })
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	lens := Map([]string{"go", "is", "fun"}, func(s string) int { return len(s) })

	fmt.Println("doubled:", doubled)
	fmt.Println("evens:", evens)
	fmt.Println("lens:", lens)
}
`,
		Notes: []string{
			"Type-param inference: Map(nums, fn) — Go figures out [int, int]. Explicit Map[int,int](...) is also legal.",
			"`any` is the alias for interface{} — added with generics in 1.18.",
			"Generics ≠ runtime reflection — they're compiled per call-site shape; no perf hit you'd care about.",
			"Don't generic everything. If T is always User, write ProcessUsers([]User) — clearer for readers.",
		},
	},
	{
		ID:       "generics-patterns",
		Category: "Generics",
		Title:    "Generic Constraints & Patterns",
		Description: `
<p>Go 1.21 shipped two production-grade generic packages worth memorizing:
<code>slices</code> and <code>maps</code>. Plus <code>cmp.Ordered</code> —
the "anything you can &lt; or &gt;" constraint.</p>

<h3>Constraints reference</h3>
<pre>any                 // anything (alias for interface{})
comparable          // anything you can == or !=
cmp.Ordered         // anything you can &lt;, &gt;, &lt;=, &gt;=  (1.21+)
~int | ~float64     // unions; ~ means "underlying type"</pre>

<p>The <code>~</code> operator lets you accept named types whose underlying
type matches: <code>type UserID int</code> still satisfies <code>~int</code>.</p>
`,
		Code: `package main

import (
	"cmp"
	"fmt"
	"slices"
)

// Set is a generic set built on map. Keys must be comparable.
type Set[T comparable] struct{ m map[T]struct{} }

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(items))}
	for _, v := range items {
		s.m[v] = struct{}{}
	}
	return s
}

func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int     { return len(s.m) }

// Max works on anything Ordered: ints, floats, strings, custom Ordered types.
func Max[T cmp.Ordered](xs []T) T {
	if len(xs) == 0 {
		panic("Max: empty slice")
	}
	best := xs[0]
	for _, x := range xs[1:] {
		if x > best {
			best = x
		}
	}
	return best
}

func main() {
	s := NewSet("go", "rust", "ts", "go")
	fmt.Println("len:", s.Len(), "has rust:", s.Has("rust"), "has zig:", s.Has("zig"))

	fmt.Println(Max([]int{3, 1, 4, 1, 5, 9, 2, 6}))
	fmt.Println(Max([]string{"banana", "apple", "cherry"}))

	// slices/maps packages give you sorted, contains, equal, clone, etc.
	nums := []int{3, 1, 2}
	slices.Sort(nums)
	fmt.Println(nums, "contains 2?", slices.Contains(nums, 2))
}
`,
		Notes: []string{
			"Memorize: any, comparable, cmp.Ordered, and ~T (underlying-type matching).",
			"The slices and maps packages (1.21+) make most hand-rolled helpers obsolete.",
			"A generic struct's methods must repeat the type param: func (s *Set[T]) Foo() ...",
			"For method-bearing constraints, declare a normal interface: type Hashable interface { Hash() uint64 }",
		},
	},
}
