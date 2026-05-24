// Package tutor is the LLM-backed chat domain.  It exposes a Service that
// orchestrates a request (lesson context + history) against any Provider,
// plus three concrete Providers: Gemini, Anthropic, OpenAI.
package tutor

import (
	"context"
	"os"
)

// Message is the wire-shape exchanged with the frontend AND fed into each
// provider after a small role translation.  Role is "user" or "assistant".
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Provider is the LLM adapter port.  Implementations are interchangeable —
// the Service knows nothing about HTTP, vendors, or model strings.
type Provider interface {
	Name() string  // "google", "anthropic", "openai"
	Model() string // e.g. "gemini-2.5-flash"
	Complete(ctx context.Context, system string, history []Message) (string, error)
}

// SelectFromEnv returns the highest-priority provider whose API key is set
// in the environment, or nil if none is configured.
//
// Priority order: Google → Anthropic → OpenAI.
func SelectFromEnv() Provider {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return NewGemini(key)
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return NewAnthropic(key)
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return NewOpenAI(key)
	}
	return nil
}

// envOrDefault is a tiny shared helper so each provider's model lookup is
// a one-liner.
func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
