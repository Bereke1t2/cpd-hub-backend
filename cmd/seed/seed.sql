-- =============================================================
-- CPD Hub — dev seed (runs after migrations/0001-0004)
-- Rules:
--   • NEVER insert plaintext passwords.
--     Create users with real passwords via POST /api/auth/signup.
--   • Data must conform to the schema produced by migrations/0001-0004.
-- =============================================================

-- ------------------------------------------------------------------
-- Users  (password_hash left empty; use signup endpoint for real auth)
-- ------------------------------------------------------------------
INSERT INTO users (username, full_name, rating, bio, avatar_url) VALUES
  ('bereket', 'Bereket Lemma',  1750, 'Competitive programmer',          'https://example.com/avatar/bereket.png'),
  ('alice',   'Alice Johnson',  1600, 'Loves algorithms and puzzles',    'https://example.com/avatar/alice.png'),
  ('bob',     'Bob Smith',      1420, 'Backend engineer',                'https://example.com/avatar/bob.png'),
  ('carol',   'Carol Nguyen',   1985, 'Speed coder',                     'https://example.com/avatar/carol.png'),
  ('dave',    'Dave Lee',       1550, 'Math enthusiast',                 'https://example.com/avatar/dave.png'),
  ('eve',     'Eve Turner',     1700, 'Open source contributor',         'https://example.com/avatar/eve.png'),
  ('frank',   'Frank Zhao',     1300, 'Learning CP',                     'https://example.com/avatar/frank.png'),
  ('grace',   'Grace Park',     2100, 'Competitive programmer & mentor', 'https://example.com/avatar/grace.png'),
  ('heidi',   'Heidi Patel',    1800, 'Contest veteran',                 'https://example.com/avatar/heidi.png'),
  ('ivan',    'Ivan Petrov',    1650, 'Algorithms fan',                  'https://example.com/avatar/ivan.png'),
  ('judy',    'Judy Alvarez',   1480, 'Enjoys DP problems',              'https://example.com/avatar/judy.png'),
  ('karen',   'Karen O''Neill', 1525, 'Fullstack dev',                   'https://example.com/avatar/karen.png'),
  ('leo',     'Leo Martinez',   1370, 'New to competitive programming',  'https://example.com/avatar/leo.png'),
  ('mia',     'Mia Chen',       1900, 'Problemsetter',                   'https://example.com/avatar/mia.png'),
  ('nick',    'Nick Brown',     1200, 'Beginners',                       'https://example.com/avatar/nick.png')
ON CONFLICT (username) DO NOTHING;

-- ------------------------------------------------------------------
-- Profiles
-- ------------------------------------------------------------------
INSERT INTO profiles (username, bio, rating, avatar_url) VALUES
  ('bereket', 'Competitive programmer',          1750, 'https://example.com/avatar/bereket.png'),
  ('alice',   'Loves algorithms and puzzles',    1600, 'https://example.com/avatar/alice.png'),
  ('carol',   'Speed coder',                     1985, 'https://example.com/avatar/carol.png'),
  ('grace',   'Competitive programmer & mentor', 2100, 'https://example.com/avatar/grace.png'),
  ('mia',     'Problemsetter',                   1900, 'https://example.com/avatar/mia.png')
ON CONFLICT (username) DO NOTHING;

