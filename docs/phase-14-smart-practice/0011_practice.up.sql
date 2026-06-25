-- Phase 14 — Smart Practice. Copy to: migrations/0011_practice.up.sql
-- Per-user SM-2 review cards and contest upsolves.

CREATE TABLE IF NOT EXISTS review_items (
    username    TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    problem_id  TEXT NOT NULL,
    due_date    TIMESTAMPTZ NOT NULL DEFAULT now(),
    interval    INT  NOT NULL DEFAULT 1,
    ease        DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    repetitions INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (username, problem_id)
);
-- "due today" lookups: filter by user, order/scan by due_date.
CREATE INDEX IF NOT EXISTS idx_review_items_due ON review_items(username, due_date);

CREATE TABLE IF NOT EXISTS upsolve_items (
    username      TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    problem_id    TEXT NOT NULL,
    contest_id    TEXT NOT NULL DEFAULT '',
    contest_title TEXT NOT NULL DEFAULT '',
    problem_title TEXT NOT NULL DEFAULT '',
    resolved      BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (username, problem_id)
);
CREATE INDEX IF NOT EXISTS idx_upsolve_items_user ON upsolve_items(username);
