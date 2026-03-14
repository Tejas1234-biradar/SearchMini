// Package pagerank implements an iterative PageRank batch job.
// Run standalone: go run pagerank/pagerank.go
// It reads from the `outlinks` collection and writes scores to `pagerank`.
package pagerank

// TODO: implement iterative PageRank
// 1. Read all outlinks documents from MongoDB
// 2. Build adjacency graph: map[url][]outlinks
// 3. Run N iterations with damping factor 0.85
// 4. Write { _id: url, score: float64 } to `pagerank` collection
