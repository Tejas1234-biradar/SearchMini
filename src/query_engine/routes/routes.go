package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/handlers"
	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/middleware"
)

// Register mounts all API routes onto the given chi router.
func Register(r *chi.Mux) {
	// Serve static frontend
	r.Handle("/*", http.FileServer(http.Dir("./static")))

	r.Route("/api", func(r chi.Router) {
		// Search — with search-term tracking middleware
		r.With(middleware.StoreSearchTerm).Get("/search", handlers.Search)

		// Stats & rankings
		r.Get("/stats", handlers.Stats)
		r.Get("/top-pages", handlers.TopPages)
		r.Get("/random", handlers.Random)

		// Page graph
		r.Get("/page-connections", handlers.PageConnections)

		// Suggestions & trending
		r.Get("/suggestions", handlers.Suggestions)
		r.Get("/top-searches", handlers.TopSearches)
	})
}
