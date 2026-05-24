package lesson

// dataStructuresLessons covers arrays/slices, maps, pointers, and structs.
var dataStructuresLessons = []Lesson{
	{
		ID:       "arrays-slices",
		Category: "Data Structures",
		Title:    "Arrays & Slices",
		Description: `
<p><b>Arrays</b> have a fixed size in their type: <code>[5]int</code> and <code>[6]int</code>
are different types. You'll rarely use them directly.</p>
<p><b>Slices</b> are the workhorse – a dynamic view into an underlying array. A
slice has three parts: pointer, length, capacity. <code>make([]T, len, cap)</code>
allocates one; <code>append</code> grows it (returning a new header).</p>
`,
		Code: `package main

import "fmt"

func main() {
	// Array — fixed size.
	var a [3]int = [3]int{10, 20, 30}
	fmt.Println(a, len(a))

	// Slice literal.
	s := []int{1, 2, 3}
	s = append(s, 4, 5)
	fmt.Println(s, "len:", len(s), "cap:", cap(s))

	// Slicing: s[low:high] — no copy, shares memory.
	mid := s[1:4]
	fmt.Println("mid:", mid)
	mid[0] = 999
	fmt.Println("s after mid mutation:", s) // s also changed!

	// make: pre-allocate to avoid re-allocations.
	buf := make([]byte, 0, 64)
	buf = append(buf, "hello"...)
	fmt.Println(string(buf))
}
`,
		Notes: []string{
			"Slices share their backing array — be careful when re-slicing.",
			"append may return a NEW slice header. Always assign: s = append(s, x).",
			"Pre-size with make([]T, 0, n) when you know the upper bound — much faster.",
		},
	},
	{
		ID:       "maps",
		Category: "Data Structures",
		Title:    "Maps",
		Description: `
<p>Maps are hash tables: <code>map[K]V</code>. Keys must be comparable. The zero value
is <code>nil</code>; you can read from a nil map but writing panics. Always
<code>make</code> one before assigning.</p>
<p>The "comma-ok" idiom tells you whether a key was present vs. zero-valued.</p>
`,
		Code: `package main

import "fmt"

func main() {
	scores := map[string]int{
		"alice": 90,
		"bob":   75,
	}
	scores["carol"] = 88
	delete(scores, "bob")

	// comma-ok
	if v, ok := scores["dave"]; ok {
		fmt.Println("dave:", v)
	} else {
		fmt.Println("no dave")
	}

	for name, score := range scores {
		fmt.Printf("%s=%d ", name, score)
	}
	fmt.Println()
}
`,
		Notes: []string{
			"Map iteration order is RANDOMIZED on purpose — don't rely on it.",
			"For sorted iteration, collect keys into a slice and sort.Strings(keys).",
			"Maps are reference-like: passing one to a function shares the same map.",
		},
	},
	{
		ID:       "pointers",
		Category: "Data Structures",
		Title:    "Pointers",
		Description: `
<p>A pointer holds the memory address of a value. <code>&amp;x</code> takes the
address, <code>*p</code> dereferences. Go has pointers but <b>no pointer arithmetic</b>
– safer than C.</p>
<p>Use pointers to (1) mutate the caller's value, (2) avoid copying big
structs, or (3) signal an optional field.</p>
`,
		Code: `package main

import "fmt"

type User struct{ Name string }

func renameByValue(u User)    { u.Name = "Changed" }
func renameByPointer(u *User) { u.Name = "Changed" }

func main() {
	u := User{Name: "Alice"}

	renameByValue(u)
	fmt.Println("after value:", u.Name) // Alice

	renameByPointer(&u)
	fmt.Println("after ptr:", u.Name) // Changed

	// new() returns a pointer to a zero value.
	p := new(int) // *int → 0
	*p = 42
	fmt.Println(*p)
}
`,
		Notes: []string{
			"No null-pointer crashes if you check: if p != nil { ... }.",
			"Go's GC manages memory; you don't free pointers manually.",
			"Pointer receivers in methods are the usual way to mutate structs — see Methods lesson.",
		},
	},
	{
		ID:       "structs",
		Category: "Data Structures",
		Title:    "Structs",
		Description: `
<p>Structs group fields. They're values, not references. Tags (the backticks
after a field) are metadata used by libraries like encoding/json.</p>
`,
		Code: `package main

import "fmt"

type Address struct {
	Street, City string
}

type User struct {
	ID      int
	Name    string
	Email   string ` + "`" + `json:"email"` + "`" + `
	Address Address
}

func main() {
	u := User{
		ID:    1,
		Name:  "Alice",
		Email: "a@example.com",
		Address: Address{
			Street: "1 Main",
			City:   "Gopherville",
		},
	}
	fmt.Printf("%+v\n", u) // includes field names
	fmt.Println("city:", u.Address.City)

	// Anonymous struct — handy for one-off configs.
	cfg := struct {
		Port int
		Host string
	}{Port: 8080, Host: "0.0.0.0"}
	fmt.Println(cfg)
}
`,
		Notes: []string{
			"Use %+v in Printf to see field names while debugging.",
			"Struct tags drive JSON, DB, validation libraries — they're regular strings.",
			"Zero value of a struct is a struct with every field at its zero value.",
		},
	},
}
