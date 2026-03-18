package security

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComparePassword compares a hashed password with a plaintext candidate.
func ComparePassword(hashed, candidate string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(candidate))
}
