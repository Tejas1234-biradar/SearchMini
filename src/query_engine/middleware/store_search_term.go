package middleware

import (
	"context"
	"net/http"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/data"
)

// NewSearchTracker creates a middleware that uses the provided Redis client.
// It intercepts search requests and increments the search term counter in Redis.
func NewSearchTracker(redis *data.RedisClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("q")
			if query != "" {
				// Record asynchronously to not block the response
				go func(q string) {
					// Using background context for recording to ensure it finishes even if request is canceled
					_ = redis.IncrSearchTerm(context.Background(), q)
				}(query)
			}
			next.ServeHTTP(w, r)
		})
	}
}
