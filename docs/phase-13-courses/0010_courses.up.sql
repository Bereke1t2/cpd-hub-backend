-- Phase 13 — Courses. Copy to: migrations/0010_courses.up.sql
-- Structured learning modules with per-user lesson completion.

CREATE TABLE IF NOT EXISTS courses (
    id      TEXT PRIMARY KEY,
    title   TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    level   TEXT NOT NULL DEFAULT 'Beginner'
);

CREATE TABLE IF NOT EXISTS course_modules (
    id        TEXT PRIMARY KEY,
    course_id TEXT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title     TEXT NOT NULL,
    ord       INT  NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_course_modules_course ON course_modules(course_id, ord);

CREATE TABLE IF NOT EXISTS course_lessons (
    id           TEXT PRIMARY KEY,
    module_id    TEXT NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    kind         TEXT NOT NULL DEFAULT 'article', -- 'video' | 'article' | 'pdf'
    content_url  TEXT NOT NULL DEFAULT '',
    inline_text  TEXT NOT NULL DEFAULT '',
    duration_sec INT  NOT NULL DEFAULT 0,
    ord          INT  NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_course_lessons_module ON course_lessons(module_id, ord);

-- Per-user completion overlay. Presence of a row == completed.
CREATE TABLE IF NOT EXISTS user_lesson_progress (
    username     TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    lesson_id    TEXT NOT NULL REFERENCES course_lessons(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (username, lesson_id)
);
CREATE INDEX IF NOT EXISTS idx_user_lesson_progress_user ON user_lesson_progress(username);
