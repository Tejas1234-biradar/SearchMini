package handlers

import "net/http"

// Suggestions handles GET /api/suggestions?q=<prefix>
// TODO: query Redis sorted set or dictionary collection for prefix matches
func Suggestions(w http.ResponseWriter, r *http.Request) {}

// TopSearches handles GET /api/top-searches
// TODO: return most searched terms from Redis ZREVRANGE
func TopSearches(w http.ResponseWriter, r *http.Request) {}
