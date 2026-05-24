package tutor

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang-tut/internal/lesson"
)

// Service is the chat use-case orchestrator.  Given the lesson context and
// the current conversation, it builds a system prompt and asks the Provider
// for a reply.  It has no opinion on HTTP, JSON, or transport — those live
// in the api package.
type Service struct {
	provider Provider
	lessons  lesson.Repository
}

// NewService wires a Service.  Provider may be nil — in that case Status()
// reports unavailable and Reply() returns a configuration error.
func NewService(p Provider, repo lesson.Repository) *Service {
	return &Service{provider: p, lessons: repo}
}

// Status mirrors the shape returned by GET /api/chat/status.  We define it
// here, not in api, because availability is a domain concern.
type Status struct {
	Available bool   `json:"available"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Hint      string `json:"hint,omitempty"`
}

// Status reports whether chat is available and which provider/model is in use.
func (s *Service) Status() Status {
	if s.provider == nil {
		return Status{
			Available: false,
			Hint:      "Set GEMINI_API_KEY (or ANTHROPIC_API_KEY / OPENAI_API_KEY) in your environment or .env and restart the server.",
		}
	}
	return Status{
		Available: true,
		Provider:  s.provider.Name(),
		Model:     s.provider.Model(),
	}
}

// Reply produces the next assistant message for the given conversation.
//
// lessonID + currentCode supply context: the student's lesson and whatever
// they currently have in the editor.  history is the full conversation so
// far (the last entry should be a user message).
func (s *Service) Reply(ctx context.Context, lessonID, currentCode string, history []Message) (Message, error) {
	if s.provider == nil {
		return Message{}, fmt.Errorf("no LLM API key configured")
	}

	var lp *lesson.Lesson
	if found, ok := s.lessons.ByID(lessonID); ok {
		lp = &found
	}
	system := buildSystemPrompt(lp, currentCode)

	reply, err := s.provider.Complete(ctx, system, history)
	if err != nil {
		return Message{}, err
	}
	return Message{Role: "assistant", Content: reply}, nil
}

// buildSystemPrompt assembles the per-request system message.  Pure function,
// kept private to the package because it's an implementation detail of the
// Service.
func buildSystemPrompt(l *lesson.Lesson, currentCode string) string {
	var b strings.Builder
	b.WriteString("You are a friendly, concise Go programming tutor helping an experienced engineer learn Go through an interactive tutorial app. ")
	b.WriteString("Answer questions about Go syntax, idioms, tooling, the standard library, and the ecosystem. ")
	b.WriteString("Prefer short, direct answers with small runnable code examples. Always wrap Go code in ```go fenced blocks. ")
	b.WriteString("If the student appears confused or wrong about something, gently correct them. ")
	b.WriteString("If a question is unrelated to Go, briefly redirect to the lesson.\n\n")

	if l != nil {
		fmt.Fprintf(&b, "## Current lesson\n\nTitle: %s\nCategory: %s\n\n", l.Title, l.Category)
		if desc := stripHTML(l.Description); desc != "" {
			b.WriteString("Description:\n")
			b.WriteString(desc)
			b.WriteString("\n\n")
		}
		if l.Code != "" {
			b.WriteString("Reference code shipped with this lesson:\n```go\n")
			b.WriteString(l.Code)
			b.WriteString("\n```\n\n")
		}
		if len(l.Notes) > 0 {
			b.WriteString("Key takeaways:\n")
			for _, n := range l.Notes {
				fmt.Fprintf(&b, "- %s\n", n)
			}
			b.WriteString("\n")
		}
	}
	if strings.TrimSpace(currentCode) != "" && (l == nil || strings.TrimSpace(currentCode) != strings.TrimSpace(l.Code)) {
		b.WriteString("The student's CURRENT editor contents (may differ from the reference code):\n```go\n")
		b.WriteString(currentCode)
		b.WriteString("\n```\n")
	}
	return b.String()
}

var htmlTagRE = regexp.MustCompile(`<[^>]+>`)

// stripHTML removes tags and decodes the handful of entities that show up
// in our lesson descriptions.  Good enough for prompt building — we don't
// need a real HTML parser.
func stripHTML(s string) string {
	s = htmlTagRE.ReplaceAllString(s, "")
	s = strings.NewReplacer(
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
		"&quot;", "\"",
		"&#39;", "'",
		"&nbsp;", " ",
	).Replace(s)
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(s)
}
