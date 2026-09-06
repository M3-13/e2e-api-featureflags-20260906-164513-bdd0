package httpapi

import (
	"net/http"

	"featureflags/internal/store"
)

// Evaluate behandelt GET /flags/{key}/evaluate.
func Evaluate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
