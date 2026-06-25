-- Template for Phase 2 — copy to: migrations/0002_analytics.up.sql
-- Tables that power the profile analytics endpoints. EnsureAllTables never created
-- these, which is why heatmap/rating-history/attendance/submissions can't persist today.

CREATE TABLE IF NOT EXISTS submissions (
    id             TEXT PRIMARY KEY,
    username       TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    problem_id     TEXT REFERENCES problems(id) ON DELETE SET NULL,
    problem_title  TEXT NOT NULL,
    status         TEXT NOT NULL,            -- Accepted | Wrong Answer | TLE | ...
    language       TEXT,
    execution_time TEXT,
    memory_used    TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rating_history (
    id         BIGSERIAL PRIMARY KEY,
    username   TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    day        DATE NOT NULL,
    rating     INT NOT NULL,
    UNIQUE (username, day)
);

-- One row per user per active day; the heatmap is SELECT day, count FROM daily_solves.
CREATE TABLE IF NOT EXISTS daily_solves (
    username   TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    day        DATE NOT NULL,
    count      INT NOT NULL DEFAULT 0,
    PRIMARY KEY (username, day)
);

CREATE TABLE IF NOT EXISTS attendance (
    id         BIGSERIAL PRIMARY KEY,
    username   TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    day        DATE NOT NULL,
    status     TEXT NOT NULL,                -- Present | Absent | Excused
    UNIQUE (username, day)
);

CREATE INDEX IF NOT EXISTS idx_submissions_username_time ON submissions(username, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rating_history_username ON rating_history(username, day);
CREATE INDEX IF NOT EXISTS idx_daily_solves_username_day ON daily_solves(username, day);
CREATE INDEX IF NOT EXISTS idx_attendance_username_day ON attendance(username, day);
