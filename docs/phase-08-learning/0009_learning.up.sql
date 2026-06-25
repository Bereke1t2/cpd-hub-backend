-- Template for Phase 8 — copy to: migrations/0009_learning.up.sql
-- Topic DAG + tracks + lessons. Arrays are modeled as edge tables so they're
-- FK-checked and queryable.

CREATE TABLE IF NOT EXISTS topics (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    category   TEXT NOT NULL DEFAULT '',
    summary    TEXT NOT NULL DEFAULT '',
    difficulty INT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS topic_prerequisites (
    topic_id        TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    prerequisite_id TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (topic_id, prerequisite_id),
    CHECK (topic_id <> prerequisite_id)            -- no self-loop
);

CREATE TABLE IF NOT EXISTS topic_problems (
    topic_id   TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    PRIMARY KEY (topic_id, problem_id)
);

CREATE TABLE IF NOT EXISTS topic_references (
    topic_id TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    url      TEXT NOT NULL,
    PRIMARY KEY (topic_id, url)
);

CREATE TABLE IF NOT EXISTS tracks (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon_name   TEXT NOT NULL DEFAULT 'school'
);

CREATE TABLE IF NOT EXISTS track_topics (
    track_id TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    topic_id TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    ord      INT NOT NULL DEFAULT 0,
    PRIMARY KEY (track_id, topic_id)
);

CREATE TABLE IF NOT EXISTS lessons (
    topic_id  TEXT PRIMARY KEY REFERENCES topics(id) ON DELETE CASCADE,
    body      TEXT NOT NULL DEFAULT '',
    key_ideas TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_topic_prereq_topic ON topic_prerequisites(topic_id);
CREATE INDEX IF NOT EXISTS idx_topic_problems_topic ON topic_problems(topic_id);
CREATE INDEX IF NOT EXISTS idx_track_topics_track ON track_topics(track_id, ord);
