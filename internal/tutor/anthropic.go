package tutor

import (
	"context"
	"fmt"
	"strings"
)

// anthropic is the Provider implementation for Anthropic's Messages API
// (Claude). Unexported on purpose — callers depend on the Provider interface.
type anthropic struct {
	apiKey string
	model  string
}

// NewAnthropic builds a Provider backed by Anthropic's Messages API.
//
// The model defaults to "claude-haiku-4-5" and can be overridden via the
// ANTHROPIC_MODEL environment variable.
func NewAnthropic(apiKey string) Provider {
	return &anthropic{apiKey: apiKey, model: envOrDefault("ANTHROPIC_MODEL", "claude-haiku-4-5")}
}

// Name returns the provider id surfaced to the UI ("anthropic").
func (a *anthropic) Name() string { return "anthropic" }

// Model returns the model identifier currently in use.
func (a *anthropic) Model() string { return a.model }

// Complete posts the (system, history) pair to Anthropic's /v1/messages
// endpoint and concatenates the returned text blocks into a single reply.
//
// Anthropic's schema accepts `system` as a top-level field and our
// (role, content) message shape unchanged, so no translation is needed.
func (a *anthropic) Complete(ctx context.Context, system string, history []Message) (string, error) {
	payload := map[string]any{
		"model":      a.model,
		"system":     system,
		"messages":   history,
		"max_tokens": 1024,
	}

	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	headers := map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
	}
	if err := postJSON(ctx, "https://api.anthropic.com/v1/messages", headers, payload, &out); err != nil {
		return "", fmt.Errorf("Anthropic %w", err)
	}

	parts := make([]string, 0, len(out.Content))
	for _, c := range out.Content {
		if c.Type == "text" {
			parts = append(parts, c.Text)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("Anthropic returned no text content")
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n")), nil
}
