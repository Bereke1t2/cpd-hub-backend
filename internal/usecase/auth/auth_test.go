package auth

import (
	"errors"
	"testing"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeAuthRepo struct {
	usernameTaken map[string]bool
	emailTaken    bool
	inserted      *domain.UserRecord
}

func (f *fakeAuthRepo) FindByEmailOrUsername(string) (*domain.UserRecord, error) {
	return nil, errors.New("not used in this test")
}

func (f *fakeAuthRepo) ExistsEmail(string) (bool, error) { return f.emailTaken, nil }

func (f *fakeAuthRepo) UsernameTaken(username string) (bool, error) {
	return f.usernameTaken[username], nil
}

func (f *fakeAuthRepo) Insert(rec *domain.UserRecord) error {
	f.inserted = rec
	return nil
}

func TestSignupDerivesUniqueHandle(t *testing.T) {
	cases := []struct {
		name         string
		email        string
		username     string
		taken        map[string]bool
		wantUsername string
	}{
		{
			name:         "derive from local part",
			email:        "alice@example.com",
			wantUsername: "alice",
			taken:        map[string]bool{},
		},
		{
			name:         "suffix when taken",
			email:        "alice@example.com",
			wantUsername: "alice2",
			taken:        map[string]bool{"alice": true},
		},
		{
			name:         "keep explicit handle",
			email:        "zoe@example.com",
			username:     "zoe42",
			wantUsername: "zoe42",
			taken:        map[string]bool{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeAuthRepo{usernameTaken: tc.taken}
			uc := New(repo)

			res, err := uc.Signup(&domain.SignupRequest{
				FullName:        "Alice Example",
				Username:        tc.username,
				Email:           tc.email,
				Password:        "password123",
				ConfirmPassword: "password123",
			})
			if err != nil {
				t.Fatalf("Signup() error = %v", err)
			}
			if res.User.Username != tc.wantUsername {
				t.Fatalf("username = %q, want %q", res.User.Username, tc.wantUsername)
			}
			if repo.inserted == nil || repo.inserted.Username != tc.wantUsername {
				t.Fatalf("inserted username = %v, want %q", repo.inserted, tc.wantUsername)
			}
		})
	}
}
