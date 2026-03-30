-- filepath: internal/infrastructure/postgres/seed_additions.sql
-- Additional seed and DDL for attendance, heatmap, submissions and contest registrants

-- Create auxiliary tables if they don't exist
CREATE TABLE IF NOT EXISTS attendance (
  id TEXT PRIMARY KEY,
  username TEXT,
  date DATE,
  status TEXT
);

CREATE TABLE IF NOT EXISTS heatmap (
  id TEXT PRIMARY KEY,
  username TEXT,
  date DATE,
  solve_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS submissions (
  id TEXT PRIMARY KEY,
  username TEXT,
  problem_id TEXT,
  problem_title TEXT,
  status TEXT,
  language TEXT,
  execution_time TEXT,
  memory_used TEXT,
  timestamp TIMESTAMP
);

CREATE TABLE IF NOT EXISTS contest_registrants (
  contest_id TEXT PRIMARY KEY,
  registrant_count INT
);

-- Sample attendance data
INSERT INTO attendance (id, username, date, status) VALUES
('att1','alice','2026-03-01','Present'),
('att2','bob','2026-03-01','Absent'),
('att3','carol','2026-03-02','Present')
ON CONFLICT (id) DO NOTHING;

-- Sample heatmap data
INSERT INTO heatmap (id, username, date, solve_count) VALUES
('h1','alice','2026-03-01',2),
('h2','alice','2026-03-02',1),
('h3','bereket','2026-03-02',3),
('h4','grace','2026-03-01',5)
ON CONFLICT (id) DO NOTHING;

-- Sample submissions
INSERT INTO submissions (id, username, problem_id, problem_title, status, language, execution_time, memory_used, timestamp) VALUES
('s1','alice','p1','Two Sum','Accepted','Python','45ms','14.2MB','2026-03-01 08:30:00'),
('s2','bob','p2','Reverse Linked List','Wrong Answer','C++','120ms','25.0MB','2026-03-02 11:00:00'),
('s3','carol','p5','Longest Increasing Subsequence','Accepted','Java','200ms','40MB','2026-02-15 12:30:00')
ON CONFLICT (id) DO NOTHING;

-- Sample contest registrant counts
INSERT INTO contest_registrants (contest_id, registrant_count) VALUES
('c4', 320),
('codeforces-1932', 1500)
ON CONFLICT (contest_id) DO NOTHING;
