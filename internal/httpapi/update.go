package httpapi

import (
	"net/http"

	"featureflags/internal/store"
)

// Update behandelt PUT /flags/{key}.
func Update(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

// Delete behandelt DELETE /flags/{key}.
func Delete(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