-- ------------------------------------------------------------------
-- Problems
-- ------------------------------------------------------------------
INSERT INTO problems (id, title, difficulty, topic_tags, likes, dislikes, deep_link) VALUES
  ('p1',  'Two Sum',                          'Easy',   'Array,Hash Table',          245, 12, 'https://example.com/problems/p1'),
  ('p2',  'Reverse Linked List',              'Easy',   'Linked List',               180,  5, 'https://example.com/problems/p2'),
  ('p3',  'Valid Parentheses',                'Easy',   'Stack,String',              210,  8, 'https://example.com/problems/p3'),
  ('p4',  'Merge Intervals',                  'Medium', 'Interval,Sort',             130, 10, 'https://example.com/problems/p4'),
  ('p5',  'Longest Increasing Subsequence',   'Hard',   'DP,Binary Search',           95, 25, 'https://example.com/problems/p5'),
  ('p6',  'Median of Two Sorted Arrays',      'Hard',   'Divide and Conquer',         60, 15, 'https://example.com/problems/p6'),
  ('p7',  'Binary Tree Maximum Path Sum',     'Hard',   'Tree,DFS',                   88, 20, 'https://example.com/problems/p7'),
  ('p8',  'Word Break',                       'Medium', 'DP',                        150,  9, 'https://example.com/problems/p8'),
  ('p9',  'Minimum Window Substring',         'Hard',   'Sliding Window,String',      70, 22, 'https://example.com/problems/p9'),
  ('p10', 'Clone Graph',                      'Medium', 'Graph,DFS',                  55,  4, 'https://example.com/problems/p10'),
  ('p11', 'Course Schedule',                  'Medium', 'Graph,Topological Sort',     99,  6, 'https://example.com/problems/p11'),
  ('p12', 'Number of Islands',                'Medium', 'DFS,BFS',                   140,  7, 'https://example.com/problems/p12'),
  ('p13', 'LRU Cache',                        'Medium', 'Design,Hash Table',         112,  3, 'https://example.com/problems/p13'),
  ('p14', 'Kth Smallest Element in a BST',    'Medium', 'BST',                        85,  2, 'https://example.com/problems/p14'),
  ('p15', 'Lowest Common Ancestor',           'Medium', 'Tree',                      128, 11, 'https://example.com/problems/p15'),
  ('p16', 'Subarray Sum Equals K',            'Easy',   'Array,Hash Table',          160,  6, 'https://example.com/problems/p16'),
  ('p17', 'Sliding Window Maximum',           'Hard',   'Deque',                      72,  9, 'https://example.com/problems/p17'),
  ('p18', 'Palindromic Substrings',           'Medium', 'DP,String',                  94,  7, 'https://example.com/problems/p18'),
  ('p19', 'Find Median from Data Stream',     'Hard',   'Heap,Design',                53, 12, 'https://example.com/problems/p19'),
  ('p20', 'Permutation in String',            'Medium', 'Sliding Window,String',      67,  5, 'https://example.com/problems/p20'),
  ('dp1', 'Longest Common Subsequence',       'Medium', 'Dynamic Programming,String',342, 18, 'https://example.com/problems/dp1'),
  ('p21', 'Graph Valid Tree',                 'Medium', 'Graph',                      44,  3, 'https://example.com/problems/p21'),
  ('p22', 'Implement Trie',                   'Medium', 'Design,String',              37,  2, 'https://example.com/problems/p22'),
  ('p23', 'Design Twitter',                   'Medium', 'Design',                     29,  1, 'https://example.com/problems/p23'),
  ('p24', 'Word Ladder',                      'Hard',   'BFS,Graph',                  41,  8, 'https://example.com/problems/p24'),
  ('p25', 'Insert Interval',                  'Medium', 'Interval',                   48,  4, 'https://example.com/problems/p25'),
  ('p26', 'Find Peak Element',                'Medium', 'Array,Binary Search',        66,  2, 'https://example.com/problems/p26'),
  ('p27', 'Minimum Path Sum',                 'Medium', 'DP',                         54,  3, 'https://example.com/problems/p27'),
  ('p28', 'Unique Paths',                     'Medium', 'DP',                         89,  6, 'https://example.com/problems/p28'),
  ('p29', 'Combination Sum',                  'Medium', 'Backtracking',               73,  5, 'https://example.com/problems/p29'),
  ('p30', 'House Robber',                     'Easy',   'DP',                        120,  8, 'https://example.com/problems/p30'),
  ('p31', 'Binary Search',                    'Easy',   'Binary Search',              88,  2, 'https://example.com/problems/p31'),
  ('p32', 'Heapify',                          'Medium', 'Heap',                       34,  1, 'https://example.com/problems/p32'),
  ('p33', 'Top K Frequent Elements',          'Medium', 'Hash Table,Heap',            77,  5, 'https://example.com/problems/p33'),
  ('p34', 'KMP Algorithm',                    'Hard',   'String',                     22,  9, 'https://example.com/problems/p34'),
  ('p35', 'Bellman-Ford',                     'Hard',   'Graph',                      11,  4, 'https://example.com/problems/p35')
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------------
-- Contests
-- ------------------------------------------------------------------
INSERT INTO contests (id, title, contest_url, start_time, duration, platform,
                      number_of_problems, number_of_contestants, date, is_past, is_participating) VALUES
  ('c1', 'Global Round #26',    'https://example.com/contests/c1', NOW(),                       '2h 30m', 'CPD Hub', 6,  1240, 'Feb 15, 2026', false, true),
  ('c2', 'Weekly Challenge #45','https://example.com/contests/c2', NOW() - INTERVAL '7 days',  '3h',     'CPD Hub', 8,   580, 'Mar 10, 2026', true,  false),
  ('c3', 'Monthly Marathon',    'https://example.com/contests/c3', NOW() - INTERVAL '30 days', '24h',    'CPD Hub', 50, 2400, 'Feb 15, 2026', true,  false),
  ('c4', 'Beginner Cup',        'https://example.com/contests/c4', NOW() + INTERVAL '2 days',  '1h 30m', 'CPD Hub', 5,   300, 'Mar 20, 2026', false, false)
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------------
-- Contest leaderboard
-- ------------------------------------------------------------------
INSERT INTO contest_leaderboard (contest_id, rank, username, rating, score, penalty, problems_solved) VALUES
  ('c1', 1, 'grace', 2400,  600,  30, 'A,B,C,D,E,F'),
  ('c1', 2, 'carol', 1985,  580,  45, 'A,B,C,D,E'),
  ('c1', 3, 'mia',   1900,  560,  60, 'A,B,C,D'),
  ('c2', 1, 'alice', 1600,  420,  90, 'A,B,C'),
  ('c2', 2, 'heidi', 1800,  410, 100, 'A,B'),
  ('c3', 1, 'grace', 2400, 1200, 300, '...')
ON CONFLICT (contest_id, rank) DO NOTHING;

