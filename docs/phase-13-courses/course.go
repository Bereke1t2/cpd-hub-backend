//go:build ignore
// Template for Phase 13 — copy to: internal/domain/course.go
//
// Courses feature. Entities mirror api.md §8 and the Flutter CourseModel (camelCase).
// `Completed` on a lesson is per-user and computed at read time from user_lesson_progress —
// it is NOT a stored column on course_lessons.
package domain

// Course is a structured learning module with ordered modules and lessons.
type Course struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Level   string   `json:"level"` // "Beginner" | "Intermediate" | "Advanced"
	Modules []Module `json:"modules"`
}

// Module groups lessons within a course, rendered in `ord` order.
type Module struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Lessons []Lesson `json:"lessons"`
}

// Lesson is a single unit of content. Kind is one of "video", "article", "pdf".
type Lesson struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Kind        string `json:"kind"`
	ContentURL  string `json:"contentUrl"`
	InlineText  string `json:"inlineText,omitempty"`  // only for kind=="article"
	DurationSec int    `json:"durationSec,omitempty"` // only for kind=="video"
	Completed   bool   `json:"completed"`             // per-user overlay, computed at read time
}

// CourseRepository is implemented in infrastructure/databases.
// All reads take the authenticated username so `Completed` is scoped to the caller.
type CourseRepository interface {
	List(username string) ([]*Course, error)
	Get(username, id string) (*Course, error)
	// CompleteLesson is idempotent. Returns ErrNotFound if the lesson does not
	// belong to the given course.
	CompleteLesson(username, courseID, lessonID string) error
}
