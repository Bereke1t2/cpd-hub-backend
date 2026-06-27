package domain

type Topic struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Summary         string   `json:"summary"`
	Difficulty      int      `json:"difficulty"`
	PrerequisiteIDs []string `json:"prerequisite_ids"`
	ProblemIDs      []string `json:"problem_ids"`
	ReferenceURLs   []string `json:"reference_urls"`
}

type Track struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TopicIDs    []string `json:"topic_ids"`
	IconName    string   `json:"icon_name"`
}

type Lesson struct {
	TopicID  string   `json:"topic_id"`
	Body     string   `json:"body"`
	KeyIdeas []string `json:"key_ideas"`
}

type LearningRepository interface {
	GetTopics() ([]*Topic, error)
	GetTracks() ([]*Track, error)
	GetLesson(topicID string) (*Lesson, error)
}
