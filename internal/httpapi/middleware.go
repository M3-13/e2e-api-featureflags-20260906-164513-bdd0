package httpapi

import (
	"log"
	"net/http"
)

// Logging protokolliert jeden Request (Implementierung folgt im
// Middleware-Ticket; das Skelett reicht den Request nur weiter).
func Logging(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Recover fängt Panics in nachfolgenden Handlern ab
// (Implementierung folgt im Middleware-Ticket).
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// LimitBody begrenzt die Größe des Request-Bodys
// (Implementierung folgt im Middleware-Ticket).
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
