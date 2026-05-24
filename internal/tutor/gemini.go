package tutor

import (
	"context"
	"fmt"
	"strings"
)

// gemini is the Provider implementation for Google's Generative Language API
// (Gemini). Unexported on purpose — callers depend on the Provider interface.
type gemini struct {
	apiKey string
	model  string
}

// NewGemini builds a Provider backed by Google's Gemini API.
//
// The model defaults to "gemini-2.5-flash" and can be overridden via the
// GEMINI_MODEL environment variable.
func NewGemini(apiKey string) Provider {
	return &gemini{apiKey: apiKey, model: envOrDefault("GEMINI_MODEL", "gemini-2.5-flash")}
}

// Name returns the provider id surfaced to the UI ("google").
func (g *gemini) Name() string { return "google" }

// Model returns the model identifier currently in use.
func (g *gemini) Model() string { return g.model }

// Complete sends `history` (translated into Gemini's schema) plus the
// per-request system prompt to the model and returns the text reply.
//
// Gemini's content schema differs from the OpenAI lineage:
//
//	contents:          [{role: "user"|"model", parts: [{text: "…"}]}]
//	systemInstruction: {parts: [{text: "…"}]}
//
// "assistant" turns are renamed to "model" before transmission.
func (g *gemini) Complete(ctx context.Context, system string, history []Message) (string, error) {
	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}

	contents := make([]content, 0, len(history))
	for _, m := range history {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, content{Role: role, Parts: []part{{Text: m.Content}}})
	}

	payload := map[string]any{
		"contents":          contents,
		"systemInstruction": map[string]any{"parts": []part{{Text: system}}},
		"generationConfig": map[string]any{
			"maxOutputTokens": 1024,
			"temperature":     0.4,
		},
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", g.model)
	headers := map[string]string{"x-goog-api-key": g.apiKey}
	if err := postJSON(ctx, url, headers, payload, &out); err != nil {
		return "", fmt.Errorf("Gemini %w", err)
	}

	if out.PromptFeedback.BlockReason != "" {
		return "", fmt.Errorf("Gemini blocked the prompt: %s", out.PromptFeedback.BlockReason)
	}
	if len(out.Candidates) == 0 {
		return "", fmt.Errorf("Gemini returned no candidates")
	}
	parts := make([]string, 0, len(out.Candidates[0].Content.Parts))
	for _, p := range out.Candidates[0].Content.Parts {
		if p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("Gemini returned empty content (finish reason: %s)", out.Candidates[0].FinishReason)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n")), nil
}
