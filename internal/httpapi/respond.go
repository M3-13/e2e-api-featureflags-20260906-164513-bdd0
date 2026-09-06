package httpapi

import (
	"encoding/json"
	"net/http"
)

// writeJSON schreibt v als JSON-Antwort mit Status status und setzt den
// Content-Type auf application/json.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError schreibt ein JSON-Fehlerobjekt {"error": msg} mit Status status.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
