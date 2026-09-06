package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"featureflags/internal/httpapi"
	"featureflags/internal/store"
)

const maxFlags = 10000

const (
	defaultAddr    = "127.0.0.1:8080"
	defaultRateRPS = 50
)

// getenv liest eine Umgebungsvariable und liefert andernfalls den Default.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// newHandler baut die vollständige Handler-Kette inklusive Routing und
// Middleware. Als eigene Funktion ausgelagert, damit Tests sie direkt
// aufrufen können.
func newHandler() http.Handler {
	s := store.NewStore(maxFlags)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /flags", httpapi.Create(s))
	mux.HandleFunc("GET /flags", httpapi.List(s))
	mux.HandleFunc("GET /flags/{key}", httpapi.Get(s))
	mux.HandleFunc("PUT /flags/{key}", httpapi.Update(s))
	mux.HandleFunc("DELETE /flags/{key}", httpapi.Delete(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", httpapi.Evaluate(s))

	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Printf("warning: AUTH_TOKEN is not set; protected routes will answer 401")
	}
	rateRPS := defaultRateRPS
	if v := os.Getenv("RATE_LIMIT_RPS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateRPS = n
		} else {
			log.Printf("warning: invalid RATE_LIMIT_RPS %q; using default %d", v, defaultRateRPS)
		}
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	return httpapi.Logging(logger, httpapi.Recover(httpapi.LimitBody(httpapi.RateLimit(httpapi.Auth(mux, authToken), rateRPS))))
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	srv := &http.Server{
		Addr:              getenv("ADDR", defaultAddr),
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}
