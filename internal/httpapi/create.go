package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"featureflags/internal/store"
)

// createRequest ist das lokale Request-Schema für POST /flags.
type createRequest struct {
	Key            string `json:"key"`
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent *int   `json:"rollout_percent"`
}

// Create behandelt POST /flags.
func Create(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}

		rolloutPercent := 100
		if req.RolloutPercent != nil {
			rolloutPercent = *req.RolloutPercent
		}

		if !validKey(req.Key) {
			writeError(w, http.StatusBadRequest, "invalid key")
			return
		}
		if !validRolloutPercent(rolloutPercent) {
			writeError(w, http.StatusBadRequest, "invalid rollout_percent")
			return
		}

		flag := store.Flag{
			Key:            req.Key,
			Enabled:        *req.Enabled,
			Description:    req.Description,
			RolloutPercent: rolloutPercent,
		}

		if err := s.Create(flag); err != nil {
			switch {
			case errors.Is(err, store.ErrDuplicate):
				writeError(w, http.StatusConflict, "flag already exists")
			case errors.Is(err, store.ErrLimit):
				writeError(w, http.StatusBadRequest, "flag limit reached")
			default:
				writeError(w, http.StatusInternalServerError, "internal error")
			}
			return
		}

		writeJSON(w, http.StatusCreated, flag)
	}
}
