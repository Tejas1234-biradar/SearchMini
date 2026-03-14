package middleware

import "net/http"

// StoreSearchTerm increments the search term counter in Redis after a search response.
// TODO: wrap handler, extract ?q= param, call RedisClient.IncrSearchTerm
func StoreSearchTerm(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
