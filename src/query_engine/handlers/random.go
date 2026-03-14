package handlers

import "net/http"

// Random handles GET /api/random
// TODO: fetch a random page via $sample aggregation on metadata collection
func Random(w http.ResponseWriter, r *http.Request) {}
