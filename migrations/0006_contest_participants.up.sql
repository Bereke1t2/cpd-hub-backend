-- Phase 6 — Contest Participation
-- Tracks which users have marked themselves as participating in a contest.

CREATE TABLE contest_participants (
    username   TEXT REFERENCES users(username) ON DELETE CASCADE,
    contest_id TEXT NOT NULL,
    PRIMARY KEY (username, contest_id)
);

-- Index for checking participation in a list
CREATE INDEX idx_contest_participants_username ON contest_participants(username);
