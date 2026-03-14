package schemas

// SearchResult represents a single search hit returned by the API.
type SearchResult struct {
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	SummaryText string  `json:"summary_text"`
	LastCrawled string  `json:"last_crawled"`
	TFIDFWeight float64 `json:"tfidf_weight"`
	PageRank    float64 `json:"pagerank"`
	Score       float64 `json:"score"` // 0.6*tfidf + 0.4*pagerank
}

// Metadata mirrors the metadata MongoDB collection.
type Metadata struct {
	ID          string `bson:"_id" json:"url"`
	Title       string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
	SummaryText string `bson:"summary_text" json:"summary_text"`
	LastCrawled string `bson:"last_crawled" json:"last_crawled"`
}

// PageConnection holds outlink/backlink data for a URL.
type PageConnection struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
}

// TermScore holds a search term and its hit count from Redis.
type TermScore struct {
	Term  string  `json:"term"`
	Score float64 `json:"count"`
}

// PageRankResult holds a URL and its PageRank score.
type PageRankResult struct {
	URL   string  `bson:"_id" json:"url"`
	Score float64 `bson:"score" json:"score"`
}
