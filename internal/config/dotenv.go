// Package config holds runtime configuration helpers — currently just a tiny
// .env loader.  No external dependencies on purpose; this is a teaching app.
package config

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"strings"
)

// LoadDotEnv reads KEY=VALUE pairs from `path` (if it exists) and inserts
// them into the process environment.  Variables that are ALREADY set in the
// real environment win, so `GEMINI_API_KEY=… go run .` always overrides
// the file.
//
// Supports:
//   - blank lines and `# comments`
//   - KEY=value, KEY="quoted", KEY='quoted'
//   - leading `export KEY=value` (so the same file can be `source`d)
//
// Returns the number of keys it set.  A missing file is NOT an error —
// the loader is optional by design.
func LoadDotEnv(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	defer f.Close()

	loaded := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])

		if len(val) >= 2 {
			first, last := val[0], val[len(val)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return loaded, err
		}
		loaded++
	}
	return loaded, sc.Err()
}
