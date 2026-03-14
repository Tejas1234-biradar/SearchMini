package handlers

import (
	"net/http"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/data"
)

// RandomHandler handles the random page endpoint.
type RandomHandler struct {
	Mongo *data.MongoClient
}

func NewRandomHandler(mongo *data.MongoClient) *RandomHandler {
	return &RandomHandler{Mongo: mongo}
}

// Random handles GET /api/random
func (h *RandomHandler) Random(w http.ResponseWriter, r *http.Request) {
	meta, err := h.Mongo.GetRandomPage(r.Context())
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, meta)
}
