// smoke compiles every non-blocking lesson's Code via the /api/run endpoint
// and reports any that fail. A handful of lessons that intentionally block
// the runner (HTTP servers, network clients) are skipped by id.
//
// Usage:
//
//	go run ./cmd/smoke                                  # check the whole curriculum
//	go run ./cmd/smoke -base http://localhost:8774      # custom server URL
//	go run ./cmd/smoke -only intro,slices               # check a subset
//
// Exit status is 0 on success, 1 on lesson failures, 2 on infrastructure
// errors (server unreachable, etc.). Designed to be run against a freshly
// started dev server after curriculum edits.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// skip lists lesson ids that are expected to NOT terminate within the
// runner's per-request timeout. These are HTTP-server demos and the like
// where blocking forever is the whole point of the lesson.
var skip = map[string]bool{
	"http-server":     true, // blocks: starts ListenAndServe
	"http-client":     true, // hits go.dev, slow / network-dependent
	"middleware":      true, // blocks
	"production-http": true, // blocks
}

// lesson mirrors the shape returned by /api/lessons/{id}. Only the fields
// we use are declared; the rest are dropped during JSON decode.
type lesson struct {
	ID, Category, Title, Code string
}

// runResp mirrors the shape returned by /api/run.
type runResp struct {
	Stdout, Stderr, Error, Duration string
}

func main() {
	base := flag.String("base", "http://localhost:8774", "server base URL")
	only := flag.String("only", "", "comma-separated lesson IDs to test (default: all)")
	flag.Parse()

	summaries := fetchLessons(*base)
	filter := buildFilter(*only)

	var failures []string
	checked := 0
	for _, s := range summaries {
		if len(filter) > 0 && !filter[s.ID] {
			continue
		}
		if skip[s.ID] {
			fmt.Printf("[SKIP]  %s\n", s.ID)
			continue
		}
		l := fetchLesson(*base, s.ID)
		if strings.TrimSpace(l.Code) == "" {
			fmt.Printf("[EMPTY] %s\n", s.ID)
			continue
		}
		checked++
		if err := runLesson(*base, l); err != nil {
			fmt.Printf("[FAIL]  %s  %v\n", s.ID, err)
			failures = append(failures, s.ID)
		}
	}

	fmt.Printf("\nchecked %d, failures %d\n", checked, len(failures))
	if len(failures) > 0 {
		os.Exit(1)
	}
}

// fetchLessons gets the catalog summary list from the server.
func fetchLessons(base string) []struct{ ID string } {
	resp, err := http.Get(base + "/api/lessons")
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()
	var out []struct{ ID string }
	must(json.NewDecoder(resp.Body).Decode(&out))
	return out
}

// fetchLesson gets a single fully-populated lesson by id.
func fetchLesson(base, id string) lesson {
	resp, err := http.Get(base + "/api/lessons/" + id)
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()
	var l lesson
	must(json.NewDecoder(resp.Body).Decode(&l))
	return l
}

// buildFilter parses the -only flag value into a set. Returns nil/empty
// when no filter is requested.
func buildFilter(only string) map[string]bool {
	if only == "" {
		return nil
	}
	set := map[string]bool{}
	for _, id := range strings.Split(only, ",") {
		if id = strings.TrimSpace(id); id != "" {
			set[id] = true
		}
	}
	return set
}

// runLesson POSTs the lesson's code to /api/run and reports success or the
// captured error. On success it also prints a one-line "[OK]" trace with
// the run duration.
func runLesson(base string, l lesson) error {
	body, _ := json.Marshal(map[string]string{"code": l.Code})
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Post(base+"/api/run", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var res runResp
	if err := json.Unmarshal(raw, &res); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if res.Error != "" {
		return fmt.Errorf("%s stderr=%q", res.Error, truncate(res.Stderr, 120))
	}
	fmt.Printf("[OK]    %s  (%s)\n", l.ID, res.Duration)
	return nil
}

// truncate clamps `s` to at most n bytes for log readability, replacing
// embedded newlines with single spaces.
func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

// must aborts the program with an exit status of 2 when err is non-nil.
// Used for "this should never happen" infrastructure errors.
func must(err error) {
	if err != nil {
		die(err)
	}
}

// die prints err to stderr and exits with status 2 (distinct from the
// status-1 "some lessons failed" exit so CI can differentiate).
func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
