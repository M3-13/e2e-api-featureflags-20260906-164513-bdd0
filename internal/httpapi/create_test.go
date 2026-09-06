package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func doCreate(t *testing.T, s *store.Store, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Create(s)(rec, req)
	return rec
}

func TestCreateReturnsCreatedFlag(t *testing.T) {
	s := store.NewStore(100)
	rec := doCreate(t, s, `{"key":"feature_a","enabled":true,"description":"desc","rollout_percent":50}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var flag store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flag); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if flag.Key != "feature_a" || !flag.Enabled || flag.Description != "desc" || flag.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", flag)
	}
}

func TestCreateDefaultsRolloutPercentTo100(t *testing.T) {
	s := store.NewStore(100)
	rec := doCreate(t, s, `{"key":"feature_b","enabled":false}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var flag store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flag); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if flag.RolloutPercent != 100 {
		t.Fatalf("expected rollout_percent 100, got %d", flag.RolloutPercent)
	}
}

func TestCreateDuplicateReturns409(t *testing.T) {
	s := store.NewStore(100)
	if rec := doCreate(t, s, `{"key":"dup","enabled":true}`); rec.Code != http.StatusCreated {
		t.Fatalf("first create failed: %d", rec.Code)
	}

	rec := doCreate(t, s, `{"key":"dup","enabled":true}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var errBody map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] == "" {
		t.Fatalf("expected error object, got %v", errBody)
	}
}

func TestCreateValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing key", `{"enabled":true}`},
		{"empty key", `{"key":"","enabled":true}`},
		{"invalid key character", `{"key":"bad key!","enabled":true}`},
		{"rollout too low", `{"key":"k","enabled":true,"rollout_percent":-1}`},
		{"rollout too high", `{"key":"k","enabled":true,"rollout_percent":101}`},
		{"invalid json", `{not json}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.NewStore(100)
			rec := doCreate(t, s, tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			var errBody map[string]string
			_ = json.NewDecoder(rec.Body).Decode(&errBody)
			if errBody["error"] == "" {
				t.Fatalf("expected error object, got %v", errBody)
			}
		})
	}
}

func TestCreateDefaultsEnabledToFalse(t *testing.T) {
	s := store.NewStore(100)
	rec := doCreate(t, s, `{"key":"default_enabled"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var flag store.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flag); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if flag.Enabled {
		t.Fatalf("expected enabled=false by default, got %+v", flag)
	}
	if flag.Key != "default_enabled" {
		t.Fatalf("unexpected key: %q", flag.Key)
	}
}

func TestCreateLimitReturns400(t *testing.T) {
	s := store.NewStore(1)
	if rec := doCreate(t, s, `{"key":"only","enabled":true}`); rec.Code != http.StatusCreated {
		t.Fatalf("first create failed: %d", rec.Code)
	}

	rec := doCreate(t, s, `{"key":"second","enabled":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for full store, got %d", rec.Code)
	}
	var errBody map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] == "" {
		t.Fatalf("expected error object, got %v", errBody)
	}
}
