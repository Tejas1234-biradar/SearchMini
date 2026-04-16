package handlers

import "net/http"

// FrontendHTML serves the debug frontend HTML page.
func FrontendHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "./static/home.html")
}
