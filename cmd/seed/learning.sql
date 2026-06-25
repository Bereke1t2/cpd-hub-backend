-- ------------------------------------------------------------------
-- Learning: Topics
-- ------------------------------------------------------------------
INSERT INTO topics (id, name, category, summary, difficulty) VALUES
  ('implementation', 'Implementation Basics', 'Core',    'Turning ideas into code efficiently.', 1),
  ('math-basics',    'Competitive Math',    'Math',    'Number theory, modular arithmetic basics.', 1),
  ('sorting',        'Sorting & Searching', 'Search',  'Sorting algorithms and binary search.', 1),
  ('binary-search',  'Binary Search Magic', 'Search',  'Beyond finding an element: searching on answers.', 2),
  ('two-pointers',   'Two Pointers',        'Linear',  'Efficiently processing ranges and pairs.', 2),
  ('prefix-sums',    'Prefix Sums',         'Core',    'Range sum queries in O(1).', 2),
  ('greedy',         'Greedy Thinking',     'Logic',   'Making the locally optimal choice.', 2),
  ('graphs-bfs-dfs', 'Graph Traversals',    'Graphs',  'BFS and DFS basics.', 2),
  ('dp-intro',       'Intro to DP',         'DP',      'Breaking problems into subproblems.', 3),
  ('dp-knapsack',    'Knapsack & Beyond',   'DP',      'Classical DP optimizations.', 4),
  ('shortest-paths', 'Shortest Paths',      'Graphs',  'Dijkstra, Bellman-Ford, Floyd-Warshall.', 3),
  ('segment-tree',   'Segment Trees',       'Trees',   'Powerful range query data structure.', 4)
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------------
-- Learning: Prerequisites (The DAG)
-- ------------------------------------------------------------------
INSERT INTO topic_prerequisites (topic_id, prerequisite_id) VALUES
  ('sorting',        'implementation'),
  ('binary-search',  'sorting'),
  ('two-pointers',   'sorting'),
  ('prefix-sums',    'implementation'),
  ('greedy',         'sorting'),
  ('graphs-bfs-dfs', 'implementation'),
  ('dp-intro',       'math-basics'),
  ('dp-intro',       'greedy'),
  ('dp-knapsack',    'dp-intro'),
  ('shortest-paths', 'graphs-bfs-dfs'),
  ('segment-tree',   'prefix-sums'),
  ('segment-tree',   'binary-search')
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------------
-- Learning: Topic-Problem Links (Linking to seed problems)
-- ------------------------------------------------------------------
INSERT INTO topic_problems (topic_id, problem_id) VALUES
  ('implementation', 'p1'),
  ('sorting',        'p2'),
  ('prefix-sums',    'p3'),
  ('dp-intro',       'p4'),
  ('graphs-bfs-dfs', 'p5')
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------------
-- Learning: Tracks
-- ------------------------------------------------------------------
INSERT INTO tracks (id, title, description, icon_name) VALUES
  ('foundations',  'CP Foundations',      'Every journey starts with implementation and basics.', 'school'),
  ('intermediate', 'Greedy & Search',     'Moving beyond brute force.', 'bolt'),
  ('advanced',     'Advanced Algorithms', 'DP and complex data structures.', 'psychology')
ON CONFLICT (id) DO NOTHING;

INSERT INTO track_topics (track_id, topic_id, ord) VALUES
  ('foundations',  'implementation', 1),
  ('foundations',  'math-basics',    2),
  ('foundations',  'sorting',        3),
  ('foundations',  'prefix-sums',    4),
  ('intermediate', 'binary-search',  1),
  ('intermediate', 'two-pointers',   2),
  ('intermediate', 'greedy',         3),
  ('intermediate', 'graphs-bfs-dfs', 4),
  ('advanced',     'dp-intro',       1),
  ('advanced',     'dp-knapsack',    2),
  ('advanced',     'shortest-paths', 3),
  ('advanced',     'segment-tree',   4)
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------------
-- Learning: Lessons
-- ------------------------------------------------------------------
INSERT INTO lessons (topic_id, body, key_ideas) VALUES
  ('implementation', 'Content for implementation...', ARRAY['Think before you type', 'Handle edge cases', 'Write clean code']),
  ('prefix-sums',    'Pre-calculating range sums...', ARRAY['O(1) range queries', 'Preprocessing in O(N)', 'Can be extended to 2D']),
  ('dp-intro',       'Dynamic Programming is about...', ARRAY['Overlapping subproblems', 'Optimal substructure', 'Memoization vs Tabulation'])
ON CONFLICT (topic_id) DO NOTHING;
