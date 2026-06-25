-- Template for Phase 2 — copy to: migrations/0001_init.up.sql
-- Baseline schema: the tables the app already uses, cleaned up with proper types,
-- timestamps, and a separate email column (Phase 3 backfills/enforces it).

CREATE TABLE IF NOT EXISTS users (
    username      TEXT PRIMARY KEY,
    email         TEXT UNIQUE,
    full_name     TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    bio           TEXT,
    avatar_url    TEXT,
    rating        INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS problems (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    difficulty  TEXT NOT NULL DEFAULT 'Medium',
    topic_tags  TEXT,                       -- comma-separated; parsed in repo
    likes       INT NOT NULL DEFAULT 0,     -- denormalized counter (see Phase 4)
    dislikes    INT NOT NULL DEFAULT 0,
    deep_link   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contests (
    id                    TEXT PRIMARY KEY,
    title                 TEXT NOT NULL,
    contest_url           TEXT,
    start_time            TIMESTAMPTZ,
    duration              TEXT,
    platform              TEXT,
    number_of_problems    INT NOT NULL DEFAULT 0,
    number_of_contestants INT NOT NULL DEFAULT 0,
    date                  TEXT,
    is_past               BOOLEAN NOT NULL DEFAULT false,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS profiles (
    username    TEXT PRIMARY KEY REFERENCES users(username) ON DELETE CASCADE,
    bio         TEXT,
    rating      INT NOT NULL DEFAULT 0,
    avatar_url  TEXT,
    rank        TEXT,
    division    TEXT,
    global_rank INT
);

CREATE TABLE IF NOT EXISTS activity (
    id         TEXT PRIMARY KEY,
    username   TEXT REFERENCES users(username) ON DELETE CASCADE,
    action     TEXT NOT NULL,
    type       TEXT NOT NULL,
    timestamp  TEXT,                          -- legacy display string, kept for client-compat
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS info (
    title       TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_activity_created_at ON activity(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_contests_start_time ON contests(start_time);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
