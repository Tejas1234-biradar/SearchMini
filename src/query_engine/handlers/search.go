package handlers

import "net/http"

// Search handles GET /api/search?q=<query>&page=<n>
// TODO: tokenize query, call MongoClient.SearchPages, enrich with metadata, return JSON
func Search(w http.ResponseWriter, r *http.Request) {}

// Stats handles GET /api/stats
// TODO: return { status: "up", pages: <count> }
func Stats(w http.ResponseWriter, r *http.Request) {}

// PageConnections handles GET /api/page-connections?url=<url>
// TODO: return outlinks and backlinks with titles
func PageConnections(w http.ResponseWriter, r *http.Request) {}

// TopPages handles GET /api/top-pages
// TODO: return top ranked pages by cumulative TF-IDF weight
func TopPages(w http.ResponseWriter, r *http.Request) {}
