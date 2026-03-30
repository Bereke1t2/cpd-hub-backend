package domain

type UserProfile struct {
	Username              string       `json:"username"`
	FullName              string       `json:"fullName"`
	Bio                   string       `json:"bio,omitempty"`
	AvatarURL             string       `json:"avatarUrl,omitempty"`
	Rating                int          `json:"rating,omitempty"`
	Rank                  string       `json:"rank,omitempty"`
	Division              string       `json:"division,omitempty"`
	SolvedProblems        int          `json:"solvedProblems,omitempty"`
	Contributions         int          `json:"contributions,omitempty"`
	GlobalRank            int          `json:"globalRank,omitempty"`
	AttendedContestsCount int          `json:"attendedContestsCount,omitempty"`
	SocialLinks           []SocialLink `json:"socialLinks,omitempty"`
}

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Handle   string `json:"handle"`
}

// Heatmap entry for profile activity
type HeatmapEntry struct {
	Date       string `json:"date"`
	SolveCount int    `json:"solveCount"`
}

// Rating history entry
type RatingEntry struct {
	Date   string `json:"date"`
	Rating int    `json:"rating"`
}

// Attendance entry
type AttendanceEntry struct {
	Date   string `json:"date"`
	Status string `json:"status"`
}

// Submission entry
type Submission struct {
	ID           string `json:"id"`
	ProblemID    string `json:"problemId"`
	ProblemTitle string `json:"problemTitle"`
	Status       string `json:"status"`
	Language     string `json:"language"`
	ExecutionTime string `json:"executionTime,omitempty"`
	MemoryUsed    string `json:"memoryUsed,omitempty"`
	Timestamp     string `json:"timestamp"`
}

// Profile repository
type ProfileRepository interface {
	ListUsers() ([]*UserProfile, error)
	GetProfile(username string) (*UserProfile, error)
	CreateUser(user *UserProfile) error
	UpdateUser(user *UserProfile) error
	DeleteUser(username string) error

	// Profile related data
	GetProfileHeatmap(username string) ([]HeatmapEntry, error)
	GetProfileRatingHistory(username string) ([]RatingEntry, error)
	GetProfileAttendance(username string) ([]AttendanceEntry, error)
	GetProfileSubmissions(username string) ([]Submission, error)
}
