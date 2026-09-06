package httpapi

import (
	"net/http"

	"featureflags/internal/store"
)

// Create behandelt POST /flags.
func Create(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
