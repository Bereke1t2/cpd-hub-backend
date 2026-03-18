-- Seed file for CPD Hub (Postgres)

-- Ensure auxiliary leaderboard table exists (migrations may not have created it)
CREATE TABLE IF NOT EXISTS contest_leaderboard (
  contest_id TEXT,
  rank INT,
  username TEXT,
  rating INT,
  score INT,
  penalty INT,
  problems_solved TEXT,
  PRIMARY KEY (contest_id, rank)
);

-- Users: note passwords are not stored here; use /api/auth/signup to create users with bcrypt.
INSERT INTO users (username, full_name, rating, bio, avatar_url) VALUES
('bereket', 'Bereket Lemma', 1750, 'Competitive programmer', 'https://example.com/avatar/bereket.png'),
('alice', 'Alice Johnson', 1600, 'Loves algorithms and puzzles', 'https://example.com/avatar/alice.png'),
('bob', 'Bob Smith', 1420, 'Backend engineer and problem solver', 'https://example.com/avatar/bob.png'),
('carol', 'Carol Nguyen', 1985, 'Speed coder', 'https://example.com/avatar/carol.png'),
('dave', 'Dave Lee', 1550, 'Math enthusiast', 'https://example.com/avatar/dave.png'),
('eve', 'Eve Turner', 1700, 'Open source contributor', 'https://example.com/avatar/eve.png'),
('frank', 'Frank Zhao', 1300, 'Learning CP', 'https://example.com/avatar/frank.png'),
('grace', 'Grace Park', 2100, 'Competitive programmer & mentor', 'https://example.com/avatar/grace.png'),
('heidi', 'Heidi Patel', 1800, 'Contest veteran', 'https://example.com/avatar/heidi.png'),
('ivan', 'Ivan Petrov', 1650, 'Algorithms fan', 'https://example.com/avatar/ivan.png'),
('judy', 'Judy Alvarez', 1480, 'Enjoys DP problems', 'https://example.com/avatar/judy.png'),
('karen', 'Karen O''Neill', 1525, 'Fullstack dev', 'https://example.com/avatar/karen.png'),
('leo', 'Leo Martinez', 1370, 'New to competitive programming', 'https://example.com/avatar/leo.png'),
('mia', 'Mia Chen', 1900, 'Problemsetter', 'https://example.com/avatar/mia.png'),
('nick', 'Nick Brown', 1200, 'Beginners', 'https://example.com/avatar/nick.png')
ON CONFLICT (username) DO NOTHING;

-- Profiles (lightweight) - mirror some users
INSERT INTO profiles (username, bio, rating, avatar_url) VALUES
('bereket', 'Competitive programmer', 1750, 'https://example.com/avatar/bereket.png'),
('alice', 'Loves algorithms and puzzles', 1600, 'https://example.com/avatar/alice.png'),
('carol', 'Speed coder', 1985, 'https://example.com/avatar/carol.png'),
('grace', 'Competitive programmer & mentor', 2100, 'https://example.com/avatar/grace.png'),
('mia', 'Problemsetter', 1900, 'https://example.com/avatar/mia.png')
ON CONFLICT (username) DO NOTHING;

