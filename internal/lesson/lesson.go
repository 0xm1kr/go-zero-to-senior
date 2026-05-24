// Package lesson is the lesson domain.
//
// It exposes the Lesson value object, a lightweight Summary projection for
// the sidebar, and a Repository port so callers don't have to care whether
// lessons are loaded from memory, disk, or a database. The static curriculum
// lives in this package as well (catalog.go + lessons_*.go).
package lesson

// Lesson is one tutorial unit shown by the UI.
//
// Description is raw HTML so the curriculum can use headings, lists, links,
// tables, and inline <code> without pulling in a markdown dependency. Code
// is the initial buffer rendered into the in-browser editor when the lesson
// is opened. Notes appear in the "Key takeaways" callout below the editor.
type Lesson struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Code        string   `json:"code"`
	Notes       []string `json:"notes"`
}

// Summary is the lightweight projection used by the sidebar. Stripping out
// description/code/notes keeps the /api/lessons response small even as the
// curriculum grows; the full Lesson is fetched lazily on selection.
type Summary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// Summary projects a Lesson into a Summary suitable for the sidebar list.
func (l Lesson) Summary() Summary {
	return Summary{ID: l.ID, Title: l.Title, Category: l.Category}
}

// Repository is the access port for lessons.
//
// Today there's a single implementation (InMemoryRepository); tomorrow a
// FileRepository or SQLRepository could swap in without touching the api
// or tutor packages.
type Repository interface {
	// All returns every lesson in pedagogical order.
	All() []Lesson
	// ByID returns the lesson with the given id, or (zero, false) if no
	// such lesson exists.
	ByID(id string) (Lesson, bool)
}

// InMemoryRepository serves lessons from an in-process slice. Lookups by id
// are O(1) via a precomputed index built at construction time.
type InMemoryRepository struct {
	items []Lesson
	index map[string]int
}

// NewInMemoryRepository builds a repository from the given catalog,
// preserving its slice order (which the UI's Prev/Next buttons rely on).
func NewInMemoryRepository(items []Lesson) *InMemoryRepository {
	r := &InMemoryRepository{
		items: items,
		index: make(map[string]int, len(items)),
	}
	for i, l := range items {
		r.index[l.ID] = i
	}
	return r
}

// All returns the catalog in its original order.
func (r *InMemoryRepository) All() []Lesson { return r.items }

// ByID returns the lesson with the given id, or (Lesson{}, false) if no
// such lesson exists.
func (r *InMemoryRepository) ByID(id string) (Lesson, bool) {
	i, ok := r.index[id]
	if !ok {
		return Lesson{}, false
	}
	return r.items[i], true
}
