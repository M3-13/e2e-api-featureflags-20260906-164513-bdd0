package httpapi

import (
	"encoding/json"
	"net/http"

	"featureflags/internal/store"
)

// updateRequest ist der Body von PUT /flags/{key}.
type updateRequest struct {
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent *int   `json:"rollout_percent"`
}

// Update behandelt PUT /flags/{key}.
func Update(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		var req updateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}

		rollout := 100
		if req.RolloutPercent != nil {
			rollout = *req.RolloutPercent
		}
		if !validRolloutPercent(rollout) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		f := store.Flag{
			Key:            key,
			Enabled:        *req.Enabled,
			Description:    req.Description,
			RolloutPercent: rollout,
		}

		if !s.Update(key, f) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		writeJSON(w, http.StatusOK, f)
	}
}

// Delete behandelt DELETE /flags/{key}.
func Delete(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		if !s.Delete(key) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
