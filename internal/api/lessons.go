package api

import (
	"net/http"
	"strings"

	"golang-tut/internal/lesson"
)

// handleLessons returns the catalog as an ordered slice of Summary records
// (id, title, category only). The frontend uses this to build the sidebar
// without paying for description/code/notes payloads it doesn't yet need.
func (s *Server) handleLessons(w http.ResponseWriter, r *http.Request) {
	all := s.Lessons.All()
	out := make([]lesson.Summary, 0, len(all))
	for _, l := range all {
		out = append(out, l.Summary())
	}
	writeJSON(w, http.StatusOK, out)
}

// handleLessonByID returns one full Lesson (description, starter code,
// notes) keyed by the trailing path segment. Returns 404 if no lesson
// matches the requested id.
func (s *Server) handleLessonByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/lessons/")
	l, ok := s.Lessons.ByID(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, l)
}
