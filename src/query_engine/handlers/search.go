package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/data"
)

// SearchHandler holds dependencies for search-related API endpoints.
type SearchHandler struct {
	Mongo *data.MongoClient
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(mongo *data.MongoClient) *SearchHandler {
	return &SearchHandler{Mongo: mongo}
}

// Stats handles GET /api/stats
func (h *SearchHandler) Stats(w http.ResponseWriter, r *http.Request) {
	count, err := h.Mongo.GetStats(r.Context())
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"status": "up",
		"pages":  count,
	})
}

// Search handles GET /api/search?q=<query>&page=<n>
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		sendJSON(w, http.StatusOK, map[string]interface{}{
			"total":   0,
			"page":    1,
			"results": []interface{}{},
		})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	// Normalize and tokenize
	words := strings.Fields(strings.ToLower(query))

	results, total, err := h.Mongo.SearchPages(r.Context(), words, page, 20)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"total":   total,
		"page":    page,
		"results": results,
	})
}

// PageConnections handles GET /api/page-connections?url=<url>
func (h *SearchHandler) PageConnections(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "url parameter is required"})
		return
	}

	outlinks, backlinks, err := h.Mongo.GetPageConnections(r.Context(), targetURL)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"url":       targetURL,
		"outlinks":  outlinks,
		"backlinks": backlinks,
	})
}

// TopPages handles GET /api/top-pages
func (h *SearchHandler) TopPages(w http.ResponseWriter, r *http.Request) {
	results, err := h.Mongo.GetTopRankedPages(r.Context(), 10)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, results)
}
