package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type CoursesRepositoryDB struct {
	client *postgres.Client
}

func NewCoursesRepositoryDB(client *postgres.Client) *CoursesRepositoryDB {
	return &CoursesRepositoryDB{client: client}
}

func (r *CoursesRepositoryDB) List(username string) ([]*domain.Course, error) {
	return r.load(username, "")
}

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

func (r *CoursesRepositoryDB) CompleteLesson(username, courseID, lessonID string) error {
	ctx := context.Background()
	var exists bool
	if err := r.client.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM course_lessons l
			JOIN course_modules m ON m.id = l.module_id
			WHERE l.id = $1 AND m.course_id = $2
		)`, lessonID, courseID).Scan(&exists); err != nil {
		return domain.ErrInternal("could not verify lesson").Wrap(err)
	}
	if !exists {
		return domain.ErrNotFound("lesson not found in course")
	}

	if _, err := r.client.Pool.Exec(ctx, `
		INSERT INTO user_lesson_progress (username, lesson_id)
		VALUES ($1, $2)
		ON CONFLICT (username, lesson_id) DO NOTHING`, username, lessonID); err != nil {
		return domain.ErrInternal("could not complete lesson").Wrap(err)
	}
	return nil
}

func (r *CoursesRepositoryDB) load(username, onlyID string) ([]*domain.Course, error) {
	ctx := context.Background()
	courses, byID, err := r.loadCourses(ctx, onlyID)
	if err != nil {
		return nil, err
	}
	if len(courses) == 0 {
		return courses, nil
	}

	moduleIDs, modules, err := r.loadModules(ctx, onlyID, byID)
	if err != nil {
		return nil, err
	}
	if len(moduleIDs) == 0 {
		return courses, nil
	}

	completed, err := r.loadCompletedLessons(ctx, username)
	if err != nil {
		return nil, err
	}
	if err := r.loadLessons(ctx, moduleIDs, modules, completed); err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *CoursesRepositoryDB) loadCourses(ctx context.Context, onlyID string) ([]*domain.Course, map[string]*domain.Course, error) {
	query := `SELECT id, title, summary, level FROM courses`
	args := []interface{}{}
	if onlyID != "" {
		query += ` WHERE id = $1`
		args = append(args, onlyID)
	}
	query += ` ORDER BY id`

	rows, err := r.client.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, domain.ErrInternal("could not list courses").Wrap(err)
	}
	defer rows.Close()

	courses := []*domain.Course{}
	byID := map[string]*domain.Course{}
	for rows.Next() {
		c := &domain.Course{Modules: []domain.CourseModule{}}
		if err := rows.Scan(&c.ID, &c.Title, &c.Summary, &c.Level); err != nil {
			continue
		}
		courses = append(courses, c)
		byID[c.ID] = c
	}
	return courses, byID, nil
}

type courseModuleRef struct {
	course *domain.Course
	index  int
}

func (r *CoursesRepositoryDB) loadModules(ctx context.Context, onlyID string, courses map[string]*domain.Course) ([]string, map[string]courseModuleRef, error) {
	query := `
		SELECT id, course_id, title
		FROM course_modules`
	args := []interface{}{}
	if onlyID != "" {
		query += ` WHERE course_id = $1`
		args = append(args, onlyID)
	}
	query += ` ORDER BY course_id, ord, id`

	rows, err := r.client.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, domain.ErrInternal("could not list course modules").Wrap(err)
	}
	defer rows.Close()

	moduleIDs := []string{}
	modules := map[string]courseModuleRef{}
	for rows.Next() {
		var module domain.CourseModule
		var courseID string
		if err := rows.Scan(&module.ID, &courseID, &module.Title); err != nil {
			continue
		}
		parent := courses[courseID]
		if parent == nil {
			continue
		}
		module.Lessons = []domain.CourseLesson{}
		parent.Modules = append(parent.Modules, module)
		modules[module.ID] = courseModuleRef{course: parent, index: len(parent.Modules) - 1}
		moduleIDs = append(moduleIDs, module.ID)
	}
	return moduleIDs, modules, nil
}

func (r *CoursesRepositoryDB) loadCompletedLessons(ctx context.Context, username string) (map[string]bool, error) {
	rows, err := r.client.Pool.Query(ctx, `SELECT lesson_id FROM user_lesson_progress WHERE username = $1`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not load lesson progress").Wrap(err)
	}
	defer rows.Close()

	completed := map[string]bool{}
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err == nil {
			completed[lessonID] = true
		}
	}
	return completed, nil
}

func (r *CoursesRepositoryDB) loadLessons(ctx context.Context, moduleIDs []string, modules map[string]courseModuleRef, completed map[string]bool) error {
	rows, err := r.client.Pool.Query(ctx, `
		SELECT id, module_id, title, kind, content_url, inline_text, duration_sec
		FROM course_lessons
		WHERE module_id = ANY($1)
		ORDER BY module_id, ord, id`, moduleIDs)
	if err != nil {
		return domain.ErrInternal("could not list course lessons").Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		var lesson domain.CourseLesson
		var moduleID string
		if err := rows.Scan(
			&lesson.ID,
			&moduleID,
			&lesson.Title,
			&lesson.Kind,
			&lesson.ContentURL,
			&lesson.InlineText,
			&lesson.DurationSec,
		); err != nil {
			continue
		}
		ref, ok := modules[moduleID]
		if !ok {
			continue
		}
		lesson.Completed = completed[lesson.ID]
		ref.course.Modules[ref.index].Lessons = append(ref.course.Modules[ref.index].Lessons, lesson)
	}
	return nil
}
