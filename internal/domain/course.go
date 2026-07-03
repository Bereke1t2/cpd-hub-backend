package domain

// Course is a structured learning path with ordered modules and lessons.
// Lesson completion is overlaid per authenticated user at read time.
type Course struct {
	ID      string         `json:"id"`
	Title   string         `json:"title"`
	Summary string         `json:"summary"`
	Level   string         `json:"level"`
	Modules []CourseModule `json:"modules"`
}

type CourseModule struct {
	ID      string         `json:"id"`
	Title   string         `json:"title"`
	Lessons []CourseLesson `json:"lessons"`
}

type CourseLesson struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Kind        string `json:"kind"`
	ContentURL  string `json:"contentUrl"`
	InlineText  string `json:"inlineText,omitempty"`
	DurationSec int    `json:"durationSec,omitempty"`
	Completed   bool   `json:"completed"`
}

type LessonCompletion struct {
	LessonID  string `json:"lessonId"`
	Completed bool   `json:"completed"`
}

type CourseRepository interface {
	List(username string) ([]*Course, error)
	Get(username, id string) (*Course, error)
	CompleteLesson(username, courseID, lessonID string) error
}
