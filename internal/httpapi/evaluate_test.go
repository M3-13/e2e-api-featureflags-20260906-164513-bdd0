package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/store"
)

func newEvaluateStore(t *testing.T) *store.Store {
	t.Helper()
	s := store.NewStore(100)
	if err := s.Create(store.Flag{Key: "enabled", Enabled: true, Description: "", RolloutPercent: 100}); err != nil {
		t.Fatalf("create enabled flag: %v", err)
	}
	if err := s.Create(store.Flag{Key: "disabled", Enabled: false, Description: "", RolloutPercent: 100}); err != nil {
		t.Fatalf("create disabled flag: %v", err)
	}
	return s
}

func evaluateRequest(key, query string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/flags/"+key+"/evaluate"+query, nil)
	req.SetPathValue("key", key)
	return req
}

func TestEvaluateUnknownKeyReturns404(t *testing.T) {
	s := newEvaluateStore(t)
	req := evaluateRequest("missing", "?user=u1")
	rec := httptest.NewRecorder()
	Evaluate(s)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestEvaluateMissingUserReturns400(t *testing.T) {
	s := newEvaluateStore(t)
	req := evaluateRequest("enabled", "")
	rec := httptest.NewRecorder()
	Evaluate(s)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestEvaluateEmptyUserReturns400(t *testing.T) {
	s := newEvaluateStore(t)
	req := evaluateRequest("enabled", "?user=")
	rec := httptest.NewRecorder()
	Evaluate(s)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestEvaluateDisabledAlwaysFalse(t *testing.T) {
	s := newEvaluateStore(t)
	for _, user := range []string{"u1", "u2", "u3", "u4", "u5"} {
		req := evaluateRequest("disabled", "?user="+user)
		rec := httptest.NewRecorder()
		Evaluate(s)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var body map[string]bool
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if body["result"] {
			t.Fatalf("expected result=false for disabled flag, got true for user %s", user)
		}
	}
}

func TestEvaluateDeterministicSameResult(t *testing.T) {
	s := newEvaluateStore(t)

	var first bool
	firstSet := false
	for i := 0; i < 20; i++ {
		req := evaluateRequest("enabled", "?user=alice")
		rec := httptest.NewRecorder()
		Evaluate(s)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var body map[string]bool
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if !firstSet {
			first = body["result"]
			firstSet = true
			continue
		}
		if body["result"] != first {
			t.Fatalf("expected deterministic result, got %v then %v", first, body["result"])
		}
	}
}

func TestEvaluateResponseHasNoUserValue(t *testing.T) {
	s := newEvaluateStore(t)
	req := evaluateRequest("enabled", "?user=secret-user-123")
	rec := httptest.NewRecorder()
	Evaluate(s)(rec, req)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := body["user"]; ok {
		t.Fatalf("response must not contain user value, got %v", body)
	}
	if len(body) != 1 {
		t.Fatalf("response must contain only result, got %v", body)
	}
	if _, ok := body["result"].(bool); !ok {
		t.Fatalf("response must contain bool result, got %v", body)
	}
}
