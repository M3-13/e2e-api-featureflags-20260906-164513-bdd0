package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitAllowsUpToLimit(t *testing.T) {
	h := RateLimit(testEchoHandler(), 3)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimitRejectsBeyondLimit(t *testing.T) {
	h := RateLimit(testEchoHandler(), 2)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 beyond limit, got %d", w.Code)
	}
}

func TestRateLimitZeroDisables(t *testing.T) {
	h := RateLimit(testEchoHandler(), 0)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200 with rps=0, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimitSeparatesClients(t *testing.T) {
	h := RateLimit(testEchoHandler(), 1)

	first := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	first.RemoteAddr = "10.0.0.1:1000"
	second := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	second.RemoteAddr = "10.0.0.2:1000"

	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, first)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, second)

	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Fatalf("expected both clients to pass, got %d and %d", w1.Code, w2.Code)
	}
}
