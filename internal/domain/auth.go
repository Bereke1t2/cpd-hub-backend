package domain

// Auth request/response models

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	FullName        string `json:"fullName"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// Auth repository interface (e.g. for real auth backend)
type AuthRepository interface {
	Login(req *LoginRequest) (*AuthResponse, error)
	Signup(req *SignupRequest) (*AuthResponse, error)
}
