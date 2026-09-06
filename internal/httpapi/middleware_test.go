package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	h := Logging(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags/abc?user=secret", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	line := strings.TrimSpace(buf.String())
	fields := strings.Fields(line)
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields (method, path, status, duration), got %d in %q", len(fields), line)
	}
	if fields[0] != http.MethodPost {
		t.Fatalf("expected method POST, got %q", fields[0])
	}
	if fields[1] != "/flags/abc" {
		t.Fatalf("expected path /flags/abc without query, got %q", fields[1])
	}
	if fields[2] != "201" {
		t.Fatalf("expected status 201, got %q", fields[2])
	}
	if fields[3] == "" {
		t.Fatalf("expected a duration, got empty")
	}
	if strings.Contains(line, "secret") || strings.Contains(line, "?") {
		t.Fatalf("log line leaked query string: %q", line)
	}
}

func TestRecover(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			panic("boom")
		}
		w.WriteHeader(http.StatusOK)
	})
	h := Recover(next)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/flags", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["error"] != "internal error" {
		t.Fatalf("expected generic error %q, got %q", "internal error", body["error"])
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("response leaked panic details: %q", rec.Body.String())
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/flags", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("process did not continue after panic: expected 200, got %d", rec2.Code)
	}
}

func TestLimitBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				writeError(w, http.StatusBadRequest, "request body too large")
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	h := LimitBody(next)

	body := bytes.Repeat([]byte("a"), (1<<20)+1)
	req := httptest.NewRequest(http.MethodPost, "/flags", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for oversized body, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if resp["error"] == "" {
		t.Fatalf("expected error field in response")
	}
}

func TestServeMux405AsJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /flags/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	h := Logging(logger, mux)

	req := httptest.NewRequest(http.MethodPost, "/flags/abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["error"] != "method not allowed" {
		t.Fatalf("expected error %q, got %q", "method not allowed", body["error"])
	}
}
