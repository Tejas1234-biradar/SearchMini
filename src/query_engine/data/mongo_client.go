package data

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/schemas"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	wordsCollection     = "words"
	metadataCollection  = "metadata"
	outlinksCollection  = "outlinks"
	backlinksCollection = "backlinks"
	pageRankCollection  = "pagerank"
)

// MongoClient wraps the MongoDB driver for query engine operations.
type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
}

// Database exposes the underlying database handle for read/write operations
// needed by other query engine modules (e.g., PageRank computation).
func (m *MongoClient) Database() *mongo.Database {
	return m.db
}

func NewMongoClient(host, username, password, dbName string, port int) (*MongoClient, error) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		username, password, host, port, dbName,
	)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	slog.Info("Successfully connected to mongo")

	return &MongoClient{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

func (m *MongoClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// ---------------------------------------------------------------------------
// SearchPages
// ---------------------------------------------------------------------------

// SearchPages returns paginated search results for the given query words,
// sorted by a combined score of TF-IDF (60%) and PageRank (40%).
//
// Pipeline (mirrors Moogle's QuerySearchController::search):
//  1. $match   — keep only documents whose `word` is in the query words list
//  2. $group   — group by `url`, accumulating cumulative TF-IDF weight and match count
//  3. $sort    — sort by matchCount desc, then cumWeight desc
//  4. $skip/$limit — paginate
//
// After aggregation we enrich each result with metadata and PageRank from
// their respective collections.
func (m *MongoClient) SearchPages(ctx context.Context, words []string, page, perPage int) ([]schemas.SearchResult, int, error) {
	col := m.db.Collection(wordsCollection)

	// ---- 1. Count total matching URLs (for pagination UI) ----
	countPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "word", Value: bson.D{{Key: "$in", Value: words}}}}}},
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$url"}}}},
		{{Key: "$count", Value: "total"}},
	}

	countCursor, err := col.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, fmt.Errorf("count aggregation failed: %w", err)
	}
	defer countCursor.Close(ctx)

	total := 0
	if countCursor.Next(ctx) {
		var countResult struct {
			Total int `bson:"total"`
		}
		if err := countCursor.Decode(&countResult); err == nil {
			total = countResult.Total
		}
	}

	if total == 0 {
		return []schemas.SearchResult{}, 0, nil
	}

	// ---- 2. Paginated aggregation ----
	skip := int64((page - 1) * perPage)
	limit := int64(perPage)

	searchPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "word", Value: bson.D{{Key: "$in", Value: words}}}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$url"},
			{Key: "cumWeight", Value: bson.D{{Key: "$sum", Value: "$tfidf"}}},
			{Key: "matchCount", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "matchCount", Value: -1},
			{Key: "cumWeight", Value: -1},
		}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := col.Aggregate(ctx, searchPipeline)
	if err != nil {
		return nil, 0, fmt.Errorf("search aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	type rawResult struct {
		URL        string  `bson:"_id"`
		CumWeight  float64 `bson:"cumWeight"`
		MatchCount int     `bson:"matchCount"`
	}

	var raws []rawResult
	if err := cursor.All(ctx, &raws); err != nil {
		return nil, 0, fmt.Errorf("decoding search results: %w", err)
	}

	// ---- 3. Collect URLs for batch lookups ----
	urls := make([]string, len(raws))
	for i, r := range raws {
		urls[i] = r.URL
	}

	// ---- 4. Fetch metadata & PageRank in batch ----
	metaMap, err := m.GetMetadataBatch(ctx, urls)
	if err != nil {
		slog.Warn("metadata batch lookup failed", "err", err)
		metaMap = map[string]schemas.Metadata{}
	}

	prMap, err := m.getPageRankBatch(ctx, urls)
	if err != nil {
		slog.Warn("pagerank batch lookup failed", "err", err)
		prMap = map[string]float64{}
	}

	// ---- 5. Merge and compute combined score ----
	results := make([]schemas.SearchResult, 0, len(raws))
	for _, r := range raws {
		meta := metaMap[r.URL]
		pr := prMap[r.URL]

		// 60% TF-IDF cumulative weight, 40% PageRank
		score := 0.6*r.CumWeight + 0.4*pr

		results = append(results, schemas.SearchResult{
			URL:         r.URL,
			Title:       meta.Title,
			Description: meta.Description,
			SummaryText: meta.SummaryText,
			LastCrawled: meta.LastCrawled,
			TFIDFWeight: r.CumWeight,
			PageRank:    pr,
			Score:       score,
		})
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// GetMetadataBatch
// ---------------------------------------------------------------------------

// GetMetadataBatch fetches metadata documents for a slice of URLs and returns
// them as a map keyed by URL for O(1) lookups.
func (m *MongoClient) GetMetadataBatch(ctx context.Context, urls []string) (map[string]schemas.Metadata, error) {
	col := m.db.Collection(metadataCollection)

	cursor, err := col.Find(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: urls}}}})
	if err != nil {
		return nil, fmt.Errorf("metadata find failed: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]schemas.Metadata, len(urls))
	for cursor.Next(ctx) {
		var meta schemas.Metadata
		if err := cursor.Decode(&meta); err != nil {
			continue
		}
		result[meta.ID] = meta
	}

	return result, cursor.Err()
}

// ---------------------------------------------------------------------------
// GetPageConnections
// ---------------------------------------------------------------------------

// GetPageConnections returns the outlinks and backlinks for the given URL,
// each enriched with the page title from the metadata collection.
func (m *MongoClient) GetPageConnections(ctx context.Context, url string) (outlinks, backlinks []schemas.PageConnection, err error) {
	outlinks, err = m.fetchLinks(ctx, outlinksCollection, url)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching outlinks: %w", err)
	}

	backlinks, err = m.fetchLinks(ctx, backlinksCollection, url)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching backlinks: %w", err)
	}

	return outlinks, backlinks, nil
}

