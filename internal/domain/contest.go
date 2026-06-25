package domain

import "time"

type Contest struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	ContestURL          string    `json:"contestUrl"`
	StartTime           time.Time `json:"startTime"`
	Duration            string    `json:"duration"`
	Platform            string    `json:"platform"`
	NumberOfProblems    int       `json:"numberOfProblems"`
	NumberOfContestants int       `json:"numberOfContestants"`
	Date                string    `json:"date"`
	IsPast              bool      `json:"isPast"`
	IsParticipating     bool      `json:"isParticipating"`
}

type ContestRepository interface {
	ListForUser(username string) ([]*Contest, error)
	GetByID(id string) (*Contest, error)
	Leaderboard(contestID string) ([]*LeaderboardEntry, error)

	Participate(username, contestID string) error
	Unparticipate(username, contestID string) error
}

type LeaderboardEntry struct {
	Rank           int      `json:"rank"`
	Username       string   `json:"username"`
	Rating         int      `json:"rating"`
	Score          int      `json:"score"`
	Penalty        int      `json:"penalty"`
	SolvedCount    int      `json:"solvedCount"`
	ProblemsSolved []string `json:"problemsSolved"`
}
