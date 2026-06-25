//go:build ignore
// Template for Phase 13 — copy to: internal/infrastructure/databases/courses_repo.go
//
// Loads the course tree in flat ordered queries (no N+1) and overlays the caller's
// lesson completion from user_lesson_progress. Errors are *domain.AppError.
package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type CoursesRepositoryDB struct{ client *postgres.Client }

func NewCoursesRepositoryDB(c *postgres.Client) *CoursesRepositoryDB {
	return &CoursesRepositoryDB{client: c}
}

// List returns every course with nested modules/lessons, `completed` overlaid for username.
func (r *CoursesRepositoryDB) List(username string) ([]*domain.Course, error) {
	return r.load(username, "") // empty id == all courses
}

// Get returns a single course or ErrNotFound.
func (r *CoursesRepositoryDB) Get(username, id string) (*domain.Course, error) {
	courses, err := r.load(username, id)
	if err != nil {
		return nil, err
	}
	if len(courses) == 0 {
		return nil, domain.ErrNotFound("course not found")
	}
	return courses[0], nil
}

// load fetches courses (optionally one by id), assembling the tree in Go.
// completedSet holds the lesson ids this user has finished.
func (r *CoursesRepositoryDB) load(username, onlyID string) ([]*domain.Course, error) {
	ctx := context.Background()

	// 1. completed lesson ids for this user (single round-trip).
	completed := map[string]bool{}
	cr, err := r.client.Pool.Query(ctx, `SELECT lesson_id FROM user_lesson_progress WHERE username=$1`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not load progress").Wrap(err)
	}
	for cr.Next() {
		var lid string
		if cr.Scan(&lid) == nil {
			completed[lid] = true
		}
	}
	cr.Close()

	// 2. courses (filtered if onlyID set).
	courseQ := `SELECT id, title, summary, level FROM courses`
	args := []interface{}{}
	if onlyID != "" {
		courseQ += ` WHERE id=$1`
		args = append(args, onlyID)
	}
	courseQ += ` ORDER BY id`
	rows, err := r.client.Pool.Query(ctx, courseQ, args...)
	if err != nil {
		return nil, domain.ErrInternal("could not list courses").Wrap(err)
	}
	order := []*domain.Course{}
	byID := map[string]*domain.Course{}
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Summary, &c.Level); err != nil {
			continue
		}
		c.Modules = []domain.Module{}
		order = append(order, &c)
		byID[c.ID] = &c
	}
	rows.Close()
	if len(order) == 0 {
		return order, nil
	}

	// 3. modules for those courses.
	modByID := map[string]*domain.Module{}
	modParent := map[string]string{} // moduleID -> courseID
	mr, err := r.client.Pool.Query(ctx,
		`SELECT m.id, m.course_id, m.title FROM course_modules m
		   JOIN courses c ON c.id = m.course_id
		  ORDER BY m.course_id, m.ord`)
	if err != nil {
		return nil, domain.ErrInternal("could not list modules").Wrap(err)
	}
	for mr.Next() {
		var m domain.Module
		var courseID string
		if err := mr.Scan(&m.ID, &courseID, &m.Title); err != nil {
			continue
		}
		parent, ok := byID[courseID]
		if !ok {
			continue // module of a course we're not returning
		}
		m.Lessons = []domain.Lesson{}
		parent.Modules = append(parent.Modules, m)
		// point at the slice element we just appended
		modRef := &parent.Modules[len(parent.Modules)-1]
		modByID[m.ID] = modRef
		modParent[m.ID] = courseID
	}
	mr.Close()

	// 4. lessons for those modules.
	lr, err := r.client.Pool.Query(ctx,
		`SELECT l.id, l.module_id, l.title, l.kind, l.content_url, l.inline_text, l.duration_sec
		   FROM course_lessons l
		   JOIN course_modules m ON m.id = l.module_id
		  ORDER BY l.module_id, l.ord`)
	if err != nil {
		return nil, domain.ErrInternal("could not list lessons").Wrap(err)
	}
	for lr.Next() {
		var l domain.Lesson
		var moduleID string
		if err := lr.Scan(&l.ID, &moduleID, &l.Title, &l.Kind, &l.ContentURL, &l.InlineText, &l.DurationSec); err != nil {
			continue
		}
		mod, ok := modByID[moduleID]
		if !ok {
			continue
		}
		l.Completed = completed[l.ID]
		mod.Lessons = append(mod.Lessons, l)
	}
	lr.Close()

	return order, nil
}

// CompleteLesson upserts a completion row. 404 if the lesson is not part of the course.
func (r *CoursesRepositoryDB) CompleteLesson(username, courseID, lessonID string) error {
	ctx := context.Background()

	// validate lesson belongs to the course
	var exists bool
	err := r.client.Pool.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM course_lessons l
		   JOIN course_modules m ON m.id = l.module_id
		   WHERE l.id=$1 AND m.course_id=$2)`, lessonID, courseID).Scan(&exists)
	if err != nil {
		return domain.ErrInternal("could not verify lesson").Wrap(err)
	}
	if !exists {
		return domain.ErrNotFound("lesson not found in course")
	}

	_, err = r.client.Pool.Exec(ctx,
		`INSERT INTO user_lesson_progress (username, lesson_id) VALUES ($1,$2)
		 ON CONFLICT (username, lesson_id) DO NOTHING`, username, lessonID)
	if err != nil {
		return domain.ErrInternal("could not mark lesson complete").Wrap(err)
	}
	return nil
}