// fetchLinks retrieves link URLs from the given collection for a page, then
// enriches them with titles from the metadata collection.
func (m *MongoClient) fetchLinks(ctx context.Context, collectionName, url string) ([]schemas.PageConnection, error) {
	// Outlinks/backlinks documents: { _id: url, links: [url1, url2, ...] }
	var doc struct {
		Links []string `bson:"links"`
	}

	err := m.db.Collection(collectionName).
		FindOne(ctx, bson.D{{Key: "_id", Value: url}}).
		Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return []schemas.PageConnection{}, nil
	}
	if err != nil {
		return nil, err
	}

	if len(doc.Links) == 0 {
		return []schemas.PageConnection{}, nil
	}

	// Batch fetch titles
	metaMap, err := m.GetMetadataBatch(ctx, doc.Links)
	if err != nil {
		slog.Warn("could not fetch link metadata", "err", err)
		metaMap = map[string]schemas.Metadata{}
	}

	connections := make([]schemas.PageConnection, 0, len(doc.Links))
	for _, link := range doc.Links {
		title := "Page Not Indexed"
		if meta, ok := metaMap[link]; ok && meta.Title != "" {
			title = meta.Title
		}
		connections = append(connections, schemas.PageConnection{URL: link, Title: title})
	}

	return connections, nil
}

// ---------------------------------------------------------------------------
// GetStats
// ---------------------------------------------------------------------------

// GetStats returns the total number of indexed pages in the metadata collection.
func (m *MongoClient) GetStats(ctx context.Context) (int64, error) {
	count, err := m.db.Collection(metadataCollection).CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("count documents failed: %w", err)
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// GetTopRankedPages
// ---------------------------------------------------------------------------

// GetTopRankedPages returns the top `limit` pages ranked by their combined
// TF-IDF weight across all words (i.e. sum of all word weights per URL).
//
// Pipeline:
//  1. $group  — sum all `tfidf` values per URL to get a total weight
//  2. $sort   — descending by total weight
//  3. $limit  — keep only the top N
//  4. $lookup — join metadata for titles
func (m *MongoClient) GetTopRankedPages(ctx context.Context, limit int) ([]schemas.PageRankResult, error) {
	col := m.db.Collection(wordsCollection)

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$url"},
			{Key: "score", Value: bson.D{{Key: "$sum", Value: "$tfidf"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "score", Value: -1}}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("top-pages aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []schemas.PageRankResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decoding top-pages: %w", err)
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// getPageRankBatch fetches pre-computed PageRank scores from the `pagerank`
// collection for the given URLs. Returns a zero score for any missing URL.
func (m *MongoClient) getPageRankBatch(ctx context.Context, urls []string) (map[string]float64, error) {
	col := m.db.Collection(pageRankCollection)

	cursor, err := col.Find(
		ctx,
		bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: urls}}}},
		options.Find().SetProjection(bson.D{{Key: "score", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]float64, len(urls))
	for cursor.Next(ctx) {
		var pr schemas.PageRankResult
		if err := cursor.Decode(&pr); err != nil {
			continue
		}
		result[pr.URL] = pr.Score
	}

	return result, cursor.Err()
}

// GetRandomPage returns a random page metadata using $sample aggregation.
func (m *MongoClient) GetRandomPage(ctx context.Context) (*schemas.Metadata, error) {
	col := m.db.Collection(metadataCollection)

	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var meta schemas.Metadata
		if err := cursor.Decode(&meta); err != nil {
			return nil, err
		}
		return &meta, nil
	}

	return nil, mongo.ErrNoDocuments
}
