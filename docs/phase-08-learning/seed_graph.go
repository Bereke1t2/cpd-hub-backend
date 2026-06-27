//go:build ignore

// Template for Phase 8 — reference snippet for: cmd/seed (topic graph + acyclicity)
//
// Author the curriculum as data, then assert it's a DAG before inserting so a bad
// prerequisite edge can't ship and hang the client's path engine.
//
package main

import "fmt"

// edge: topic -> its prerequisites
var topicPrereqs = map[string][]string{
	"implementation": {},
	"math-basics":    {},
	"sorting":        {"implementation"},
	"binary-search":  {"sorting"},
	"two-pointers":   {"sorting"},
	"prefix-sums":    {"implementation"},
	"greedy":         {"sorting"},
	"graphs-bfs-dfs": {"implementation"},
	"dp-intro":       {"math-basics", "greedy"},
	"dp-knapsack":    {"dp-intro"},
	"shortest-paths": {"graphs-bfs-dfs"},
	"segment-tree":   {"prefix-sums", "binary-search"},
}

// assertAcyclic topologically sorts the prerequisite graph; returns an error
// naming a node involved in a cycle. Call this from seed before inserting.
func assertAcyclic(prereqs map[string][]string) error {
	const (
		white = 0 // unvisited
		gray  = 1 // on the current DFS stack
		black = 2 // done
	)
	color := map[string]int{}
	for n := range prereqs {
		color[n] = white
	}

	var visit func(n string) error
	visit = func(n string) error {
		color[n] = gray
		for _, p := range prereqs[n] {
			switch color[p] {
			case gray:
				return fmt.Errorf("prerequisite cycle through %q -> %q", n, p)
			case white:
				if err := visit(p); err != nil {
					return err
				}
			}
		}
		color[n] = black
		return nil
	}

	for n := range prereqs {
		if color[n] == white {
			if err := visit(n); err != nil {
				return err
			}
		}
	}
	return nil
}

// Usage in cmd/seed/main.go:
//
//	if err := assertAcyclic(topicPrereqs); err != nil {
//		log.Fatalf("topic graph is not a DAG: %v", err)
//	}
//	// ... INSERT topics, then topic_prerequisites from the map, etc.
