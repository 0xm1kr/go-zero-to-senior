package api

import (
	"context"
	"encoding/json"
	"net/http"
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
	Message tutor.Message `json:"message,omitempty"`
	Error   string        `json:"error,omitempty"`
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
	writeJSON(w, http.StatusOK, s.Tutor.Status())
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
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "bad json"})
		return
	}
	if len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "no messages"})
		return
	}
	if len(req.Messages) > maxHistoryTurns {
		req.Messages = req.Messages[len(req.Messages)-maxHistoryTurns:]
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	msg, err := s.Tutor.Reply(ctx, req.LessonID, req.Code, req.Messages)
	if err != nil {
		writeJSON(w, http.StatusOK, chatResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Message: msg})
}
