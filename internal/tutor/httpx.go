package tutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// postJSON is the single HTTP roundtrip shared by every provider. It exists
// to remove the JSON-encode / build-request / check-status / JSON-decode
// boilerplate that would otherwise be copy-pasted across Gemini, Anthropic,
// and OpenAI clients.
//
// Behaviour:
//  1. JSON-encode `body` and POST it to `url` with the supplied headers.
//  2. Execute via http.DefaultClient; the per-request timeout comes from ctx.
//  3. On HTTP 4xx/5xx return an error containing the status line plus a
//     truncated response body so callers can surface the upstream message.
//  4. Otherwise JSON-decode the response into `out` (pass nil to discard).
func postJSON(ctx context.Context, url string, headers map[string]string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(raw)))
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}
	return nil
}
