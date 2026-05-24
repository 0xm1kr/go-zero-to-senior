package lesson

// stdlibLessons covers io, encoding/json, time, and log/slog.
var stdlibLessons = []Lesson{
	{
		ID:       "io",
		Category: "Standard Library",
		Title:    "Files & io.Reader/Writer",
		Description: `
<p>The <code>io.Reader</code> and <code>io.Writer</code> interfaces are everywhere – files,
HTTP bodies, gzip, encryption, network connections. Learn them once, reuse
them with everything.</p>
`,
		Code: `package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	// Write to a temp file.
	f, err := os.CreateTemp("", "demo-*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())

	io.Copy(f, strings.NewReader("hello from io.Copy\n"))
	f.Close()

	// Read it back.
	data, _ := os.ReadFile(f.Name())
	fmt.Printf("file %q contains: %s", f.Name(), data)
}
`,
		Notes: []string{
			"os.ReadFile / os.WriteFile cover the common cases in one call.",
			"io.Copy(dst, src) is the universal pipe — use it everywhere.",
			"bufio.Scanner is ideal for line-by-line reading; check scanner.Err() at the end.",
		},
	},
	{
		ID:       "json",
		Category: "Standard Library",
		Title:    "encoding/json",
		Description: `
<p>JSON is encoded/decoded with struct tags. <code>json.Marshal</code> produces
bytes; <code>json.Unmarshal</code> parses them. Streaming versions exist via
<code>json.Encoder</code>/<code>Decoder</code> for HTTP bodies and files.</p>
`,
		Code: `package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID    int    ` + "`" + `json:"id"` + "`" + `
	Name  string ` + "`" + `json:"name"` + "`" + `
	Email string ` + "`" + `json:"email,omitempty"` + "`" + ` // omit if empty
	pwd   string // unexported = never serialized
}

func main() {
	u := User{ID: 1, Name: "Alice", pwd: "secret"}

	bytes, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(bytes))

	// Decode.
	input := []byte(` + "`" + `{"id":2,"name":"Bob","email":"b@x.com"}` + "`" + `)
	var u2 User
	if err := json.Unmarshal(input, &u2); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", u2)
}
`,
		Notes: []string{
			"omitempty: skip field if it's the zero value.",
			"For arbitrary JSON without a struct, decode into map[string]interface{}.",
			"Use json.Decoder for streaming (HTTP bodies, big files) — avoids buffering.",
		},
	},
	{
		ID:       "time",
		Category: "Standard Library",
		Title:    "time: Durations & Layouts",
		Description: `
<p>The <code>time</code> package is great but quirky in one place: format
strings. Instead of strftime tokens, Go uses the reference date
<code>Mon Jan 2 15:04:05 MST 2006</code> – memorize it as <b>01/02 03:04:05PM '06 -0700</b>.</p>
`,
		Code: `package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println("RFC3339:", now.Format(time.RFC3339))
	fmt.Println("custom :", now.Format("2006-01-02 15:04:05"))

	// Parse.
	t, _ := time.Parse("2006-01-02", "2025-05-24")
	fmt.Println("parsed:", t)

	// Durations.
	d := 90 * time.Minute
	fmt.Println("90m =", d, "=", d.Hours(), "hours")

	// Timer.
	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	fmt.Println("slept:", time.Since(start))
}
`,
		Notes: []string{
			"time.Duration is int64 nanoseconds underneath — math just works.",
			"Always store/transmit times in UTC; convert to local only for display.",
			"For sleep with cancellation, prefer time.After in a select over time.Sleep.",
		},
	},
	{
		ID:       "slog",
		Category: "Standard Library",
		Title:    "log/slog: Structured Logging",
		Description: `
<p>Go 1.21 added <code>log/slog</code> to the standard library —
structured, leveled, contextual logging. Before 1.21, every team picked
between <code>zap</code>, <code>logrus</code>, <code>zerolog</code>, etc.
slog is now the lingua franca; you'll find it in modern codebases and
job interviews alike.</p>

<p>Coming from Node? Think of it as <code>pino</code> baked into the
standard library: structured JSON by default, child loggers, leveled,
and a single canonical API.</p>
`,
		Code: `package main

import (
	"errors"
	"log/slog"
	"os"
)

func main() {
	// JSON handler is production-ready. TextHandler is friendlier for dev.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("server started",
		slog.String("addr", ":8080"),
		slog.Int("pid", os.Getpid()),
	)

	// Child logger with persistent attributes.
	reqLog := logger.With(
		"req_id", "abc-123",
		"user", "alice",
	)
	reqLog.Info("incoming request", "path", "/api/users")
	reqLog.Warn("slow query", "ms", 542)

	// Errors have a conventional "err" key.
	if err := doWork(); err != nil {
		reqLog.Error("work failed", "err", err)
	}

	// Groups create nested JSON.
	logger.Info("payment", slog.Group("amount",
		"value", 4250,
		"currency", "USD",
	))
}

func doWork() error { return errors.New("kaboom") }
`,
		Notes: []string{
			"Use the package-level default with slog.SetDefault(logger) so every package logs consistently.",
			"slog.With(...) creates a child logger — perfect for request-scoped fields (trace_id, user_id).",
			"Two handlers ship: NewJSONHandler (prod) and NewTextHandler (dev). Plug in custom handlers for OTel, etc.",
			"Always log err with the key \"err\" — both stdlib and most third-party tools key off it.",
		},
	},
}
