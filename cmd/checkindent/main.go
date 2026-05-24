// checkindent runs gofmt on every lesson's Code field and reports any that
// differ from canonical formatting. Pass -fix to also rewrite the lesson
// source files in place. Pass -diff <id> to print a unified diff for one
// lesson.
//
// The patcher locates each Lesson composite literal in the source via AST,
// finds the Code key, and replaces its value expression byte-range with a
// freshly-encoded literal. This handles both simple raw-string literals
// and the `<...>` + "`" + `<...>` concatenation trick used to embed
// backticks (struct tags).
//
// Usage:
//
//	go run ./cmd/checkindent             # report only
//	go run ./cmd/checkindent -diff <id>  # show diff for one lesson
//	go run ./cmd/checkindent -fix        # rewrite lesson files in place
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang-tut/internal/lesson"
)

func main() {
	fix := flag.Bool("fix", false, "rewrite lesson source files in place")
	diffID := flag.String("diff", "", "show unified diff for a single lesson id and exit")
	flag.Parse()

	if *diffID != "" {
		runDiff(*diffID)
		return
	}

	jobs := findDirty()
	if len(jobs) == 0 {
		fmt.Println("ALL_LESSONS_GOFMT_CLEAN")
		return
	}
	if !*fix {
		fmt.Printf("\n%d lesson(s) need re-indentation. Re-run with -fix to apply.\n", len(jobs))
		return
	}
	applyFixes(jobs)
}

// runDiff prints a side-by-side diff for one lesson and exits.
func runDiff(id string) {
	for _, l := range lesson.Catalog {
		if l.ID != id {
			continue
		}
		formatted, err := format.Source([]byte(l.Code))
		if err != nil {
			fail(fmt.Errorf("parse error: %w", err))
		}
		printSideBySide(l.Code, string(formatted))
		return
	}
	fail(fmt.Errorf("lesson not found: %s", id))
}

// job pairs a lesson id with the gofmt-canonical version of its code that
// should overwrite the current value in the source.
type job struct{ id, formatted string }

// findDirty walks the catalog, formats every non-empty lesson via
// go/format, and returns the set whose canonical form differs from the
// source. Side-effect: prints a one-line summary per dirty lesson.
func findDirty() []job {
	var jobs []job
	for _, l := range lesson.Catalog {
		if strings.TrimSpace(l.Code) == "" {
			continue
		}
		formatted, err := format.Source([]byte(l.Code))
		if err != nil {
			fmt.Printf("[PARSE-ERR] %-30s  %v\n", l.ID, err)
			continue
		}
		if string(formatted) == l.Code {
			continue
		}
		fmt.Printf("[DIFF]      %-30s  (%d -> %d bytes)\n", l.ID, len(l.Code), len(formatted))
		jobs = append(jobs, job{id: l.ID, formatted: string(formatted)})
	}
	return jobs
}

// applyFixes patches each lessons_*.go file in place, replacing every
// affected lesson's Code expression with a freshly-encoded literal.
func applyFixes(jobs []job) {
	wantedByID := make(map[string]string, len(jobs))
	for _, j := range jobs {
		wantedByID[j.id] = j.formatted
	}

	matches, err := filepath.Glob("internal/lesson/lessons_*.go")
	if err != nil {
		fail(err)
	}
	for _, path := range matches {
		if err := patchFile(path, wantedByID); err != nil {
			fail(fmt.Errorf("%s: %w", path, err))
		}
	}
	for id := range wantedByID {
		fmt.Printf("[MISSED] %s (no matching Lesson literal found)\n", id)
	}
}

// fail prints an error to stderr and exits with status 1.
func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// patchFile parses one lessons_*.go file, finds each Lesson literal whose
// ID matches a key in wantedByID, and replaces the Code value's byte range
// with a freshly-encoded literal. Replacements are applied back-to-front
// so earlier offsets aren't invalidated.
func patchFile(path string, wantedByID map[string]string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return err
	}

	type edit struct {
		start, end int    // byte offsets in src
		replace    string // new literal expression
		id         string // lesson id (for logging)
	}
	var edits []edit

	ast.Inspect(f, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		id := lessonID(cl)
		if id == "" {
			return true
		}
		wantedCode, ok := wantedByID[id]
		if !ok {
			return true
		}
		val := codeValueExpr(cl)
		if val == nil {
			return true
		}
		start := fset.Position(val.Pos()).Offset
		end := fset.Position(val.End()).Offset
		edits = append(edits, edit{
			start:   start,
			end:     end,
			replace: encodeRawString(wantedCode),
			id:      id,
		})
		delete(wantedByID, id)
		return true
	})

	if len(edits) == 0 {
		return nil
	}

	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	out := append([]byte{}, src...)
	for _, e := range edits {
		var buf bytes.Buffer
		buf.Write(out[:e.start])
		buf.WriteString(e.replace)
		buf.Write(out[e.end:])
		out = buf.Bytes()
		fmt.Printf("[FIXED] %-25s  in %s\n", e.id, filepath.Base(path))
	}

	return os.WriteFile(path, out, 0o644)
}

// lessonID returns the ID literal of a Lesson composite literal, or "" if
// this isn't a Lesson literal.
func lessonID(cl *ast.CompositeLit) string {
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "ID" {
			continue
		}
		bl, ok := kv.Value.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return ""
		}
		s, err := strconv.Unquote(bl.Value)
		if err != nil {
			return ""
		}
		return s
	}
	return ""
}

// codeValueExpr returns the AST expression for the Code field of a Lesson.
func codeValueExpr(cl *ast.CompositeLit) ast.Expr {
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Code" {
			continue
		}
		return kv.Value
	}
	return nil
}

// encodeRawString turns an arbitrary string into a Go expression that
// evaluates back to that string.
//
// When the value contains no backticks (the common case) it returns the
// straightforward raw-string form: `<s>`.
//
// When the value contains one or more backticks, we can't use a single raw
// string (raw strings have no escape syntax), so we split on backticks and
// concatenate the parts with explicit "`" runs, e.g.:
//
//	`part1` + "`" + `part2`
//
// This mirrors the hand-written form already used by lessons that need to
// embed struct tags.
func encodeRawString(s string) string {
	if !strings.Contains(s, "`") {
		return "`" + s + "`"
	}
	parts := strings.Split(s, "`")

	// `"` + a literal backtick + `"`, written as an interpreted-string
	// literal "`". Stored as a constant for clarity.
	const literalBacktick = `"` + "`" + `"`

	var pieces []string
	for i, p := range parts {
		if i > 0 {
			pieces = append(pieces, literalBacktick)
		}
		if p != "" {
			pieces = append(pieces, "`"+p+"`")
		}
	}
	return strings.Join(pieces, " + ")
}

// printSideBySide renders the old vs new code with a "|" gutter, expanding
// tabs to 4 spaces so the eye can spot indentation drift.
func printSideBySide(oldCode, newCode string) {
	oldLines := strings.Split(oldCode, "\n")
	newLines := strings.Split(newCode, "\n")
	n := len(oldLines)
	if len(newLines) > n {
		n = len(newLines)
	}
	for i := 0; i < n; i++ {
		var l, r string
		if i < len(oldLines) {
			l = strings.ReplaceAll(oldLines[i], "\t", "    ")
		}
		if i < len(newLines) {
			r = strings.ReplaceAll(newLines[i], "\t", "    ")
		}
		marker := "  "
		if l != r {
			marker = "≠ "
		}
		fmt.Printf("%s%-50s | %s\n", marker, l, r)
	}
}
