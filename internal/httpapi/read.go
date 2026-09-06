package httpapi

import (
	"net/http"

	"featureflags/internal/store"
)

// List behandelt GET /flags.
func List(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

// Get behandelt GET /flags/{key}.
func Get(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
