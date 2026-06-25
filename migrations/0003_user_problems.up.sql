CREATE TABLE IF NOT EXISTS user_problems (
    username    TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    problem_id  TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    liked       BOOLEAN NOT NULL DEFAULT false,
    disliked    BOOLEAN NOT NULL DEFAULT false,
    solved      BOOLEAN NOT NULL DEFAULT false,
    solved_at   TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (username, problem_id)
);

CREATE INDEX IF NOT EXISTS idx_user_problems_username ON user_problems(username);
CREATE INDEX IF NOT EXISTS idx_user_problems_problem ON user_problems(problem_id);
CREATE INDEX IF NOT EXISTS idx_user_problems_solved ON user_problems(problem_id) WHERE solved;
