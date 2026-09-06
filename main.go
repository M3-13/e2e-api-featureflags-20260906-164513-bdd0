package main

import (
	"log"
	"net/http"
	"os"

	"featureflags/internal/httpapi"
	"featureflags/internal/store"
)

const maxFlags = 10000

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

	logger := log.New(os.Stdout, "", log.LstdFlags)
	return httpapi.Recover(httpapi.Logging(logger, httpapi.LimitBody(mux)))
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
	if err := http.ListenAndServe(":8080", newHandler()); err != nil {
		logger.Fatal(err)
	}
}
