package lesson

// Catalog is the ordered list of lessons shown in the sidebar.  Order matters
// because the UI's "Next" / "Prev" buttons walk the slice, and the sidebar
// groups by category in first-appearance order.
//
// The catalog is assembled from per-category slices living in their own
// `lessons_<category>.go` files.  This file is the single place that decides
// the pedagogical order; to add a new section, define a new slice and append
// its name here in the right slot.
var Catalog = concat(
	// ── Foundations ─────────────────────────────────────────────────────
	basicsLessons,
	controlFlowLessons,
	dataStructuresLessons,
	methodsLessons,
	errorsLessons,
	concurrencyLessons,
	toolingLessons,
	stdlibLessons,
	webLessons,
	ecosystemLessons,

	// ── Senior / Interview Track ────────────────────────────────────────
	genericsLessons,
	memoryLessons,
	pitfallsLessons,
	algorithmsLessons,
	interviewPrepLessons,
)

// concat flattens N lesson slices into one, preallocating exact capacity.
func concat(groups ...[]Lesson) []Lesson {
	n := 0
	for _, g := range groups {
		n += len(g)
	}
	out := make([]Lesson, 0, n)
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}