-- Problems: add many seeded problems
INSERT INTO problems (id, title, difficulty, topic_tags, likes, dislikes, deep_link, is_liked, is_disliked, solved) VALUES
('p1','Two Sum','Easy','Array,Hash Table',245,12,'https://example.com/problems/p1',false,false,true),
('p2','Reverse Linked List','Easy','Linked List',180,5,'https://example.com/problems/p2',false,false,false),
('p3','Valid Parentheses','Easy','Stack,String',210,8,'https://example.com/problems/p3',false,false,false),
('p4','Merge Intervals','Medium','Interval,Sort',130,10,'https://example.com/problems/p4',false,false,false),
('p5','Longest Increasing Subsequence','Hard','DP,Binary Search',95,25,'https://example.com/problems/p5',false,false,false),
('p6','Median of Two Sorted Arrays','Hard','Divide and Conquer',60,15,'https://example.com/problems/p6',false,false,false),
('p7','Binary Tree Maximum Path Sum','Hard','Tree,DFS',88,20,'https://example.com/problems/p7',false,false,false),
('p8','Word Break','Medium','DP',150,9,'https://example.com/problems/p8',false,false,false),
('p9','Minimum Window Substring','Hard','Sliding Window,String',70,22,'https://example.com/problems/p9',false,false,false),
('p10','Clone Graph','Medium','Graph,DFS',55,4,'https://example.com/problems/p10',false,false,false),
('p11','Course Schedule','Medium','Graph,Topological Sort',99,6,'https://example.com/problems/p11',false,false,false),
('p12','Number of Islands','Medium','DFS,BFS',140,7,'https://example.com/problems/p12',false,false,false),
('p13','LRU Cache','Medium','Design,Hash Table',112,3,'https://example.com/problems/p13',false,false,false),
('p14','Kth Smallest Element in a BST','Medium','BST',85,2,'https://example.com/problems/p14',false,false,false),
('p15','Lowest Common Ancestor','Medium','Tree',128,11,'https://example.com/problems/p15',false,false,false),
('p16','Subarray Sum Equals K','Easy','Array,Hash Table',160,6,'https://example.com/problems/p16',false,false,false),
('p17','Sliding Window Maximum','Hard','Deque',72,9,'https://example.com/problems/p17',false,false,false),
('p18','Palindromic Substrings','Medium','DP,String',94,7,'https://example.com/problems/p18',false,false,false),
('p19','Find Median from Data Stream','Hard','Heap,Design',53,12,'https://example.com/problems/p19',false,false,false),
('p20','Permutation in String','Medium','Sliding Window,String',67,5,'https://example.com/problems/p20',false,false,false),
('dp1','Longest Common Subsequence','Medium','Dynamic Programming,String',342,18,'https://example.com/problems/dp1',false,false,false),
('p21','Graph Valid Tree','Medium','Graph',44,3,'https://example.com/problems/p21',false,false,false),
('p22','Implement Trie','Medium','Design,String',37,2,'https://example.com/problems/p22',false,false,false),
('p23','Design Twitter','Medium','Design',29,1,'https://example.com/problems/p23',false,false,false),
('p24','Word Ladder','Hard','BFS,Graph',41,8,'https://example.com/problems/p24',false,false,false),
('p25','Insert Interval','Medium','Interval',48,4,'https://example.com/problems/p25',false,false,false),
('p26','Find Peak Element','Medium','Array,Binary Search',66,2,'https://example.com/problems/p26',false,false,false),
('p27','Minimum Path Sum','Medium','DP',54,3,'https://example.com/problems/p27',false,false,false),
('p28','Unique Paths','Medium','DP',89,6,'https://example.com/problems/p28',false,false,false),
('p29','Combination Sum','Medium','Backtracking',73,5,'https://example.com/problems/p29',false,false,false),
('p30','House Robber','Easy','DP',120,8,'https://example.com/problems/p30',false,false,false),
('p31','Binary Search','Easy','Binary Search',88,2,'https://example.com/problems/p31',false,false,false),
('p32','Heapify','Medium','Heap',34,1,'https://example.com/problems/p32',false,false,false),
('p33','Top K Frequent Elements','Medium','Hash Table,Heap',77,5,'https://example.com/problems/p33',false,false,false),
('p34','KMP Algorithm','Hard','String',22,9,'https://example.com/problems/p34',false,false,false),
('p35','Bellman-Ford','Hard','Graph',11,4,'https://example.com/problems/p35',false,false,false)
ON CONFLICT (id) DO NOTHING;

-- Contests: add several contests
INSERT INTO contests (id, title, contest_url, start_time, duration, platform, number_of_problems, number_of_contestants, date, is_past, is_participating) VALUES
('c1','Global Round #26','https://example.com/contests/c1', NOW(), '2h 30m', 'CPD Hub', 6, 1240, 'Feb 15, 2026', false, true),
('c2','Weekly Challenge #45','https://example.com/contests/c2', NOW() - INTERVAL '7 days', '3h', 'CPD Hub', 8, 580, 'Mar 10, 2026', true, false),
('c3','Monthly Marathon','https://example.com/contests/c3', NOW() - INTERVAL '30 days', '24h', 'CPD Hub', 50, 2400, 'Feb 15, 2026', true, false),
('c4','Beginner Cup','https://example.com/contests/c4', NOW() + INTERVAL '2 days', '1h 30m', 'CPD Hub', 5, 300, 'Mar 20, 2026', false, false)
ON CONFLICT (id) DO NOTHING;

-- Contest leaderboard entries (store solved problems as comma-separated)
INSERT INTO contest_leaderboard (contest_id, rank, username, rating, score, penalty, problems_solved) VALUES
('c1',1,'grace',2400,600,30,'A,B,C,D,E,F'),
('c1',2,'carol',1985,580,45,'A,B,C,D,E'),
('c1',3,'mia',1900,560,60,'A,B,C,D'),
('c2',1,'alice',1600,420,90,'A,B,C'),
('c2',2,'heidi',1800,410,100,'A,B'),
('c3',1,'grace',2400,1200,300,'...')
ON CONFLICT (contest_id, rank) DO NOTHING;

-- Info / system messages
INSERT INTO info (title, description) VALUES
('System Maintenance', 'Scheduled maintenance on Feb 20th from 2-4 AM'),
('New Feature', 'Problem tagging and contest subscriptions added'),
('Privacy Update', 'Updated privacy policy — minor changes')
ON CONFLICT (title) DO NOTHING;

-- Activity feed (recent actions)
INSERT INTO activity (id, username, action, type, timestamp) VALUES
('a1','bereket','solved Two Sum','Solve','2026-02-15T10:00:00Z'),
('a2','alice','liked Two Sum','Like','2026-03-01T08:30:00Z'),
('a3','carol','participated in Global Round #26','Contest','2026-02-15T12:00:00Z'),
('a4','grace','set a new problem','Admin','2026-03-05T14:00:00Z'),
('a5','mia','edited problem p5','Edit','2026-02-20T09:00:00Z'),
('a6','bob','solved Reverse Linked List','Solve','2026-03-02T11:00:00Z'),
('a7','dave','commented on p4','Comment','2026-03-02T11:05:00Z'),
('a8','eve','added a new solution to p8','Solve','2026-03-03T09:00:00Z'),
('a9','ivan','improved rating to 1650','Update','2026-03-04T16:00:00Z'),
('a10','judy','bookmarked p13','Bookmark','2026-03-04T17:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- End of seed
