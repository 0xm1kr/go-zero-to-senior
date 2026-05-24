package lesson

// errorsLessons covers error wrapping with errors.Is/As and the sentinel-vs-typed-vs-opaque design decision.
var errorsLessons = []Lesson{
	{
		ID:       "errors",
		Category: "Errors",
		Title:    "Errors, Wrapping, errors.Is/As",
		Description: `
<p>Go has no exceptions. Functions that can fail return <code>(T, error)</code>.
The <code>error</code> interface is just <code>interface { Error() string }</code>.</p>
<p>Wrapping (<code>fmt.Errorf("...: %w", err)</code>) preserves the chain so callers
can use <code>errors.Is</code> (sentinel match) or <code>errors.As</code> (extract a typed
error).</p>
`,
		Code: `package main

import (
	"errors"
	"fmt"
)

// Sentinel error — compared with errors.Is.
var ErrNotFound = errors.New("not found")

// Typed error — extracted with errors.As.
type ValidationError struct{ Field string }

func (v *ValidationError) Error() string { return "invalid field: " + v.Field }

func loadUser(id int) error {
	if id < 0 {
		return &ValidationError{Field: "id"}
	}
	return fmt.Errorf("loadUser %d: %w", id, ErrNotFound)
}

func main() {
	err := loadUser(42)
	fmt.Println("raw:", err)

	if errors.Is(err, ErrNotFound) {
		fmt.Println("→ matched sentinel")
	}

	err2 := loadUser(-1)
	var ve *ValidationError
	if errors.As(err2, &ve) {
		fmt.Println("→ validation failed on:", ve.Field)
	}
}
`,
		Notes: []string{
			"Wrap with %w to keep the chain; %v / %s flatten it.",
			"Don't ignore errors. If you really must, _ = something() makes it explicit.",
			"Sentinel errors live in your package's public API: var ErrFoo = errors.New(...).",
		},
	},
	{
		ID:       "error-design",
		Category: "Errors",
		Title:    "Error Design Patterns (Senior-Level)",
		Description: `
<p>At interview level you'll be expected to articulate <b>why</b> a particular
error style fits a particular API. The three patterns:</p>

<ol>
  <li><b>Sentinel errors</b> — <code>var ErrNotFound = errors.New("not found")</code>.
    Compare with <code>errors.Is</code>. Best when there's a small, well-known set of
    failure modes callers will branch on. Examples: <code>io.EOF</code>, <code>sql.ErrNoRows</code>.</li>

  <li><b>Typed errors</b> — concrete struct with extra fields, extracted via
    <code>errors.As</code>. Best when the caller wants programmatic access to
    structured detail (which field failed validation, which HTTP status, which retry-after).</li>

  <li><b>Opaque errors</b> — just a string the caller logs. Best when the
    caller cannot meaningfully act on the failure. Most internal errors are
    (and should be) opaque.</li>
</ol>

<p><code>errors.Join</code> (Go 1.20+) lets you return multiple errors at once,
useful when validating many fields.</p>

<p>Coming from TypeScript: this is roughly the same decision tree as choosing
between throwing <code>new Error</code>, a custom Error subclass, or a tagged
union of <code>Result&lt;T, E&gt;</code> values.</p>
`,
		Code: `package main

import (
	"errors"
	"fmt"
)

// Sentinel — well-known, compared with errors.Is.
var ErrUnauthorized = errors.New("unauthorized")

// Typed — carries data, extracted with errors.As.
type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", v.Field, v.Message)
}

func validate(name, email string) error {
	var errs []error
	if name == "" {
		errs = append(errs, &ValidationError{Field: "name", Message: "required"})
	}
	if email == "" {
		errs = append(errs, &ValidationError{Field: "email", Message: "required"})
	}
	return errors.Join(errs...) // returns nil if errs is empty
}

func main() {
	err := validate("", "")
	if err == nil {
		fmt.Println("ok")
		return
	}
	fmt.Println("raw:", err)

	// Walk the joined errors with errors.As.
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Println("first validation field:", ve.Field)
	}

	// Sentinel check still works through wrapping/joining.
	wrapped := fmt.Errorf("loading user: %w", ErrUnauthorized)
	fmt.Println("is unauthorized?", errors.Is(wrapped, ErrUnauthorized))
}
`,
		Notes: []string{
			"Choose sentinel when callers branch on identity; typed when they need data; opaque otherwise.",
			"errors.Join (1.20+) replaces hand-rolled multi-error helpers.",
			"Don't expose internal package errors in your public API — wrap them.",
			"Interview tip: be ready to compare error handling here vs. Result types in Rust/TS.",
		},
	},
}
