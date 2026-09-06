package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"featureflags/internal/store"
)

func newUpdateMux(s *store.Store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /flags/{key}", Update(s))
	mux.HandleFunc("DELETE /flags/{key}", Delete(s))
	return mux
}

func seedFlag(s *store.Store, key string, enabled bool, desc string, rollout int) {
	_ = s.Create(store.Flag{Key: key, Enabled: enabled, Description: desc, RolloutPercent: rollout})
}

func TestUpdateSuccess(t *testing.T) {
	s := store.NewStore(100)
	seedFlag(s, "beta", false, "old", 50)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodPut, "/flags/beta", `{"enabled":true,"description":"new","rollout_percent":80}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got store.Flag
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	want := store.Flag{Key: "beta", Enabled: true, Description: "new", RolloutPercent: 80}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestUpdateDefaults(t *testing.T) {
	s := store.NewStore(100)
	seedFlag(s, "beta", false, "old", 50)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodPut, "/flags/beta", `{"enabled":true}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got store.Flag
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	want := store.Flag{Key: "beta", Enabled: true, Description: "", RolloutPercent: 100}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestUpdateUnknownKey(t *testing.T) {
	s := store.NewStore(100)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodPut, "/flags/missing", `{"enabled":true}`)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateMissingEnabled(t *testing.T) {
	s := store.NewStore(100)
	seedFlag(s, "beta", false, "old", 50)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodPut, "/flags/beta", `{"description":"x"}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateInvalidRollout(t *testing.T) {
	s := store.NewStore(100)
	seedFlag(s, "beta", false, "old", 50)
	mux := newUpdateMux(s)

	for _, p := range []int{-1, 101} {
		body := `{"enabled":true,"rollout_percent":` + strconv.Itoa(p) + `}`
		w := doRequest(t, mux, http.MethodPut, "/flags/beta", body)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent %d: expected 400, got %d", p, w.Code)
		}
	}
}

func TestUpdateInvalidBody(t *testing.T) {
	s := store.NewStore(100)
	seedFlag(s, "beta", false, "old", 50)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodPut, "/flags/beta", `not json`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteSuccess(t *testing.T) {
	s := store.NewStore(100)
	seedFlag(s, "beta", false, "old", 50)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodDelete, "/flags/beta", "")

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", w.Body.String())
	}
	if _, ok := s.Get("beta"); ok {
		t.Fatal("expected flag to be removed")
	}
}

func TestDeleteUnknownKey(t *testing.T) {
	s := store.NewStore(100)
	mux := newUpdateMux(s)

	w := doRequest(t, mux, http.MethodDelete, "/flags/missing", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
