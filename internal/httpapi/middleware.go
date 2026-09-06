package httpapi

import (
	"log"
	"net/http"
	"time"
)

// statusWriter wraps an http.ResponseWriter to capture the status code of a
// response and to convert the 405 answers produced by http.ServeMux into the
// JSON error object {"error":"method not allowed"}.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if sw.wroteHeader {
		return
	}
	sw.status = code
	sw.wroteHeader = true
	if code == http.StatusMethodNotAllowed {
		sw.Header().Set("Content-Type", "application/json")
		sw.Header().Set("Cache-Control", "no-store")
		sw.Header().Set("X-Content-Type-Options", "nosniff")
	}
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
		sw.WriteHeader(http.StatusOK)
	}
	if sw.status == http.StatusMethodNotAllowed {
		return sw.ResponseWriter.Write([]byte(`{"error":"method not allowed"}`))
	}
	return sw.ResponseWriter.Write(b)
}

// Logging protokolliert jeden Request mit Methode, Pfad (ohne Query-String),
// Statuscode und Dauer. Es werden ausschließlich diese vier Werte geloggt,
// niemals benutzerbezogene Daten wie der user-Parameter.
func Logging(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(sw, r)
		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		logger.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

// Recover fängt Panics in nachfolgenden Handlern ab und antwortet mit einem
// generischen 500-JSON-Fehlerobjekt, ohne den Serverprozess zu beenden. Der
// Fehler wird nur intern geloggt, ohne Stacktrace oder Details an den Client.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("recovered from panic (type %T)", rec)
				writeError(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LimitBody begrenzt die Größe des Request-Bodys auf 1 MiB, bevor dieser
// eingelesen wird. Wird die Grenze beim Einlesen überschritten, antwortet der
// Handler mit 400 über writeError.
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}
