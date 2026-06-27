package activity

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeActivityRepo struct {
	inserted []*domain.Activity
	err      error
}

func (f *fakeActivityRepo) List(int, int) ([]*domain.Activity, error) { return nil, nil }

func (f *fakeActivityRepo) InsertActivity(a *domain.Activity) error {
	f.inserted = append(f.inserted, a)
	return f.err
}

func TestRecorderRecordsSolveAndLike(t *testing.T) {
	now := time.Date(2026, 6, 27, 10, 0, 0, 0, time.FixedZone("EAT", 3*60*60))
	repo := &fakeActivityRepo{err: errors.New("ignored")}
	recorder := NewRecorder(repo)

	recorder.RecordSolve("alice", "Two Sum", now)
	recorder.RecordLike("alice", "Two Sum", now)

	if len(repo.inserted) != 2 {
		t.Fatalf("inserted = %d, want 2", len(repo.inserted))
	}
	if got := repo.inserted[0]; got.Username != "alice" || got.Type != "Solve" || got.Action != "solved 'Two Sum'" {
		t.Fatalf("solve activity = %+v", got)
	}
	if got := repo.inserted[1]; got.Username != "alice" || got.Type != "Like" || got.Action != "liked 'Two Sum'" {
		t.Fatalf("like activity = %+v", got)
	}
	if !strings.HasPrefix(repo.inserted[0].ID, "act-") || !strings.HasPrefix(repo.inserted[1].ID, "act-") {
		t.Fatalf("ids = %q, %q; want act- prefix", repo.inserted[0].ID, repo.inserted[1].ID)
	}
	if repo.inserted[0].Timestamp != now.UTC().Format(time.RFC3339) {
		t.Fatalf("timestamp = %q, want %q", repo.inserted[0].Timestamp, now.UTC().Format(time.RFC3339))
	}
}

func TestHumanizeSince(t *testing.T) {
	now := time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		ts   string
		want string
	}{
		{name: "invalid", ts: "not-a-time", want: "not-a-time"},
		{name: "seconds", ts: now.Add(-30 * time.Second).Format(time.RFC3339), want: "just now"},
		{name: "minutes", ts: now.Add(-5 * time.Minute).Format(time.RFC3339), want: "5 min ago"},
		{name: "hours", ts: now.Add(-3 * time.Hour).Format(time.RFC3339), want: "3 hr ago"},
		{name: "days", ts: now.Add(-48 * time.Hour).Format(time.RFC3339), want: "2 days ago"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HumanizeSince(tc.ts, now); got != tc.want {
				t.Fatalf("HumanizeSince() = %q, want %q", got, tc.want)
			}
		})
	}
}
