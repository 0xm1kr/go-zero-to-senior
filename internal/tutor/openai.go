package tutor

import (
	"context"
	"fmt"
	"strings"
)

// openai is the Provider implementation for OpenAI's chat-completions API.
// Unexported on purpose — callers depend on the Provider interface.
type openai struct {
	apiKey string
	model  string
}

// NewOpenAI builds a Provider backed by OpenAI's chat-completions API.
//
// The model defaults to "gpt-4o-mini" (a small, cheap, capable model that
// suits a tutoring chat) and can be overridden via OPENAI_MODEL.
func NewOpenAI(apiKey string) Provider {
	return &openai{apiKey: apiKey, model: envOrDefault("OPENAI_MODEL", "gpt-4o-mini")}
}

// Name returns the provider id surfaced to the UI ("openai").
func (o *openai) Name() string { return "openai" }

// Model returns the model identifier currently in use.
func (o *openai) Model() string { return o.model }

// Complete sends the conversation to /v1/chat/completions and returns the
// first choice's content. OpenAI carries the system prompt as a leading
// {role: "system"} message rather than a separate field, so we prepend it
// here before transmission.
func (o *openai) Complete(ctx context.Context, system string, history []Message) (string, error) {
	full := make([]Message, 0, len(history)+1)
	full = append(full, Message{Role: "system", Content: system})
	full = append(full, history...)

	payload := map[string]any{
		"model":       o.model,
		"messages":    full,
		"max_tokens":  1024,
		"temperature": 0.4,
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	headers := map[string]string{"Authorization": "Bearer " + o.apiKey}
	if err := postJSON(ctx, "https://api.openai.com/v1/chat/completions", headers, payload, &out); err != nil {
		return "", fmt.Errorf("OpenAI %w", err)
	}

	if len(out.Choices) == 0 {
		return "", fmt.Errorf("OpenAI returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
