package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang-tut/internal/tutor"
)

// chatRequest is the wire shape POSTed by the frontend to /api/chat.
// LessonID and Code give the tutor service the context it needs to build
// a meaningful system prompt; Messages is the full conversation history,
// with the latest user turn at the end.
type chatRequest struct {
	LessonID string          `json:"lessonId"`
	Code     string          `json:"code"`
	Messages []tutor.Message `json:"messages"`
}

// chatResponse is the reply envelope. Exactly one of Message or Error is
// populated on a successful HTTP exchange; the frontend renders Error as a
// system-style message in the chat UI.
type chatResponse struct {
	Message    tutor.Message `json:"message,omitempty"`
	Error      string        `json:"error,omitempty"`
	RetryAfter int           `json:"retryAfter,omitempty"`
}

// chatStatusResponse extends the tutor status payload with static limits.
type chatStatusResponse struct {
	tutor.Status
	Limits ChatLimits `json:"limits"`
}

// maxHistoryTurns caps how many turns we forward to the LLM per request.
// The frontend persists chat history in localStorage and may send a very
// long conversation; trimming here protects against runaway token usage
// and keeps prompt latency predictable.
const maxHistoryTurns = 40

// handleChatStatus returns the tutor service's availability snapshot so the
// frontend can show "AI chat enabled (gemini-2.5-flash)" or a configuration
// hint without trying a full request first.
func (s *Server) handleChatStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, chatStatusResponse{
		Status: s.Tutor.Status(),
		Limits: s.ChatLimiter.Limits,
	})
}

// handleChat is the main chat endpoint. It validates the incoming request,
// trims the history to maxHistoryTurns, and asks the tutor service for the
// next assistant message. Provider/transport errors are returned with HTTP
// 200 + an Error field so the UI can render them inline rather than as a
// hard network failure.
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	limits := s.ChatLimiter.Limits
	r.Body = http.MaxBytesReader(w, r.Body, int64(limits.MaxBodyBytes))

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSON(w, http.StatusRequestEntityTooLarge, chatResponse{
				Error: "request body too large (max " + strconv.Itoa(limits.MaxBodyBytes) + " bytes)",
			})
			return
		}
		if err == io.EOF {
			writeJSON(w, http.StatusBadRequest, chatResponse{Error: "bad json"})
			return
		}
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "bad json"})
		return
	}
	if len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "no messages"})
		return
	}

	filtered := make([]tutor.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role != "user" && msg.Role != "assistant" {
			writeJSON(w, http.StatusBadRequest, chatResponse{Error: "invalid message role"})
			return
		}
		if len(msg.Content) > limits.MaxMessageChars {
			writeJSON(w, http.StatusBadRequest, chatResponse{
				Error: "message too long (max " + strconv.Itoa(limits.MaxMessageChars) + " characters)",
			})
			return
		}
		if strings.TrimSpace(msg.Content) != "" {
			filtered = append(filtered, msg)
		}
	}
	if len(filtered) == 0 {
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "no messages"})
		return
	}
	if filtered[len(filtered)-1].Role != "user" {
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "last message must be from the user"})
		return
	}
	if len(req.Code) > limits.MaxCodeChars {
		writeJSON(w, http.StatusBadRequest, chatResponse{
			Error: "code too long (max " + strconv.Itoa(limits.MaxCodeChars) + " characters)",
		})
		return
	}
	if len(filtered) > maxHistoryTurns {
		filtered = filtered[len(filtered)-maxHistoryTurns:]
	}

	if err := s.ChatLimiter.Check(clientIP(r)); err != nil {
		w.Header().Set("Retry-After", strconv.Itoa(err.RetryAfter))
		writeJSON(w, http.StatusTooManyRequests, chatResponse{
			Error:      err.Message,
			RetryAfter: err.RetryAfter,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	msg, err := s.Tutor.Reply(ctx, req.LessonID, req.Code, filtered)
	if err != nil {
		writeJSON(w, http.StatusOK, chatResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Message: msg})
}
