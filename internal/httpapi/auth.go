package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// Auth erzwingt eine Bearer-Token-Authentifizierung auf allen Routen mit
// Ausnahme von GET /healthz und des Evaluate-Pfads (jeder Pfad, der auf
// "/evaluate" endet). Fehlende oder falsche Tokens werden mit 401 beantwortet.
// Bei leerem token bleiben alle geschützten Routen per Fail-Closed auf 401.
func Auth(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublic(r) {
			next.ServeHTTP(w, r)
			return
		}

		provided := bearerToken(r)
		if token == "" || provided == "" ||
			subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isPublic meldet, ob der Request ohne Token erreichbar ist: ausschließlich
// GET /healthz und der Evaluate-Pfad.
func isPublic(r *http.Request) bool {
	if r.URL.Path == "/healthz" {
		return true
	}
	return strings.HasSuffix(r.URL.Path, "/evaluate")
}

// bearerToken extrahiert das Token aus dem Authorization-Header der Form
// "Bearer <token>". Bei abweichendem Format wird ein leerer String geliefert.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, prefix))
}
