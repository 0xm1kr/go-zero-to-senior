package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON is the single place we serialize a response body.
//
// Indented output is intentional: this is a teaching app, and human-friendly
// `curl` debugging is more valuable than the handful of bytes pretty-printing
// costs. For a high-traffic API you'd want to flip the indent flag off.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("encode: %v", err)
	}
}
