package domain

// Auth request/response models

// LoginRequest is accepted by POST /api/auth/login.
// The email field also accepts a username handle (the API has a single input field).
type LoginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignupRequest is accepted by POST /api/auth/signup.
type SignupRequest struct {
	FullName        string `json:"fullName"         binding:"required,min=2"`
	Username        string `json:"username"         binding:"omitempty,alphanum,min=3,max=20"`
	Email           string `json:"email"            binding:"required,email"`
	Password        string `json:"password"         binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword"  binding:"required,eqfield=Password"`
}

// RefreshRequest is accepted by POST /api/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// AuthResponse is returned by login, signup, and refresh.
type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken,omitempty"`
	User         UserProfile `json:"user"`
}

// UserRecord is the internal representation that includes the password hash.
// It is NEVER serialised — only used between the repo and usecase layers.
type UserRecord struct {
	Username     string
	Email        string
	FullName     string
	PasswordHash string
}

// AuthRepository is the data-access surface for the auth usecase.
// The usecase handles all business rules; the repo does only DB I/O.
type AuthRepository interface {
	FindByEmailOrUsername(login string) (*UserRecord, error)
	ExistsEmail(email string) (bool, error)
	UsernameTaken(username string) (bool, error)
	Insert(rec *UserRecord) error
}
