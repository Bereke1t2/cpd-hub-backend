package domain

// User is a lightweight user entity used for user management endpoints.
// It references SocialLink declared in profile.go to avoid duplication.
type User struct {
	Username        string       `json:"username"`
	FullName        string       `json:"fullName"`
	Bio             string       `json:"bio,omitempty"`
	AvatarURL       string       `json:"avatarUrl,omitempty"`
	Rating          int          `json:"rating,omitempty"`
	Rank            string       `json:"rank,omitempty"`
	Division        string       `json:"division,omitempty"`
	SolvedProblems  int          `json:"solvedProblems,omitempty"`
	Contributions   int          `json:"contributions,omitempty"`
	SocialLinks     []SocialLink `json:"socialLinks,omitempty"`
}

// UserRepository defines operations for persisting users keyed by username.
type UserRepository interface {
	Create(user *User) error
	GetByUsername(username string) (*User, error)
	List() ([]*User, error)
}
