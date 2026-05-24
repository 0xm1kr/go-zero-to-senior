package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// runRequest is the wire shape posted to /api/run. Just a single field
// today, but kept as a struct so adding (filename, args, env) later is a
// non-breaking change.
type runRequest struct {
	Code string `json:"code"`
}

// runResponse is what the UI's playground panel renders. Stdout/Stderr are
// always populated (possibly empty); Error is set only on compile failure,
// non-zero exit, or timeout. Duration is a human-readable string ("123ms")
// rather than a number so the frontend can display it verbatim.
type runResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

// maxRunCodeBytes caps the size of a single submission to keep the runner
// (and the Playground API) from being flooded with multi-megabyte payloads.
// 64 KiB is roughly 1,500 lines of Go and well above any realistic lesson.
const maxRunCodeBytes = 64 * 1024

// handleRun executes user-submitted Go source through the configured Runner
// (Local or Playground) and returns the captured stdout/stderr plus a
// wall-clock duration. Errors are always returned with HTTP 200 so the UI
// can render them in the same output panel.
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if len(req.Code) > maxRunCodeBytes {
		http.Error(w, "code too large", http.StatusRequestEntityTooLarge)
		return
	}

	start := time.Now()
	result := s.Runner.Run(r.Context(), req.Code)

	resp := runResponse{
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: time.Since(start).Round(time.Millisecond).String(),
	}
	if result.Err != nil {
		resp.Error = result.Err.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}
