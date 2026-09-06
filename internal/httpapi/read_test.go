package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func TestListEmpty(t *testing.T) {
	s := store.NewStore(10)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)

	List(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var got []store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got == nil {
		t.Fatal("list is null, want []")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestListAfterCreate(t *testing.T) {
	s := store.NewStore(10)
	if err := s.Create(store.Flag{Key: "a", Enabled: true, Description: "A", RolloutPercent: 50}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.Create(store.Flag{Key: "b", Enabled: false, Description: "B", RolloutPercent: 100}); err != nil {
		t.Fatalf("create: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	List(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	keys := map[string]bool{}
	for _, f := range got {
		keys[f.Key] = true
	}
	for _, k := range []string{"a", "b"} {
		if !keys[k] {
			t.Fatalf("missing key %q in list", k)
		}
	}
}

func TestGetExisting(t *testing.T) {
	s := store.NewStore(10)
	want := store.Flag{Key: "feature", Enabled: true, Description: "desc", RolloutPercent: 25}
	if err := s.Create(want); err != nil {
		t.Fatalf("create: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature", nil)
	req.SetPathValue("key", "feature")

	Get(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got != want {
		t.Fatalf("flag = %+v, want %+v", got, want)
	}
}

func TestGetUnknown(t *testing.T) {
	s := store.NewStore(10)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	req.SetPathValue("key", "missing")

	Get(s)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON error object: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("error object missing message: %v", body)
	}
}
