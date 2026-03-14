package handlers

import (
	"log/slog"
	"net/http"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/data"
)

// SuggestionHandler holds dependencies for suggestion-related API endpoints.
type SuggestionHandler struct {
	Redis *data.RedisClient
}

// NewSuggestionHandler creates a new SuggestionHandler.
func NewSuggestionHandler(redis *data.RedisClient) *SuggestionHandler {
	return &SuggestionHandler{Redis: redis}
}

// Suggestions handles GET /api/suggestions?q=<prefix>
func (h *SuggestionHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("q")
	if prefix == "" {
		sendJSON(w, http.StatusOK, map[string][]string{"suggestions": {}})
		return
	}

	suggestions, err := h.Redis.GetSearchSuggestions(r.Context(), prefix)
	if err != nil {
		slog.Error("failed to get suggestions", "error", err)
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch suggestions"})
		return
	}

	sendJSON(w, http.StatusOK, map[string][]string{"suggestions": suggestions})
}

// TopSearches handles GET /api/top-searches
func (h *SuggestionHandler) TopSearches(w http.ResponseWriter, r *http.Request) {
	top, err := h.Redis.GetTopSearches(r.Context(), 10)
	if err != nil {
		slog.Error("failed to get top searches", "error", err)
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch top searches"})
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{"top": top})
}