-- ------------------------------------------------------------------
-- Info / system messages
-- ------------------------------------------------------------------
INSERT INTO info (title, description) VALUES
  ('System Maintenance', 'Scheduled maintenance on Feb 20th from 2-4 AM'),
  ('New Feature',        'Problem tagging and contest subscriptions added'),
  ('Privacy Update',     'Updated privacy policy — minor changes'),
  ('CPD Tutorial Update','Today''s tutorial has been postponed to next weekend!')
ON CONFLICT (title) DO NOTHING;

-- ------------------------------------------------------------------
-- Activity feed
-- ------------------------------------------------------------------
INSERT INTO activity (id, username, action, type, timestamp) VALUES
  ('a1',  'bereket', 'solved Two Sum',                   'Solve',    '2026-02-15T10:00:00Z'),
  ('a2',  'alice',   'liked Two Sum',                    'Like',     '2026-03-01T08:30:00Z'),
  ('a3',  'carol',   'participated in Global Round #26', 'Contest',  '2026-02-15T12:00:00Z'),
  ('a4',  'grace',   'set a new problem',                'Admin',    '2026-03-05T14:00:00Z'),
  ('a5',  'mia',     'edited problem p5',                'Edit',     '2026-02-20T09:00:00Z'),
  ('a6',  'bob',     'solved Reverse Linked List',       'Solve',    '2026-03-02T11:00:00Z'),
  ('a7',  'dave',    'commented on p4',                  'Comment',  '2026-03-02T11:05:00Z'),
  ('a8',  'eve',     'added a new solution to p8',       'Solve',    '2026-03-03T09:00:00Z'),
  ('a9',  'ivan',    'improved rating to 1650',          'Update',   '2026-03-04T16:00:00Z'),
  ('a10', 'judy',    'bookmarked p13',                   'Bookmark', '2026-03-04T17:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------------
-- Analytics: daily_solves  (heatmap source)
-- ------------------------------------------------------------------
INSERT INTO daily_solves (username, day, count) VALUES
  ('alice',   '2026-03-01', 2),
  ('alice',   '2026-03-02', 1),
  ('bereket', '2026-03-02', 3),
  ('grace',   '2026-03-01', 5),
  ('carol',   '2026-02-15', 4),
  ('bob',     '2026-03-02', 1)
ON CONFLICT (username, day) DO UPDATE SET count = EXCLUDED.count;

-- ------------------------------------------------------------------
-- Analytics: rating_history
-- ------------------------------------------------------------------
INSERT INTO rating_history (username, day, rating) VALUES
  ('bereket', '2026-01-01', 1600),
  ('bereket', '2026-02-01', 1680),
  ('bereket', '2026-03-01', 1750),
  ('alice',   '2026-01-01', 1450),
  ('alice',   '2026-02-01', 1520),
  ('alice',   '2026-03-01', 1600),
  ('grace',   '2026-01-01', 1950),
  ('grace',   '2026-02-01', 2020),
  ('grace',   '2026-03-01', 2100)
ON CONFLICT (username, day) DO NOTHING;

-- ------------------------------------------------------------------
-- Analytics: attendance
-- ------------------------------------------------------------------
INSERT INTO attendance (username, day, status) VALUES
  ('alice',   '2026-03-01', 'Present'),
  ('bob',     '2026-03-01', 'Absent'),
  ('carol',   '2026-03-02', 'Present'),
  ('bereket', '2026-03-01', 'Present'),
  ('bereket', '2026-03-02', 'Present')
ON CONFLICT (username, day) DO NOTHING;

-- ------------------------------------------------------------------
-- Analytics: submissions
-- ------------------------------------------------------------------
INSERT INTO submissions (id, username, problem_id, problem_title, status, language, execution_time, memory_used) VALUES
  ('s1', 'alice',   'p1', 'Two Sum',                        'Accepted',     'Python', '45ms',  '14.2MB'),
  ('s2', 'bob',     'p2', 'Reverse Linked List',            'Wrong Answer', 'C++',   '120ms', '25.0MB'),
  ('s3', 'carol',   'p5', 'Longest Increasing Subsequence', 'Accepted',     'Java',  '200ms', '40MB'),
  ('s4', 'bereket', 'p1', 'Two Sum',                        'Accepted',     'Go',    '32ms',  '12.1MB')
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------------
-- Consistency: Ladders
-- ------------------------------------------------------------------
INSERT INTO ladders (id, title, from_rating, to_rating) VALUES
  ('ladder-1200', 'Div. 2 A Ladder', 0, 1200),
  ('ladder-1400', 'Div. 2 B Ladder', 1201, 1400)
ON CONFLICT (id) DO NOTHING;

INSERT INTO ladder_rungs (ladder_id, problem_id, rating, topic_id, ord) VALUES
  ('ladder-1200', 'p1',  800,  'implementation', 1),
  ('ladder-1200', 'p2',  900,  'math',           2),
  ('ladder-1200', 'p3',  1000, 'greedy',         3),
  ('ladder-1400', 'p4',  1300, 'dp',             1),
  ('ladder-1400', 'p5',  1400, 'graphs',         2)
ON CONFLICT (ladder_id, problem_id) DO NOTHING;
