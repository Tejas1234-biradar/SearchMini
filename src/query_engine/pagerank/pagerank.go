package pagerank

import (
	"context"
	"fmt"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultDampingFactor = 0.85
	defaultIterations    = 20
	outlinksCollection   = "outlinks"
	pageRankCollection   = "pagerank"
)

// Config controls PageRank convergence behavior.
type Config struct {
	// DampingFactor is usually 0.85. Values outside (0,1) fallback to default.
	DampingFactor float64
	// Iterations is the number of full update passes. Values <= 0 fallback to default.
	Iterations int
}

// DefaultConfig returns recommended defaults for batch PageRank.
func DefaultConfig() Config {
	return Config{
		DampingFactor: defaultDampingFactor,
		Iterations:    defaultIterations,
	}
}

type outlinksDoc struct {
	ID    string   `bson:"_id"`
	Links []string `bson:"links"`
}

type pageRankDoc struct {
	ID    string  `bson:"_id"`
	Score float64 `bson:"score"`
}

// BuildGraphFromMongo reads all outlinks and builds an adjacency list.
// It includes destination-only pages as graph nodes with zero out-degree.
func BuildGraphFromMongo(ctx context.Context, db *mongo.Database) (map[string][]string, error) {
	cursor, err := db.Collection(outlinksCollection).Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("query outlinks collection: %w", err)
	}
	defer cursor.Close(ctx)

	graph := make(map[string][]string)
	for cursor.Next(ctx) {
		var doc outlinksDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode outlinks document: %w", err)
		}

		normalized := uniqueNonEmpty(doc.Links)
		graph[doc.ID] = normalized
		for _, link := range normalized {
			if _, exists := graph[link]; !exists {
				graph[link] = []string{}
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate outlinks cursor: %w", err)
	}

	return graph, nil
}

// Compute iteratively calculates PageRank values for the provided graph.
func Compute(graph map[string][]string, cfg Config) map[string]float64 {
	if len(graph) == 0 {
		return map[string]float64{}
	}

	damping := cfg.DampingFactor
	if damping <= 0 || damping >= 1 {
		damping = defaultDampingFactor
	}
	iterations := cfg.Iterations
	if iterations <= 0 {
		iterations = defaultIterations
	}

	nodes := make([]string, 0, len(graph))
	for url := range graph {
		nodes = append(nodes, url)
	}
	sort.Strings(nodes)

	n := float64(len(nodes))
	ranks := make(map[string]float64, len(nodes))
	for _, url := range nodes {
		ranks[url] = 1.0 / n
	}

	for i := 0; i < iterations; i++ {
		next := make(map[string]float64, len(nodes))

		danglingMass := 0.0
		for _, src := range nodes {
			if len(graph[src]) == 0 {
				danglingMass += ranks[src]
			}
		}

		base := (1.0-damping)/n + damping*danglingMass/n
		for _, url := range nodes {
			next[url] = base
		}

		for _, src := range nodes {
			targets := graph[src]
			if len(targets) == 0 {
				continue
			}

			share := damping * ranks[src] / float64(len(targets))
			for _, dst := range targets {
				next[dst] += share
			}
		}

		ranks = next
	}

	return ranks
}

// Save writes rank results to the pagerank collection as {_id: url, score: ...}.
// Existing scores are replaced to keep the collection in sync with the latest run.
func Save(ctx context.Context, db *mongo.Database, ranks map[string]float64) error {
	col := db.Collection(pageRankCollection)

	if _, err := col.DeleteMany(ctx, bson.D{}); err != nil {
		return fmt.Errorf("clear pagerank collection: %w", err)
	}
	if len(ranks) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(ranks))
	for url, score := range ranks {
		models = append(models, mongo.NewReplaceOneModel().
			SetFilter(bson.D{{Key: "_id", Value: url}}).
			SetReplacement(pageRankDoc{ID: url, Score: score}).
			SetUpsert(true),
		)
	}

	_, err := col.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return fmt.Errorf("bulk write pagerank: %w", err)
	}
	return nil
}

// Run executes a full PageRank batch pass from `outlinks` to `pagerank`.
func Run(ctx context.Context, db *mongo.Database, cfg Config) error {
	graph, err := BuildGraphFromMongo(ctx, db)
	if err != nil {
		return err
	}
	ranks := Compute(graph, cfg)
	if err := Save(ctx, db, ranks); err != nil {
		return err
	}
	return nil
}

func uniqueNonEmpty(links []string) []string {
	if len(links) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(links))
	result := make([]string, 0, len(links))
	for _, link := range links {
		if link == "" {
			continue
		}
		if _, exists := seen[link]; exists {
			continue
		}
		seen[link] = struct{}{}
		result = append(result, link)
	}
	return result
}
