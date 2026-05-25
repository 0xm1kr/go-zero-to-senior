// Package api exposes the HTTP transport: routing, middleware, and
// per-domain handlers.  It depends on the domain packages but the domain
// packages do not depend on it (no import of net/http in lesson, runner,
// or tutor).
package api

import (
	"io/fs"
	"log"
	"net/http"
	"time"

	"golang-tut/internal/lesson"
	"golang-tut/internal/runner"
	"golang-tut/internal/tutor"
)

// Server bundles the dependencies needed to serve every endpoint.
// Construct one via NewServer in main and call Handler() to mount it on
// an http.Server.
type Server struct {
	Lessons      lesson.Repository
	Runner       runner.Runner // interface — backed by Local or Playground
	Tutor        *tutor.Service
	ChatLimiter  *ChatLimiter
	Static       fs.FS
}

// NewServer is a small convenience constructor.  The fields could also be
// set directly; this just documents what's required.
func NewServer(lessons lesson.Repository, r runner.Runner, t *tutor.Service, static fs.FS) *Server {
	return &Server{
		Lessons:     lessons,
		Runner:      r,
		Tutor:       t,
		ChatLimiter: NewChatLimiterFromEnv(),
		Static:      static,
	}
}

// Handler returns the fully-wired http.Handler with routes and middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.FS(s.Static)))

	mux.HandleFunc("/healthz", handleHealth) // Cloud Run / k8s liveness probe
	mux.HandleFunc("/api/lessons", s.handleLessons)
	mux.HandleFunc("/api/lessons/", s.handleLessonByID)
	mux.HandleFunc("/api/run", s.handleRun)
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/chat/status", s.handleChatStatus)

	return logRequests(mux)
}

// handleHealth is a cheap, dependency-free probe used by orchestrators
// (Cloud Run, Kubernetes, etc.) to confirm the process is up.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// logRequests is a tiny middleware that logs each request.  Middlewares in
// Go are just functions: http.Handler in → http.Handler out.
//
// /healthz is skipped because Cloud Run / k8s hammer it on a short
// interval and flooding the logs with probe pings is pure noise.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
