package httpapi

import (
	"net/http"

	"featureflags/internal/store"
)

// Evaluate behandelt GET /flags/{key}/evaluate?user={id}.
func Evaluate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		user := r.URL.Query().Get("user")
		if user == "" {
			writeError(w, http.StatusBadRequest, "user is required")
			return
		}

		flag, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		result := false
		if flag.Enabled {
			result = int(stableHash(key, user)%100) < flag.RolloutPercent
		}

		writeJSON(w, http.StatusOK, map[string]bool{"result": result})
	}
}
