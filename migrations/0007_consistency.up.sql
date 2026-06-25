-- Phase 7 — Consistency Engine
-- PERSIST streaks, goals, and ladders

CREATE TABLE streaks (
    username          TEXT PRIMARY KEY REFERENCES users(username) ON DELETE CASCADE,
    current           INT NOT NULL DEFAULT 0,
    longest           INT NOT NULL DEFAULT 0,
    last_active_day   DATE,
    freezes_available INT NOT NULL DEFAULT 2
);

CREATE TABLE goals (
    username     TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    id           TEXT NOT NULL DEFAULT 'weekly-problems',
    type         TEXT NOT NULL DEFAULT 'problemsPerWeek',
    target       INT NOT NULL DEFAULT 5,
    progress     INT NOT NULL DEFAULT 0,
    period_start DATE NOT NULL,
    PRIMARY KEY (username, id)
);

CREATE TABLE ladders (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    from_rating INT NOT NULL,
    to_rating   INT NOT NULL
);

CREATE TABLE ladder_rungs (
    ladder_id  TEXT NOT NULL REFERENCES ladders(id) ON DELETE CASCADE,
    problem_id TEXT NOT NULL,
    rating     INT NOT NULL,
    topic_id   TEXT,
    ord        INT NOT NULL DEFAULT 0,
    PRIMARY KEY (ladder_id, problem_id)
);

CREATE INDEX idx_ladder_rungs_ladder ON ladder_rungs(ladder_id, ord);
