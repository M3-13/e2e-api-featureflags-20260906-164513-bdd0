package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestServerProcessSmoke startet den echten Server über einen TCP-Listener
// (httptest.NewServer) mit der vollen Middleware-Kette und weist die
// versprochenen Laufzeit-Verhalten über HTTP nach.
func TestServerProcessSmoke(t *testing.T) {
	t.Setenv("AUTH_TOKEN", "smoke-token")
	t.Setenv("RATE_LIMIT_RPS", "1000")

	srv := httptest.NewServer(newHandler())
	defer srv.Close()

	// Probe 1: GET /healthz -> 200 und Body {"status":"ok"}.
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want 200", resp.StatusCode)
	}
	var health map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("GET /healthz body decode: %v", err)
	}
	resp.Body.Close()
	if health["status"] != "ok" {
		t.Fatalf("GET /healthz status field = %q, want %q", health["status"], "ok")
	}

	// Probe 2: POST /flags ohne Authorization -> 401 mit JSON-Fehlerobjekt.
	resp, err = postFlag(srv.URL+"/flags", "", `{"key":"smoke","enabled":true}`)
	if err != nil {
		t.Fatalf("POST /flags without token: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("POST /flags without token status = %d, want 401", resp.StatusCode)
	}
	var errBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("POST /flags without token body decode: %v", err)
	}
	resp.Body.Close()
	if errBody["error"] == "" {
		t.Fatalf("POST /flags without token error field is empty: %v", errBody)
	}

	// Probe 3: POST /flags mit gültigem Bearer-Token -> 201 mit key "smoke".
	resp, err = postFlag(srv.URL+"/flags", "smoke-token", `{"key":"smoke","enabled":true}`)
	if err != nil {
		t.Fatalf("POST /flags with token: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /flags with token status = %d, want 201", resp.StatusCode)
	}
	var flag struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&flag); err != nil {
		t.Fatalf("POST /flags with token body decode: %v", err)
	}
	resp.Body.Close()
	if flag.Key != "smoke" {
		t.Fatalf("POST /flags with token key = %q, want %q", flag.Key, "smoke")
	}
}

// postFlag führt einen POST aus und liefert die Antwort zurück.
func postFlag(url, token, body string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
