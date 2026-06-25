CREATE TABLE IF NOT EXISTS bookmarks (
    username    TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    problem_id  TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (username, problem_id)
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_username ON bookmarks(username, created_at DESC);
