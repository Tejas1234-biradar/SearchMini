package routes

import (
	"net/http"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/data"
	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/handlers"
	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/middleware"
	"github.com/go-chi/chi/v5"
)

// Register mounts all API routes onto the given chi router.
func Register(r *chi.Mux, mongo *data.MongoClient, redis *data.RedisClient) {
	// Initialize handlers
	searchH := handlers.NewSearchHandler(mongo)
	suggestH := handlers.NewSuggestionHandler(redis)
	randomH := handlers.NewRandomHandler(mongo)

	// Middleware
	tracker := middleware.NewSearchTracker(redis)

	// Serve static frontend
	r.Handle("/*", http.FileServer(http.Dir("./static")))

	r.Route("/api", func(r chi.Router) {
		// Search — with search-term tracking middleware
		r.With(tracker).Get("/search", searchH.Search)

		// Stats & rankings
		r.Get("/stats", searchH.Stats)
		r.Get("/top-pages", searchH.TopPages)
		r.Get("/random", randomH.Random)

		// Page graph
		r.Get("/page-connections", searchH.PageConnections)

		// Suggestions & trending
		r.Get("/suggestions", suggestH.Suggestions)
		r.Get("/top-searches", suggestH.TopSearches)
	})
}
