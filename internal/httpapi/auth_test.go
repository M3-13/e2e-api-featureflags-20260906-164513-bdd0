package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRejectsMissingToken(t *testing.T) {
	h := Auth(testEchoHandler(), "secret")

	rec := doHTTPCall(t, h, http.MethodPost, "/flags", `{}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rec.Code)
	}
}

func TestAuthRejectsWrongToken(t *testing.T) {
	h := Auth(testEchoHandler(), "secret")

	r := httptest.NewRequest(http.MethodPost, "/flags", nil)
	r.Header.Set("Authorization", "Bearer wrong")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", w.Code)
	}
}

func TestAuthAllowsHealthz(t *testing.T) {
	h := Auth(testEchoHandler(), "secret")

	rec := doHTTPCall(t, h, http.MethodGet, "/healthz", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /healthz to pass without token, got %d", rec.Code)
	}
}

func TestAuthAllowsEvaluatePath(t *testing.T) {
	h := Auth(testEchoHandler(), "secret")

	rec := doHTTPCall(t, h, http.MethodGet, "/flags/some/evaluate", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /evaluate path to pass without token, got %d", rec.Code)
	}
}

func TestAuthProtectsPost(t *testing.T) {
	h := Auth(testEchoHandler(), "secret")

	rec := doHTTPCall(t, h, http.MethodPost, "/flags", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected POST /flags without token to be 401, got %d", rec.Code)
	}
}

func TestAuthAllowsValidToken(t *testing.T) {
	h := Auth(testEchoHandler(), "secret")

	r := httptest.NewRequest(http.MethodPost, "/flags", nil)
	r.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid token, got %d", w.Code)
	}
}

func TestAuthFailClosedOnEmptyToken(t *testing.T) {
	h := Auth(testEchoHandler(), "")

	rec := doHTTPCall(t, h, http.MethodPost, "/flags", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with empty configured token, got %d", rec.Code)
	}
}

func testEchoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
