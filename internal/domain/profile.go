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

// Profile repository
type ProfileRepository interface {
	ListUsers() ([]*UserProfile, error)
	GetProfile(username string) (*UserProfile, error)
	CreateUser(user *UserProfile) error
	UpdateUser(user *UserProfile) error
	DeleteUser(username string) error
}
