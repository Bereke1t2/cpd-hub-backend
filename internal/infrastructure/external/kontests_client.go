package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

// kontestItem represents the JSON shape returned by https://kontests.net API.
type kontestItem struct {
	Name      string  `json:"name"`
	Site      string  `json:"site"`
	URL       string  `json:"url"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Duration  float64 `json:"duration"` // seconds
	Status    string  `json:"status"`
}

// Codeforces API shapes
type cfResponse struct {
	Status string      `json:"status"`
	Result []cfContest `json:"result"`
}

type cfContest struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	Phase            string `json:"phase"`
	StartTimeSeconds int64  `json:"startTimeSeconds"`
	DurationSeconds  int64  `json:"durationSeconds"`
}

// ContestStandingRow represents a single row in a contest leaderboard.
type ContestStandingRow struct {
	Rank    int     `json:"rank"`
	Handle  string  `json:"handle"`
	Points  float64 `json:"points"`
	Penalty int     `json:"penalty"`
	Solved  int     `json:"solved"`
	Rating  int     `json:"rating"`
}

// KontestsClient fetches contests from kontests.net
type KontestsClient struct {
	client *http.Client
}

// NewKontestsClient creates a client with a default timeout.
func NewKontestsClient() *KontestsClient {
	return &KontestsClient{client: &http.Client{Timeout: 8 * time.Second}}
}

// FetchPlatform fetches upcoming contests for a given platform (e.g. "codeforces", "leetcode").
// Special-case Codeforces to use the official Codeforces API (more reliable for phases/start times).
func (k *KontestsClient) FetchPlatform(platform string) ([]domain.Contest, error) {
	if strings.EqualFold(platform, "codeforces") {
		return k.fetchCodeforces()
	}

	api := fmt.Sprintf("https://kontests.net/api/v1/%s", url.PathEscape(platform))
	resp, err := k.client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from kontests", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []kontestItem
	if err := json.Unmarshal(body, &items); err != nil {
		// include raw body in the error to help debugging problematic responses
		return nil, fmt.Errorf("decoding kontests response: %w; body=%s", err, string(body))
	}

	out := make([]domain.Contest, 0, len(items))
	for _, it := range items {
		// kontests.net uses statuses like "BEFORE"/"CODING" etc. Be tolerant.
		stLower := strings.ToLower(it.Status)
		if stLower != "before" && stLower != "upcoming" {
			continue
		}

		// Try multiple parse strategies for start_time (RFC3339 / RFC3339Nano / unix seconds string)
		var st time.Time
		if it.StartTime == "" {
			continue
		}
		if t, perr := time.Parse(time.RFC3339, it.StartTime); perr == nil {
			st = t
		} else if t, perr := time.Parse(time.RFC3339Nano, it.StartTime); perr == nil {
			st = t
		} else if sec, perr := strconv.ParseInt(it.StartTime, 10, 64); perr == nil {
			st = time.Unix(sec, 0)
		} else {
			// Couldn't parse start_time; skip this item but include some context in logs via error return
			// Return an error so the caller can log the raw response and decide how to proceed.
			return nil, fmt.Errorf("unable to parse start_time for kontests item; raw=%s", string(body))
		}

		c := domain.Contest{
			ID:                  fmt.Sprintf("%s-%s", strings.ToLower(it.Site), url.PathEscape(strings.ReplaceAll(it.Name, " ", "-"))),
			Title:               it.Name,
			ContestURL:          it.URL,
			StartTime:           st,
			Duration:            (time.Duration(it.Duration) * time.Second).String(),
			Platform:            it.Site,
			NumberOfProblems:    0,
			NumberOfContestants: 0,
			Date:                st.Format("Jan 2, 2006"),
			IsPast:              false,
			IsParticipating:     false,
		}
		out = append(out, c)
	}
	return out, nil
}

// fetchCodeforces uses the official Codeforces API to get contest.list and filters phase=="BEFORE".
func (k *KontestsClient) fetchCodeforces() ([]domain.Contest, error) {
	api := "https://codeforces.com/api/contest.list"
	resp, err := k.client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from codeforces", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cr cfResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decoding codeforces response: %w; body=%s", err, string(body))
	}
	if cr.Status != "OK" {
		return nil, fmt.Errorf("codeforces api status: %s", cr.Status)
	}

	out := make([]domain.Contest, 0, len(cr.Result))
	for _, c := range cr.Result {
		if c.Phase != "BEFORE" {
			continue
		}
		st := time.Unix(c.StartTimeSeconds, 0)
		var durationStr string
		if c.DurationSeconds > 0 {
			durationStr = (time.Duration(c.DurationSeconds) * time.Second).String()
		} else {
			durationStr = ""
		}
		contestURL := fmt.Sprintf("https://codeforces.com/contest/%d", c.Id)
		id := fmt.Sprintf("codeforces-%d", c.Id)

		// attempt to fetch standings with count=1 to get problems array and total participants when available
		problemsCount := 0
		participantCount := 0
		standAPI := fmt.Sprintf("https://codeforces.com/api/contest.standings?contestId=%d&from=1&count=1&showUnofficial=false", c.Id)
		if sresp, serr := k.client.Get(standAPI); serr == nil {
			defer sresp.Body.Close()
			if sresp.StatusCode == http.StatusOK {
				b, _ := io.ReadAll(sresp.Body)
				var tmp struct {
					Status string `json:"status"`
					Result struct {
						Problems []interface{} `json:"problems"`
						Rows     []interface{} `json:"rows"`
					} `json:"result"`
				}
				if err := json.Unmarshal(b, &tmp); err == nil {
					problemsCount = len(tmp.Result.Problems)
					participantCount = len(tmp.Result.Rows)
				}
			}
		}

		out = append(out, domain.Contest{
			ID:                  id,
			Title:               c.Name,
			ContestURL:          contestURL,
			StartTime:           st,
			Duration:            durationStr,
			Platform:            "Codeforces",
			NumberOfProblems:    problemsCount,
			NumberOfContestants: participantCount,
			Date:                st.Format("Jan 2, 2006"),
			IsPast:              false,
			IsParticipating:     false,
		})
	}
	return out, nil
}

// FetchContestStandings fetches the standings for a Codeforces contest using the
// official API and returns a slice of ContestStandingRow plus the total number
// of rows returned by the API (useful as a registration/participant count when
// the contest has started/finished). Be careful with large `count` values.
func (k *KontestsClient) FetchContestStandings(contestID int, from int, count int, showUnofficial bool) ([]ContestStandingRow, int, error) {
	api := fmt.Sprintf("https://codeforces.com/api/contest.standings?contestId=%d&from=%d&count=%d&showUnofficial=%v", contestID, from, count, showUnofficial)
	resp, err := k.client.Get(api)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status %d from codeforces standings", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	// Minimal decode structure for the parts we need.
	var s struct {
		Status string `json:"status"`
		Result struct {
			Problems []struct{} `json:"problems"`
			Rows     []struct {
				Rank    int         `json:"rank"`
				Points  json.Number `json:"points"`
				Penalty int         `json:"penalty"`
				Party   struct {
					Members []struct {
						Handle string `json:"handle"`
						Rating int    `json:"rating"`
					} `json:"members"`
				} `json:"party"`
				ProblemResults []struct {
					Points  json.Number `json:"points"`
					Penalty int         `json:"penalty"`
				} `json:"problemResults"`
			} `json:"rows"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &s); err != nil {
		return nil, 0, fmt.Errorf("decoding codeforces standings: %w; body=%s", err, string(body))
	}
	if s.Status != "OK" {
		return nil, 0, fmt.Errorf("codeforces standings status: %s", s.Status)
	}

	out := make([]ContestStandingRow, 0, len(s.Result.Rows))
	for _, r := range s.Result.Rows {
		// parse points which may be integer or float
		var pts float64
		if r.Points != "" {
			if f, perr := r.Points.Float64(); perr == nil {
				pts = f
			} else if ival, perr := r.Points.Int64(); perr == nil {
				pts = float64(ival)
			}
		}
		handle := ""
		rating := 0
		if len(r.Party.Members) > 0 {
			handle = r.Party.Members[0].Handle
			rating = r.Party.Members[0].Rating
		}
		solved := 0
		for _, pr := range r.ProblemResults {
			if pr.Points != "" {
				solved++
			}
		}
		out = append(out, ContestStandingRow{
			Rank:    r.Rank,
			Handle:  handle,
			Points:  pts,
			Penalty: r.Penalty,
			Solved:  solved,
			Rating:  rating,
		})
	}

	return out, len(s.Result.Rows), nil
}

