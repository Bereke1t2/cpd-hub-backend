//go:build ignore

// Template for Phase 8 — copy to: internal/infrastructure/databases/learning_repo.go
//
// Assembles each topic's arrays from edge tables in a few batched queries (no
// N+1). Groups edges in Go after a single scan per edge table.
//
package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type LearningRepositoryDB struct{ client *postgres.Client }

func NewLearningRepositoryDB(c *postgres.Client) *LearningRepositoryDB {
	return &LearningRepositoryDB{client: c}
}

func (r *LearningRepositoryDB) GetTopics() ([]*domain.Topic, error) {
	ctx := context.Background()

	// 1. base topics
	rows, err := r.client.Pool.Query(ctx,
		`SELECT id, name, category, summary, difficulty FROM topics ORDER BY difficulty, id`)
	if err != nil {
		return nil, domain.ErrInternal("could not load topics").Wrap(err)
	}
	defer rows.Close()

	byID := map[string]*domain.Topic{}
	order := []string{}
	for rows.Next() {
		var t domain.Topic
		if err := rows.Scan(&t.ID, &t.Name, &t.Category, &t.Summary, &t.Difficulty); err != nil {
			continue
		}
		t.PrerequisiteIDs = []string{}
		t.ProblemIDs = []string{}
		t.ReferenceURLs = []string{}
		byID[t.ID] = &t
		order = append(order, t.ID)
	}

	// 2. edges, each in one query, grouped in Go
	collect(ctx, r, `SELECT topic_id, prerequisite_id FROM topic_prerequisites`,
		func(tid, v string) {
			if t := byID[tid]; t != nil {
				t.PrerequisiteIDs = append(t.PrerequisiteIDs, v)
			}
		})
	collect(ctx, r, `SELECT topic_id, problem_id FROM topic_problems`,
		func(tid, v string) {
			if t := byID[tid]; t != nil {
				t.ProblemIDs = append(t.ProblemIDs, v)
			}
		})
	collect(ctx, r, `SELECT topic_id, url FROM topic_references`,
		func(tid, v string) {
			if t := byID[tid]; t != nil {
				t.ReferenceURLs = append(t.ReferenceURLs, v)
			}
		})

	out := make([]*domain.Topic, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out, nil
}

func (r *LearningRepositoryDB) GetTracks() ([]*domain.Track, error) {
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx,
		`SELECT id, title, description, COALESCE(icon_name,'school') FROM tracks ORDER BY id`)
	if err != nil {
		return nil, domain.ErrInternal("could not load tracks").Wrap(err)
	}
	defer rows.Close()
	byID := map[string]*domain.Track{}
	order := []string{}
	for rows.Next() {
		var t domain.Track
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.IconName); err != nil {
			continue
		}
		t.TopicIDs = []string{}
		byID[t.ID] = &t
		order = append(order, t.ID)
	}
	collect(ctx, r, `SELECT track_id, topic_id FROM track_topics ORDER BY ord`,
		func(tid, v string) {
			if t := byID[tid]; t != nil {
				t.TopicIDs = append(t.TopicIDs, v)
			}
		})
	out := make([]*domain.Track, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out, nil
}

func (r *LearningRepositoryDB) GetLesson(topicID string) (*domain.Lesson, error) {
	row := r.client.Pool.QueryRow(context.Background(),
		`SELECT topic_id, body, key_ideas FROM lessons WHERE topic_id=$1`, topicID)
	var l domain.Lesson
	var keyIdeas []string
	if err := row.Scan(&l.TopicID, &l.Body, &keyIdeas); err != nil {
		return nil, domain.ErrNotFound("no lesson for topic").Wrap(err)
	}
	l.KeyIdeas = keyIdeas // store key_ideas as a TEXT[] column; pgx scans into []string
	if l.KeyIdeas == nil {
		l.KeyIdeas = []string{}
	}
	return &l, nil
}

// collect runs a two-column (topic_id, value) query and calls fn per row.
func collect(ctx context.Context, r *LearningRepositoryDB, q string, fn func(tid, v string)) {
	rows, err := r.client.Pool.Query(ctx, q)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tid, v string
		if rows.Scan(&tid, &v) == nil {
			fn(tid, v)
		}
	}
}