// FetchUpcomingCodeforces fetches upcoming Codeforces contests and returns up to `limit` contests.
// This is a convenience wrapper around fetchCodeforces so callers can request the latest N.
func (k *KontestsClient) FetchUpcomingCodeforces(limit int) ([]domain.Contest, error) {
	all, err := k.fetchCodeforces()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit >= len(all) {
		return all, nil
	}
	return all[:limit], nil
}

// FetchRecentCodeforces fetches recently finished Codeforces contests and returns up to `limit` contests.
// It uses the official Codeforces API and returns contests sorted by start time descending (most recent first).
func (k *KontestsClient) FetchRecentCodeforces(limit int) ([]domain.Contest, error) {
	api := "https://codeforces.com/api/contest.list"
	resp, err := k.client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from codeforces", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cr cfResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decoding codeforces response: %w; body=%s", err, string(body))
	}
	if cr.Status != "OK" {
		return nil, fmt.Errorf("codeforces api status: %s", cr.Status)
	}

	// collect finished contests
	finished := make([]cfContest, 0, len(cr.Result))
	for _, c := range cr.Result {
		if c.Phase == "FINISHED" || c.Phase == "PENDING" || c.Phase == "POSTPONED" {
			// include FINISHED; other phases are included conservatively if you want more history
			finished = append(finished, c)
		}
	}

	// sort by start time descending
	sort.Slice(finished, func(i, j int) bool { return finished[i].StartTimeSeconds > finished[j].StartTimeSeconds })

	if limit > 0 && limit < len(finished) {
		finished = finished[:limit]
	}

	out := make([]domain.Contest, 0, len(finished))
	for _, c := range finished {
		st := time.Unix(c.StartTimeSeconds, 0)
		var durationStr string
		if c.DurationSeconds > 0 {
			durationStr = (time.Duration(c.DurationSeconds) * time.Second).String()
		} else {
			durationStr = ""
		}
		contestURL := fmt.Sprintf("https://codeforces.com/contest/%d", c.Id)
		id := fmt.Sprintf("codeforces-%d", c.Id)
		out = append(out, domain.Contest{
			ID:                  id,
			Title:               c.Name,
			ContestURL:          contestURL,
			StartTime:           st,
			Duration:            durationStr,
			Platform:            "Codeforces",
			NumberOfProblems:    0,
			NumberOfContestants: 0,
			Date:                st.Format("Jan 2, 2006"),
			IsPast:              true,
			IsParticipating:     false,
		})
	}
	return out, nil
}
